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

// FlowListFilters are optional filters for GET /flows (mmam-style subset).
type FlowListFilters struct {
	Q            string
	FlowStatus   string
	Availability string
	DataSource   string
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

func (r *PostgresRepository) UpdateUser(ctx context.Context, username string, updates map[string]any) error {
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	setParts := []string{}
	args := []any{}
	argPos := 1

	if password, ok := updates["password"].(string); ok && password != "" {
		setParts = append(setParts, fmt.Sprintf("password_hash = $%d", argPos))
		args = append(args, password)
		argPos++
	}

	if role, ok := updates["role"].(string); ok && role != "" {
		if role != "admin" && role != "editor" && role != "viewer" {
			return fmt.Errorf("invalid role: must be admin, editor, or viewer")
		}
		setParts = append(setParts, fmt.Sprintf("role = $%d", argPos))
		args = append(args, role)
		argPos++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no valid updates provided")
	}

	args = append(args, username)
	query := fmt.Sprintf(`UPDATE users SET %s WHERE username = $%d`, strings.Join(setParts, ", "), argPos)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, username string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM users WHERE username = $1`, username)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *PostgresRepository) ListFlows(ctx context.Context, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error) {
	sortBy, sortOrder = normalizeFlowSort(sortBy, sortOrder)
	query := fmt.Sprintf(`
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol,
			source_addr_a, source_port_a, multicast_addr_a, group_port_a,
			source_addr_b, source_port_b, multicast_addr_b, group_port_b,
			COALESCE(nmos_node_id::text, '') AS nmos_node_id,
			COALESCE(nmos_flow_id::text, '') AS nmos_flow_id,
			COALESCE(nmos_sender_id::text, '') AS nmos_sender_id,
			COALESCE(nmos_device_id::text, '') AS nmos_device_id,
			nmos_node_label, nmos_node_description,
			nmos_is04_host, nmos_is04_port, nmos_is05_host, nmos_is05_port,
			nmos_is04_base_url, nmos_is05_base_url, nmos_is04_version, nmos_is05_version,
			nmos_label, nmos_description, management_url,
			media_type, st2110_format, redundancy_group,
			data_source, rds_address, rds_api_url, rds_version,
			sdp_url, sdp_cache,
			alias_1, alias_2, alias_3, alias_4, alias_5, alias_6, alias_7, alias_8,
			user_field_1, user_field_2, user_field_3, user_field_4, user_field_5, user_field_6, user_field_7, user_field_8
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
			&f.SourceAddrA, &f.SourcePortA, &f.MulticastAddrA, &f.GroupPortA,
			&f.SourceAddrB, &f.SourcePortB, &f.MulticastAddrB, &f.GroupPortB,
			&f.NMOSNodeID, &f.NMOSFlowID, &f.NMOSSenderID, &f.NMOSDeviceID,
			&f.NMOSNodeLabel, &f.NMOSNodeDescription,
			&f.NMOSIS04Host, &f.NMOSIS04Port, &f.NMOSIS05Host, &f.NMOSIS05Port,
			&f.NMOSIS04BaseURL, &f.NMOSIS05BaseURL, &f.NMOSIS04Version, &f.NMOSIS05Version,
			&f.NMOSLabel, &f.NMOSDescription, &f.ManagementURL,
			&f.MediaType, &f.ST2110Format, &f.RedundancyGroup,
			&f.DataSource, &f.RDSAddress, &f.RDSAPIURL, &f.RDSVersion,
			&f.SDPURL, &f.SDPCache,
			&f.Alias1, &f.Alias2, &f.Alias3, &f.Alias4, &f.Alias5, &f.Alias6, &f.Alias7, &f.Alias8,
			&f.UserField1, &f.UserField2, &f.UserField3, &f.UserField4, &f.UserField5, &f.UserField6, &f.UserField7, &f.UserField8,
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
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol,
			source_addr_a, source_port_a, multicast_addr_a, group_port_a,
			source_addr_b, source_port_b, multicast_addr_b, group_port_b,
			COALESCE(nmos_node_id::text, '') AS nmos_node_id,
			COALESCE(nmos_flow_id::text, '') AS nmos_flow_id,
			COALESCE(nmos_sender_id::text, '') AS nmos_sender_id,
			COALESCE(nmos_device_id::text, '') AS nmos_device_id,
			nmos_node_label, nmos_node_description,
			nmos_is04_host, nmos_is04_port, nmos_is05_host, nmos_is05_port,
			nmos_is04_base_url, nmos_is05_base_url, nmos_is04_version, nmos_is05_version,
			nmos_label, nmos_description, management_url,
			media_type, st2110_format, redundancy_group,
			data_source, rds_address, rds_api_url, rds_version,
			sdp_url, sdp_cache,
			alias_1, alias_2, alias_3, alias_4, alias_5, alias_6, alias_7, alias_8,
			user_field_1, user_field_2, user_field_3, user_field_4, user_field_5, user_field_6, user_field_7, user_field_8
		FROM flows
		WHERE
			display_name ILIKE '%%' || $1 || '%%' OR
			sdp_url ILIKE '%%' || $1 || '%%' OR sdp_cache ILIKE '%%' || $1 || '%%' OR
			flow_id::text ILIKE '%%' || $1 || '%%' OR
			multicast_ip ILIKE '%%' || $1 || '%%' OR
			source_ip ILIKE '%%' || $1 || '%%' OR
			source_addr_a ILIKE '%%' || $1 || '%%' OR multicast_addr_a ILIKE '%%' || $1 || '%%' OR
			source_addr_b ILIKE '%%' || $1 || '%%' OR multicast_addr_b ILIKE '%%' || $1 || '%%' OR
			COALESCE(nmos_node_id::text, '') ILIKE '%%' || $1 || '%%' OR
			COALESCE(nmos_flow_id::text, '') ILIKE '%%' || $1 || '%%' OR
			COALESCE(nmos_sender_id::text, '') ILIKE '%%' || $1 || '%%' OR
			COALESCE(nmos_device_id::text, '') ILIKE '%%' || $1 || '%%' OR
			nmos_node_label ILIKE '%%' || $1 || '%%' OR nmos_node_description ILIKE '%%' || $1 || '%%' OR
			note ILIKE '%%' || $1 || '%%' OR
			alias_1 ILIKE '%%' || $1 || '%%' OR alias_2 ILIKE '%%' || $1 || '%%' OR alias_3 ILIKE '%%' || $1 || '%%' OR alias_4 ILIKE '%%' || $1 || '%%' OR
			alias_5 ILIKE '%%' || $1 || '%%' OR alias_6 ILIKE '%%' || $1 || '%%' OR alias_7 ILIKE '%%' || $1 || '%%' OR alias_8 ILIKE '%%' || $1 || '%%' OR
			user_field_1 ILIKE '%%' || $1 || '%%' OR user_field_2 ILIKE '%%' || $1 || '%%' OR user_field_3 ILIKE '%%' || $1 || '%%' OR user_field_4 ILIKE '%%' || $1 || '%%' OR
			user_field_5 ILIKE '%%' || $1 || '%%' OR user_field_6 ILIKE '%%' || $1 || '%%' OR user_field_7 ILIKE '%%' || $1 || '%%' OR user_field_8 ILIKE '%%' || $1 || '%%'
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
			&f.SourceAddrA, &f.SourcePortA, &f.MulticastAddrA, &f.GroupPortA,
			&f.SourceAddrB, &f.SourcePortB, &f.MulticastAddrB, &f.GroupPortB,
			&f.NMOSNodeID, &f.NMOSFlowID, &f.NMOSSenderID, &f.NMOSDeviceID,
			&f.NMOSNodeLabel, &f.NMOSNodeDescription,
			&f.NMOSIS04Host, &f.NMOSIS04Port, &f.NMOSIS05Host, &f.NMOSIS05Port,
			&f.NMOSIS04BaseURL, &f.NMOSIS05BaseURL, &f.NMOSIS04Version, &f.NMOSIS05Version,
			&f.NMOSLabel, &f.NMOSDescription, &f.ManagementURL,
			&f.MediaType, &f.ST2110Format, &f.RedundancyGroup,
			&f.DataSource, &f.RDSAddress, &f.RDSAPIURL, &f.RDSVersion,
			&f.SDPURL, &f.SDPCache,
			&f.Alias1, &f.Alias2, &f.Alias3, &f.Alias4, &f.Alias5, &f.Alias6, &f.Alias7, &f.Alias8,
			&f.UserField1, &f.UserField2, &f.UserField3, &f.UserField4, &f.UserField5, &f.UserField6, &f.UserField7, &f.UserField8,
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

func (r *PostgresRepository) CountFlowsFiltered(ctx context.Context, f FlowListFilters) (int, error) {
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if s := strings.TrimSpace(f.FlowStatus); s != "" {
		where = append(where, fmt.Sprintf("flow_status = $%d", i))
		args = append(args, s)
		i++
	}
	if s := strings.TrimSpace(f.Availability); s != "" {
		where = append(where, fmt.Sprintf("availability = $%d", i))
		args = append(args, s)
		i++
	}
	if s := strings.TrimSpace(f.DataSource); s != "" {
		where = append(where, fmt.Sprintf("data_source = $%d", i))
		args = append(args, s)
		i++
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		where = append(where, fmt.Sprintf(`(
			display_name ILIKE '%%' || $%d || '%%' OR
			flow_id::text ILIKE '%%' || $%d || '%%' OR
			multicast_ip ILIKE '%%' || $%d || '%%' OR
			source_ip ILIKE '%%' || $%d || '%%' OR
			note ILIKE '%%' || $%d || '%%' OR
			sdp_url ILIKE '%%' || $%d || '%%'
		)`, i, i, i, i, i, i))
		args = append(args, q)
		i++
	}

	var total int
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*)::int FROM flows WHERE %s`, strings.Join(where, " AND ")), args...).Scan(&total)
	return total, err
}

func (r *PostgresRepository) ListFlowsFiltered(ctx context.Context, filters FlowListFilters, limit, offset int, sortBy, sortOrder string) ([]models.Flow, error) {
	sortBy, sortOrder = normalizeFlowSort(sortBy, sortOrder)
	where := []string{"1=1"}
	args := []any{}
	i := 1

	if s := strings.TrimSpace(filters.FlowStatus); s != "" {
		where = append(where, fmt.Sprintf("flow_status = $%d", i))
		args = append(args, s)
		i++
	}
	if s := strings.TrimSpace(filters.Availability); s != "" {
		where = append(where, fmt.Sprintf("availability = $%d", i))
		args = append(args, s)
		i++
	}
	if s := strings.TrimSpace(filters.DataSource); s != "" {
		where = append(where, fmt.Sprintf("data_source = $%d", i))
		args = append(args, s)
		i++
	}
	if q := strings.TrimSpace(filters.Q); q != "" {
		where = append(where, fmt.Sprintf(`(
			display_name ILIKE '%%' || $%d || '%%' OR
			flow_id::text ILIKE '%%' || $%d || '%%' OR
			multicast_ip ILIKE '%%' || $%d || '%%' OR
			source_ip ILIKE '%%' || $%d || '%%' OR
			note ILIKE '%%' || $%d || '%%' OR
			sdp_url ILIKE '%%' || $%d || '%%'
		)`, i, i, i, i, i, i))
		args = append(args, q)
		i++
	}

	// pagination args
	args = append(args, limit, offset)
	limitArg := i
	offsetArg := i + 1

	query := fmt.Sprintf(`
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol,
			source_addr_a, source_port_a, multicast_addr_a, group_port_a,
			source_addr_b, source_port_b, multicast_addr_b, group_port_b,
			COALESCE(nmos_node_id::text, '') AS nmos_node_id,
			COALESCE(nmos_flow_id::text, '') AS nmos_flow_id,
			COALESCE(nmos_sender_id::text, '') AS nmos_sender_id,
			COALESCE(nmos_device_id::text, '') AS nmos_device_id,
			nmos_node_label, nmos_node_description,
			nmos_is04_host, nmos_is04_port, nmos_is05_host, nmos_is05_port,
			nmos_is04_base_url, nmos_is05_base_url, nmos_is04_version, nmos_is05_version,
			nmos_label, nmos_description, management_url,
			media_type, st2110_format, redundancy_group,
			data_source, rds_address, rds_api_url, rds_version,
			sdp_url, sdp_cache,
			alias_1, alias_2, alias_3, alias_4, alias_5, alias_6, alias_7, alias_8,
			user_field_1, user_field_2, user_field_3, user_field_4, user_field_5, user_field_6, user_field_7, user_field_8
		FROM flows
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), sortBy, sortOrder, limitArg, offsetArg)

	rows, err := r.pool.Query(ctx, query, args...)
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
			&f.SourceAddrA, &f.SourcePortA, &f.MulticastAddrA, &f.GroupPortA,
			&f.SourceAddrB, &f.SourcePortB, &f.MulticastAddrB, &f.GroupPortB,
			&f.NMOSNodeID, &f.NMOSFlowID, &f.NMOSSenderID, &f.NMOSDeviceID,
			&f.NMOSNodeLabel, &f.NMOSNodeDescription,
			&f.NMOSIS04Host, &f.NMOSIS04Port, &f.NMOSIS05Host, &f.NMOSIS05Port,
			&f.NMOSIS04BaseURL, &f.NMOSIS05BaseURL, &f.NMOSIS04Version, &f.NMOSIS05Version,
			&f.NMOSLabel, &f.NMOSDescription, &f.ManagementURL,
			&f.MediaType, &f.ST2110Format, &f.RedundancyGroup,
			&f.DataSource, &f.RDSAddress, &f.RDSAPIURL, &f.RDSVersion,
			&f.SDPURL, &f.SDPCache,
			&f.Alias1, &f.Alias2, &f.Alias3, &f.Alias4, &f.Alias5, &f.Alias6, &f.Alias7, &f.Alias8,
			&f.UserField1, &f.UserField2, &f.UserField3, &f.UserField4, &f.UserField5, &f.UserField6, &f.UserField7, &f.UserField8,
		); err != nil {
			return nil, err
		}
		flows = append(flows, f)
	}
	return flows, rows.Err()
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
			note ILIKE '%' || $1 || '%' OR
			alias_1 ILIKE '%' || $1 || '%' OR alias_2 ILIKE '%' || $1 || '%' OR alias_3 ILIKE '%' || $1 || '%' OR alias_4 ILIKE '%' || $1 || '%' OR
			alias_5 ILIKE '%' || $1 || '%' OR alias_6 ILIKE '%' || $1 || '%' OR alias_7 ILIKE '%' || $1 || '%' OR alias_8 ILIKE '%' || $1 || '%' OR
			user_field_1 ILIKE '%' || $1 || '%' OR user_field_2 ILIKE '%' || $1 || '%' OR user_field_3 ILIKE '%' || $1 || '%' OR user_field_4 ILIKE '%' || $1 || '%' OR
			user_field_5 ILIKE '%' || $1 || '%' OR user_field_6 ILIKE '%' || $1 || '%' OR user_field_7 ILIKE '%' || $1 || '%' OR user_field_8 ILIKE '%' || $1 || '%'
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
			INSERT INTO flows(flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, transport_protocol,
				sdp_url, sdp_cache,
				alias_1, alias_2, alias_3, alias_4, alias_5, alias_6, alias_7, alias_8,
				user_field_1, user_field_2, user_field_3, user_field_4, user_field_5, user_field_6, user_field_7, user_field_8)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)
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
				sdp_url = EXCLUDED.sdp_url, sdp_cache = EXCLUDED.sdp_cache,
				alias_1 = EXCLUDED.alias_1, alias_2 = EXCLUDED.alias_2, alias_3 = EXCLUDED.alias_3, alias_4 = EXCLUDED.alias_4,
				alias_5 = EXCLUDED.alias_5, alias_6 = EXCLUDED.alias_6, alias_7 = EXCLUDED.alias_7, alias_8 = EXCLUDED.alias_8,
				user_field_1 = EXCLUDED.user_field_1, user_field_2 = EXCLUDED.user_field_2, user_field_3 = EXCLUDED.user_field_3, user_field_4 = EXCLUDED.user_field_4,
				user_field_5 = EXCLUDED.user_field_5, user_field_6 = EXCLUDED.user_field_6, user_field_7 = EXCLUDED.user_field_7, user_field_8 = EXCLUDED.user_field_8,
				updated_at = NOW()
		`, flow.FlowID, flow.DisplayName, flow.MulticastIP, flow.SourceIP, flow.Port, flow.FlowStatus, flow.Availability, flow.Locked, flow.Note, flow.TransportProto,
			flow.SDPURL, flow.SDPCache,
			flow.Alias1, flow.Alias2, flow.Alias3, flow.Alias4, flow.Alias5, flow.Alias6, flow.Alias7, flow.Alias8,
			flow.UserField1, flow.UserField2, flow.UserField3, flow.UserField4, flow.UserField5, flow.UserField6, flow.UserField7, flow.UserField8)
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

// NMOS registry (IS-04) implementation

func (r *PostgresRepository) ListNMOSNodes(ctx context.Context) ([]models.NMOSNode, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label, description, hostname, api_version,
		       COALESCE(tags::text, '{}'), COALESCE(meta::text, '{}')
		FROM nmos_nodes
		ORDER BY label ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.NMOSNode
	for rows.Next() {
		var n models.NMOSNode
		if err := rows.Scan(&n.ID, &n.Label, &n.Description, &n.Hostname, &n.APIVersion, &n.Tags, &n.Meta); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListNMOSDevices(ctx context.Context, nodeID string) ([]models.NMOSDevice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label, description, node_id, type,
		       COALESCE(tags::text, '{}'), COALESCE(meta::text, '{}')
		FROM nmos_devices
		WHERE ($1 = '' OR node_id = $1)
		ORDER BY label ASC, id ASC
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.NMOSDevice
	for rows.Next() {
		var d models.NMOSDevice
		if err := rows.Scan(&d.ID, &d.Label, &d.Description, &d.NodeID, &d.Type, &d.Tags, &d.Meta); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListNMOSFlows(ctx context.Context) ([]models.NMOSFlow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label, description, format, source_id,
		       COALESCE(tags::text, '{}'), COALESCE(meta::text, '{}')
		FROM nmos_flows
		ORDER BY label ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.NMOSFlow
	for rows.Next() {
		var f models.NMOSFlow
		if err := rows.Scan(&f.ID, &f.Label, &f.Description, &f.Format, &f.SourceID, &f.Tags, &f.Meta); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListNMOSSenders(ctx context.Context, deviceID string) ([]models.NMOSSender, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label, description, flow_id, transport, manifest_href, device_id,
		       COALESCE(tags::text, '{}'), COALESCE(meta::text, '{}')
		FROM nmos_senders
		WHERE ($1 = '' OR device_id = $1)
		ORDER BY label ASC, id ASC
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.NMOSSender
	for rows.Next() {
		var s models.NMOSSender
		if err := rows.Scan(&s.ID, &s.Label, &s.Description, &s.FlowID, &s.Transport, &s.ManifestHREF, &s.DeviceID, &s.Tags, &s.Meta); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListNMOSReceivers(ctx context.Context, deviceID string) ([]models.NMOSReceiver, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label, description, format, transport, device_id,
		       COALESCE(tags::text, '{}'), COALESCE(meta::text, '{}')
		FROM nmos_receivers
		WHERE ($1 = '' OR device_id = $1)
		ORDER BY label ASC, id ASC
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.NMOSReceiver
	for rows.Next() {
		var rec models.NMOSReceiver
		if err := rows.Scan(&rec.ID, &rec.Label, &rec.Description, &rec.Format, &rec.Transport, &rec.DeviceID, &rec.Tags, &rec.Meta); err != nil {
			return nil, err
		}
		items = append(items, rec)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) UpsertNMOSNode(ctx context.Context, node models.NMOSNode) error {
	if len(node.Tags) == 0 {
		node.Tags = json.RawMessage(`{}`)
	}
	if len(node.Meta) == 0 {
		node.Meta = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO nmos_nodes(id, label, description, hostname, api_version, tags, meta)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			label = EXCLUDED.label,
			description = EXCLUDED.description,
			hostname = EXCLUDED.hostname,
			api_version = EXCLUDED.api_version,
			tags = EXCLUDED.tags,
			meta = EXCLUDED.meta,
			updated_at = NOW()
	`, node.ID, node.Label, node.Description, node.Hostname, node.APIVersion, string(node.Tags), string(node.Meta))
	return err
}

func (r *PostgresRepository) UpsertNMOSDevice(ctx context.Context, dev models.NMOSDevice) error {
	if len(dev.Tags) == 0 {
		dev.Tags = json.RawMessage(`{}`)
	}
	if len(dev.Meta) == 0 {
		dev.Meta = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO nmos_devices(id, node_id, label, description, type, tags, meta)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			node_id = EXCLUDED.node_id,
			label = EXCLUDED.label,
			description = EXCLUDED.description,
			type = EXCLUDED.type,
			tags = EXCLUDED.tags,
			meta = EXCLUDED.meta,
			updated_at = NOW()
	`, dev.ID, dev.NodeID, dev.Label, dev.Description, dev.Type, string(dev.Tags), string(dev.Meta))
	return err
}

func (r *PostgresRepository) UpsertNMOSFlow(ctx context.Context, flow models.NMOSFlow) error {
	if len(flow.Tags) == 0 {
		flow.Tags = json.RawMessage(`{}`)
	}
	if len(flow.Meta) == 0 {
		flow.Meta = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO nmos_flows(id, label, description, format, source_id, tags, meta)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			label = EXCLUDED.label,
			description = EXCLUDED.description,
			format = EXCLUDED.format,
			source_id = EXCLUDED.source_id,
			tags = EXCLUDED.tags,
			meta = EXCLUDED.meta,
			updated_at = NOW()
	`, flow.ID, flow.Label, flow.Description, flow.Format, flow.SourceID, string(flow.Tags), string(flow.Meta))
	return err
}

func (r *PostgresRepository) UpsertNMOSSender(ctx context.Context, sender models.NMOSSender) error {
	if len(sender.Tags) == 0 {
		sender.Tags = json.RawMessage(`{}`)
	}
	if len(sender.Meta) == 0 {
		sender.Meta = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO nmos_senders(id, device_id, flow_id, label, description, transport, manifest_href, tags, meta)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			device_id = EXCLUDED.device_id,
			flow_id = EXCLUDED.flow_id,
			label = EXCLUDED.label,
			description = EXCLUDED.description,
			transport = EXCLUDED.transport,
			manifest_href = EXCLUDED.manifest_href,
			tags = EXCLUDED.tags,
			meta = EXCLUDED.meta,
			updated_at = NOW()
	`, sender.ID, sender.DeviceID, sender.FlowID, sender.Label, sender.Description, sender.Transport, sender.ManifestHREF, string(sender.Tags), string(sender.Meta))
	return err
}

func (r *PostgresRepository) UpsertNMOSReceiver(ctx context.Context, rec models.NMOSReceiver) error {
	if len(rec.Tags) == 0 {
		rec.Tags = json.RawMessage(`{}`)
	}
	if len(rec.Meta) == 0 {
		rec.Meta = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO nmos_receivers(id, device_id, label, description, format, transport, tags, meta)
		VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8::jsonb)
		ON CONFLICT (id) DO UPDATE SET
			device_id = EXCLUDED.device_id,
			label = EXCLUDED.label,
			description = EXCLUDED.description,
			format = EXCLUDED.format,
			transport = EXCLUDED.transport,
			tags = EXCLUDED.tags,
			meta = EXCLUDED.meta,
			updated_at = NOW()
	`, rec.ID, rec.DeviceID, rec.Label, rec.Description, rec.Format, rec.Transport, string(rec.Tags), string(rec.Meta))
	return err
}

func (r *PostgresRepository) CreateFlow(ctx context.Context, flow models.Flow) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO flows(
			flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, transport_protocol,
			source_addr_a, source_port_a, multicast_addr_a, group_port_a,
			source_addr_b, source_port_b, multicast_addr_b, group_port_b,
			nmos_node_id, nmos_flow_id, nmos_sender_id, nmos_device_id,
			nmos_node_label, nmos_node_description,
			nmos_is04_host, nmos_is04_port, nmos_is05_host, nmos_is05_port,
			nmos_is04_base_url, nmos_is05_base_url, nmos_is04_version, nmos_is05_version,
			nmos_label, nmos_description, management_url,
			media_type, st2110_format, redundancy_group,
			data_source, rds_address, rds_api_url, rds_version,
			sdp_url, sdp_cache,
			alias_1, alias_2, alias_3, alias_4, alias_5, alias_6, alias_7, alias_8,
			user_field_1, user_field_2, user_field_3, user_field_4, user_field_5, user_field_6, user_field_7, user_field_8)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,
			$15,$16,$17,$18,
			NULLIF($19,'')::uuid, NULLIF($20,'')::uuid, NULLIF($21,'')::uuid, NULLIF($22,'')::uuid,
			$23,$24,
			$25,$26,$27,$28,
			$29,$30,$31,$32,
			$33,$34,$35,
			$36,$37,$38,
			$39,$40,$41,$42,
			$43,$44,
			$45,$46,$47,$48,$49,$50,$51,$52,
			$53,$54,$55,$56,$57,$58,$59,$60
		)
		RETURNING id
	`, flow.FlowID, flow.DisplayName, flow.MulticastIP, flow.SourceIP, flow.Port, flow.FlowStatus, flow.Availability, flow.Locked, flow.Note, flow.TransportProto,
		flow.SourceAddrA, flow.SourcePortA, flow.MulticastAddrA, flow.GroupPortA,
		flow.SourceAddrB, flow.SourcePortB, flow.MulticastAddrB, flow.GroupPortB,
		flow.NMOSNodeID, flow.NMOSFlowID, flow.NMOSSenderID, flow.NMOSDeviceID,
		flow.NMOSNodeLabel, flow.NMOSNodeDescription,
		flow.NMOSIS04Host, flow.NMOSIS04Port, flow.NMOSIS05Host, flow.NMOSIS05Port,
		flow.NMOSIS04BaseURL, flow.NMOSIS05BaseURL, flow.NMOSIS04Version, flow.NMOSIS05Version,
		flow.NMOSLabel, flow.NMOSDescription, flow.ManagementURL,
		flow.MediaType, flow.ST2110Format, flow.RedundancyGroup,
		flow.DataSource, flow.RDSAddress, flow.RDSAPIURL, flow.RDSVersion,
		flow.SDPURL, flow.SDPCache,
		flow.Alias1, flow.Alias2, flow.Alias3, flow.Alias4, flow.Alias5, flow.Alias6, flow.Alias7, flow.Alias8,
		flow.UserField1, flow.UserField2, flow.UserField3, flow.UserField4, flow.UserField5, flow.UserField6, flow.UserField7, flow.UserField8).Scan(&id)
	return id, err
}

func (r *PostgresRepository) GetFlowByID(ctx context.Context, id int64) (*models.Flow, error) {
	var f models.Flow
	err := r.pool.QueryRow(ctx, `
		SELECT id, flow_id, display_name, multicast_ip, source_ip, port, flow_status, availability, locked, note, updated_at, last_seen, transport_protocol,
			source_addr_a, source_port_a, multicast_addr_a, group_port_a,
			source_addr_b, source_port_b, multicast_addr_b, group_port_b,
			COALESCE(nmos_node_id::text, '') AS nmos_node_id,
			COALESCE(nmos_flow_id::text, '') AS nmos_flow_id,
			COALESCE(nmos_sender_id::text, '') AS nmos_sender_id,
			COALESCE(nmos_device_id::text, '') AS nmos_device_id,
			nmos_node_label, nmos_node_description,
			nmos_is04_host, nmos_is04_port, nmos_is05_host, nmos_is05_port,
			nmos_is04_base_url, nmos_is05_base_url, nmos_is04_version, nmos_is05_version,
			nmos_label, nmos_description, management_url,
			media_type, st2110_format, redundancy_group,
			data_source, rds_address, rds_api_url, rds_version,
			sdp_url, sdp_cache,
			alias_1, alias_2, alias_3, alias_4, alias_5, alias_6, alias_7, alias_8,
			user_field_1, user_field_2, user_field_3, user_field_4, user_field_5, user_field_6, user_field_7, user_field_8
		FROM flows WHERE id = $1
	`, id).Scan(
		&f.ID, &f.FlowID, &f.DisplayName, &f.MulticastIP, &f.SourceIP, &f.Port,
		&f.FlowStatus, &f.Availability, &f.Locked, &f.Note, &f.UpdatedAt, &f.LastSeen, &f.TransportProto,
		&f.SourceAddrA, &f.SourcePortA, &f.MulticastAddrA, &f.GroupPortA,
		&f.SourceAddrB, &f.SourcePortB, &f.MulticastAddrB, &f.GroupPortB,
		&f.NMOSNodeID, &f.NMOSFlowID, &f.NMOSSenderID, &f.NMOSDeviceID,
		&f.NMOSNodeLabel, &f.NMOSNodeDescription,
		&f.NMOSIS04Host, &f.NMOSIS04Port, &f.NMOSIS05Host, &f.NMOSIS05Port,
		&f.NMOSIS04BaseURL, &f.NMOSIS05BaseURL, &f.NMOSIS04Version, &f.NMOSIS05Version,
		&f.NMOSLabel, &f.NMOSDescription, &f.ManagementURL,
		&f.MediaType, &f.ST2110Format, &f.RedundancyGroup,
		&f.DataSource, &f.RDSAddress, &f.RDSAPIURL, &f.RDSVersion,
		&f.SDPURL, &f.SDPCache,
		&f.Alias1, &f.Alias2, &f.Alias3, &f.Alias4, &f.Alias5, &f.Alias6, &f.Alias7, &f.Alias8,
		&f.UserField1, &f.UserField2, &f.UserField3, &f.UserField4, &f.UserField5, &f.UserField6, &f.UserField7, &f.UserField8,
	)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *PostgresRepository) PatchFlow(ctx context.Context, id int64, updates map[string]any) error {
	allowed := map[string]bool{
		"display_name":          true,
		"multicast_ip":          true,
		"source_ip":             true,
		"port":                  true,
		"flow_status":           true,
		"availability":          true,
		"locked":                true,
		"note":                  true,
		"transport_protocol":    true,
		"source_addr_a":         true,
		"source_port_a":         true,
		"multicast_addr_a":      true,
		"group_port_a":          true,
		"source_addr_b":         true,
		"source_port_b":         true,
		"multicast_addr_b":      true,
		"group_port_b":          true,
		"nmos_node_id":          true,
		"nmos_flow_id":          true,
		"nmos_sender_id":        true,
		"nmos_device_id":        true,
		"nmos_node_label":       true,
		"nmos_node_description": true,
		"nmos_is04_host":        true,
		"nmos_is04_port":        true,
		"nmos_is05_host":        true,
		"nmos_is05_port":        true,
		"nmos_is04_base_url":    true,
		"nmos_is05_base_url":    true,
		"nmos_is04_version":     true,
		"nmos_is05_version":     true,
		"nmos_label":            true,
		"nmos_description":      true,
		"management_url":        true,
		"media_type":            true,
		"st2110_format":         true,
		"redundancy_group":      true,
		"data_source":           true,
		"rds_address":           true,
		"rds_api_url":           true,
		"rds_version":           true,
		"sdp_url":               true,
		"sdp_cache":             true,
		"alias_1":               true,
		"alias_2":               true,
		"alias_3":               true,
		"alias_4":               true,
		"alias_5":               true,
		"alias_6":               true,
		"alias_7":               true,
		"alias_8":               true,
		"user_field_1":          true,
		"user_field_2":          true,
		"user_field_3":          true,
		"user_field_4":          true,
		"user_field_5":          true,
		"user_field_6":          true,
		"user_field_7":          true,
		"user_field_8":          true,
	}
	uuidCols := map[string]bool{
		"nmos_node_id":   true,
		"nmos_flow_id":   true,
		"nmos_sender_id": true,
		"nmos_device_id": true,
	}

	setClauses := []string{}
	args := []any{}
	i := 1
	for key, value := range updates {
		if !allowed[key] {
			continue
		}
		if uuidCols[key] {
			// Allow passing UUIDs as strings; empty string clears the value.
			setClauses = append(setClauses, fmt.Sprintf("%s = NULLIF($%d,'')::uuid", key, i))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, i))
		}
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
