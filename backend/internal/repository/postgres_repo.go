package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go-nmos/backend/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	err := r.pool.QueryRow(ctx,
		`SELECT username, password_hash, role, created_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT username, role, created_at FROM users ORDER BY username ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.Username, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user models.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users(username, password_hash, role) VALUES ($1, $2, $3)`,
		user.Username, user.PasswordHash, user.Role,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("user already exists")
	}
	return err
}

func (r *PostgresRepository) ListFlows(ctx context.Context, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error) {
	sortBy, sortOrder = normalizeFlowSort(sortBy, sortOrder)
	query := fmt.Sprintf(`
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol
		FROM flows
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, sortBy, sortOrder)
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []models.Flow
	for rows.Next() {
		var f models.Flow
		if err := rows.Scan(
			&f.ID, &f.FlowID, &f.DisplayName, &f.MulticastIP, &f.SourceIP, &f.Port,
			&f.FlowStatus, &f.Availability, &f.Locked, &f.Note, &f.UpdatedAt, &f.LastSeen, &f.TransportProto,
		); err != nil {
			return nil, err
		}
		flows = append(flows, f)
	}
	return flows, rows.Err()
}

func (r *PostgresRepository) SearchFlows(ctx context.Context, query string, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error) {
	sortBy, sortOrder = normalizeFlowSort(sortBy, sortOrder)
	sql := fmt.Sprintf(`
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol
		FROM flows
		WHERE
			display_name ILIKE '%%' || $1 || '%%' OR
			flow_id::text ILIKE '%%' || $1 || '%%' OR
			multicast_ip ILIKE '%%' || $1 || '%%' OR
			source_ip ILIKE '%%' || $1 || '%%' OR
			note ILIKE '%%' || $1 || '%%'
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortBy, sortOrder)
	rows, err := r.pool.Query(ctx, sql, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flows []models.Flow
	for rows.Next() {
		var f models.Flow
		if err := rows.Scan(
			&f.ID, &f.FlowID, &f.DisplayName, &f.MulticastIP, &f.SourceIP, &f.Port,
			&f.FlowStatus, &f.Availability, &f.Locked, &f.Note, &f.UpdatedAt, &f.LastSeen, &f.TransportProto,
		); err != nil {
			return nil, err
		}
		flows = append(flows, f)
	}
	return flows, rows.Err()
}

func (r *PostgresRepository) GetFlowSummary(ctx context.Context) (models.FlowSummary, error) {
	var s models.FlowSummary
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*)::int AS total,
			COUNT(*) FILTER (WHERE flow_status = 'active')::int AS active,
			COUNT(*) FILTER (WHERE locked = true)::int AS locked,
			COUNT(*) FILTER (WHERE flow_status = 'unused')::int AS unused,
			COUNT(*) FILTER (WHERE flow_status = 'maintenance')::int AS maintenance
		FROM flows
	`).Scan(&s.Total, &s.Active, &s.Locked, &s.Unused, &s.Maintenance)
	return s, err
}

func (r *PostgresRepository) CountFlows(ctx context.Context) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM flows`).Scan(&total)
	return total, err
}

func (r *PostgresRepository) CountSearchFlows(ctx context.Context, query string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM flows
		WHERE
			display_name ILIKE '%' || $1 || '%' OR
			flow_id::text ILIKE '%' || $1 || '%' OR
			multicast_ip ILIKE '%' || $1 || '%' OR
			source_ip ILIKE '%' || $1 || '%' OR
			note ILIKE '%' || $1 || '%'
	`, query).Scan(&total)
	return total, err
}

func (r *PostgresRepository) ExportFlows(ctx context.Context) ([]models.Flow, error) {
	return r.ListFlows(ctx, 10000, 0, "updated_at", "desc")
}

func (r *PostgresRepository) ImportFlows(ctx context.Context, flows []models.Flow) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	imported := 0
	for _, flow := range flows {
		if strings.TrimSpace(flow.DisplayName) == "" {
			continue
		}
		if strings.TrimSpace(flow.FlowID) == "" {
			continue
		}
		if flow.TransportProto == "" {
			flow.TransportProto = "RTP/UDP"
		}
		if flow.FlowStatus == "" {
			flow.FlowStatus = "active"
		}
		if flow.Availability == "" {
			flow.Availability = "available"
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO flows(flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, transport_protocol)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (flow_id) DO UPDATE SET
				display_name = EXCLUDED.display_name,
				multicast_ip = EXCLUDED.multicast_ip,
				source_ip = EXCLUDED.source_ip,
				port = EXCLUDED.port,
				flow_status = EXCLUDED.flow_status,
				availability = EXCLUDED.availability,
				locked = EXCLUDED.locked,
				note = EXCLUDED.note,
				transport_protocol = EXCLUDED.transport_protocol,
				updated_at = NOW()
		`, flow.FlowID, flow.DisplayName, flow.MulticastIP, flow.SourceIP, flow.Port, flow.FlowStatus, flow.Availability, flow.Locked, flow.Note, flow.TransportProto)
		if err != nil {
			return imported, err
		}
		imported++
	}

	if err := tx.Commit(ctx); err != nil {
		return imported, err
	}
	return imported, nil
}

func (r *PostgresRepository) DetectCollisions(ctx context.Context) ([]models.CollisionGroup, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT multicast_ip, port, COUNT(*)::int AS cnt, ARRAY_AGG(display_name ORDER BY display_name) AS names
		FROM flows
		WHERE multicast_ip <> '' AND port > 0
		GROUP BY multicast_ip, port
		HAVING COUNT(*) > 1
		ORDER BY cnt DESC, multicast_ip ASC, port ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collisions []models.CollisionGroup
	for rows.Next() {
		var c models.CollisionGroup
		if err := rows.Scan(&c.MulticastIP, &c.Port, &c.Count, &c.FlowNames); err != nil {
			return nil, err
		}
		collisions = append(collisions, c)
	}
	return collisions, rows.Err()
}

func (r *PostgresRepository) SaveCheckerResult(ctx context.Context, kind string, result []byte) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO checker_results(kind, result)
		VALUES ($1, $2::jsonb)
	`, kind, string(result))
	return err
}

func (r *PostgresRepository) GetLatestCheckerResult(ctx context.Context, kind string) (*models.CheckerResult, error) {
	var cr models.CheckerResult
	err := r.pool.QueryRow(ctx, `
		SELECT kind, result::text, created_at
		FROM checker_results
		WHERE kind = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, kind).Scan(&cr.Kind, &cr.Result, &cr.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (r *PostgresRepository) ListAutomationJobs(ctx context.Context) ([]models.AutomationJob, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT job_id, job_type, enabled, schedule_type, schedule_value, last_run_at, last_run_status, COALESCE(last_run_result::text, '{}'), updated_at
		FROM scheduled_jobs
		ORDER BY job_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []models.AutomationJob
	for rows.Next() {
		var job models.AutomationJob
		if err := rows.Scan(
			&job.JobID,
			&job.JobType,
			&job.Enabled,
			&job.ScheduleType,
			&job.ScheduleValue,
			&job.LastRunAt,
			&job.LastRunStatus,
			&job.LastRunResult,
			&job.UpdatedAt,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *PostgresRepository) GetAutomationJob(ctx context.Context, jobID string) (*models.AutomationJob, error) {
	var job models.AutomationJob
	err := r.pool.QueryRow(ctx, `
		SELECT job_id, job_type, enabled, schedule_type, schedule_value, last_run_at, last_run_status, COALESCE(last_run_result::text, '{}'), updated_at
		FROM scheduled_jobs
		WHERE job_id = $1
	`, jobID).Scan(
		&job.JobID,
		&job.JobType,
		&job.Enabled,
		&job.ScheduleType,
		&job.ScheduleValue,
		&job.LastRunAt,
		&job.LastRunStatus,
		&job.LastRunResult,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *PostgresRepository) UpsertAutomationJob(ctx context.Context, job models.AutomationJob) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO scheduled_jobs(job_id, job_type, enabled, schedule_type, schedule_value, last_run_status, last_run_result)
		VALUES ($1, $2, $3, $4, $5, $6, COALESCE($7::jsonb, '{}'::jsonb))
		ON CONFLICT (job_id) DO UPDATE SET
			job_type = EXCLUDED.job_type,
			enabled = EXCLUDED.enabled,
			schedule_type = EXCLUDED.schedule_type,
			schedule_value = EXCLUDED.schedule_value,
			last_run_status = EXCLUDED.last_run_status,
			last_run_result = EXCLUDED.last_run_result,
			updated_at = NOW()
	`, job.JobID, job.JobType, job.Enabled, job.ScheduleType, job.ScheduleValue, job.LastRunStatus, string(job.LastRunResult))
	return err
}

func (r *PostgresRepository) SetAutomationJobEnabled(ctx context.Context, jobID string, enabled bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE scheduled_jobs
		SET enabled = $2, updated_at = NOW()
		WHERE job_id = $1
	`, jobID, enabled)
	return err
}

func (r *PostgresRepository) UpdateAutomationJobRun(ctx context.Context, jobID, status string, result []byte) error {
	if len(result) == 0 {
		result = []byte(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE scheduled_jobs
		SET
			last_run_at = NOW(),
			last_run_status = $2,
			last_run_result = $3::jsonb,
			updated_at = NOW()
		WHERE job_id = $1
	`, jobID, status, string(result))
	return err
}

func (r *PostgresRepository) ListRootBuckets(ctx context.Context) ([]models.AddressBucket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, parent_id, bucket_type, name, cidr, start_ip, end_ip, color, description, COALESCE(metadata::text, '{}'), created_at, updated_at
		FROM address_buckets
		WHERE parent_id IS NULL
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AddressBucket
	for rows.Next() {
		var b models.AddressBucket
		if err := rows.Scan(&b.ID, &b.ParentID, &b.BucketType, &b.Name, &b.CIDR, &b.StartIP, &b.EndIP, &b.Color, &b.Description, &b.Metadata, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListChildBuckets(ctx context.Context, parentID int64) ([]models.AddressBucket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, parent_id, bucket_type, name, cidr, start_ip, end_ip, color, description, COALESCE(metadata::text, '{}'), created_at, updated_at
		FROM address_buckets
		WHERE parent_id = $1
		ORDER BY id ASC
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AddressBucket
	for rows.Next() {
		var b models.AddressBucket
		if err := rows.Scan(&b.ID, &b.ParentID, &b.BucketType, &b.Name, &b.CIDR, &b.StartIP, &b.EndIP, &b.Color, &b.Description, &b.Metadata, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateBucket(ctx context.Context, bucket models.AddressBucket) (int64, error) {
	var id int64
	if len(bucket.Metadata) == 0 {
		bucket.Metadata = json.RawMessage(`{}`)
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO address_buckets(parent_id, bucket_type, name, cidr, start_ip, end_ip, color, description, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb)
		RETURNING id
	`, bucket.ParentID, bucket.BucketType, bucket.Name, bucket.CIDR, bucket.StartIP, bucket.EndIP, bucket.Color, bucket.Description, string(bucket.Metadata)).Scan(&id)
	return id, err
}

func (r *PostgresRepository) UpdateBucket(ctx context.Context, id int64, updates map[string]any) error {
	allowed := map[string]bool{
		"name":        true,
		"cidr":        true,
		"start_ip":    true,
		"end_ip":      true,
		"color":       true,
		"description": true,
		"metadata":    true,
	}
	setClauses := []string{}
	args := []any{}
	i := 1
	for key, value := range updates {
		if !allowed[key] {
			continue
		}
		if key == "metadata" {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d::jsonb", key, i))
			switch vv := value.(type) {
			case string:
				args = append(args, vv)
			default:
				data, _ := json.Marshal(value)
				args = append(args, string(data))
			}
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i))
			args = append(args, value)
		}
		i++
	}
	if len(setClauses) == 0 {
		return nil
	}
	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE address_buckets SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i)
	args = append(args, id)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *PostgresRepository) DeleteBucket(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM address_buckets WHERE id = $1 OR parent_id = $1`, id)
	return err
}

func (r *PostgresRepository) ExportBuckets(ctx context.Context) ([]models.AddressBucket, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, parent_id, bucket_type, name, cidr, start_ip, end_ip, color, description, COALESCE(metadata::text, '{}'), created_at, updated_at
		FROM address_buckets
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.AddressBucket
	for rows.Next() {
		var b models.AddressBucket
		if err := rows.Scan(&b.ID, &b.ParentID, &b.BucketType, &b.Name, &b.CIDR, &b.StartIP, &b.EndIP, &b.Color, &b.Description, &b.Metadata, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ImportBuckets(ctx context.Context, buckets []models.AddressBucket) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	imported := 0
	for _, b := range buckets {
		if strings.TrimSpace(b.BucketType) == "" || strings.TrimSpace(b.Name) == "" {
			continue
		}
		if len(b.Metadata) == 0 {
			b.Metadata = json.RawMessage(`{}`)
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO address_buckets(id, parent_id, bucket_type, name, cidr, start_ip, end_ip, color, description, metadata)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb)
			ON CONFLICT (id) DO UPDATE SET
				parent_id = EXCLUDED.parent_id,
				bucket_type = EXCLUDED.bucket_type,
				name = EXCLUDED.name,
				cidr = EXCLUDED.cidr,
				start_ip = EXCLUDED.start_ip,
				end_ip = EXCLUDED.end_ip,
				color = EXCLUDED.color,
				description = EXCLUDED.description,
				metadata = EXCLUDED.metadata,
				updated_at = NOW()
		`, b.ID, b.ParentID, b.BucketType, b.Name, b.CIDR, b.StartIP, b.EndIP, b.Color, b.Description, string(b.Metadata))
		if err != nil {
			return imported, err
		}
		imported++
	}
	if err := tx.Commit(ctx); err != nil {
		return imported, err
	}
	return imported, nil
}

func (r *PostgresRepository) CreateFlow(ctx context.Context, flow models.Flow) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO flows(flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, transport_protocol)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id
	`, flow.FlowID, flow.DisplayName, flow.MulticastIP, flow.SourceIP, flow.Port, flow.FlowStatus, flow.Availability, flow.Locked, flow.Note, flow.TransportProto).Scan(&id)
	return id, err
}

func (r *PostgresRepository) GetFlowByID(ctx context.Context, id int64) (*models.Flow, error) {
	var f models.Flow
	err := r.pool.QueryRow(ctx, `
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol
		FROM flows WHERE id = $1
	`, id).Scan(
		&f.ID, &f.FlowID, &f.DisplayName, &f.MulticastIP, &f.SourceIP, &f.Port,
		&f.FlowStatus, &f.Availability, &f.Locked, &f.Note, &f.UpdatedAt, &f.LastSeen, &f.TransportProto,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *PostgresRepository) PatchFlow(ctx context.Context, id int64, updates map[string]any) error {
	allowed := map[string]bool{
		"display_name":       true,
		"multicast_ip":       true,
		"source_ip":          true,
		"port":               true,
		"flow_status":        true,
		"availability":       true,
		"locked":             true,
		"note":               true,
		"transport_protocol": true,
	}

	setClauses := []string{}
	args := []any{}
	i := 1
	for key, value := range updates {
		if !allowed[key] {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i))
		args = append(args, value)
		i++
	}
	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE flows SET %s WHERE id = $%d", strings.Join(setClauses, ", "), i)
	args = append(args, id)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *PostgresRepository) DeleteFlow(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM flows WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&value)
	return value, err
}

func (r *PostgresRepository) SetSetting(ctx context.Context, key, value string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO settings(key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, key, value)
	return err
}

func (r *PostgresRepository) HealthCheck(ctx context.Context) error {
	var one int
	return r.pool.QueryRow(ctx, `SELECT 1`).Scan(&one)
}

func normalizeFlowSort(sortBy, sortOrder string) (string, string) {
	allowedSortBy := map[string]bool{
		"updated_at":   true,
		"created_at":   true,
		"display_name": true,
		"flow_status":  true,
		"multicast_ip": true,
		"source_ip":    true,
		"port":         true,
		"availability": true,
	}
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	if !allowedSortBy[sortBy] {
		sortBy = "updated_at"
	}
	sortOrder = strings.ToUpper(strings.TrimSpace(sortOrder))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	return sortBy, sortOrder
}
