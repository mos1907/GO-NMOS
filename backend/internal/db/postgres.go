package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, file := range files {
		var exists int
		err = pool.QueryRow(ctx, "SELECT 1 FROM schema_migrations WHERE version = $1", file).Scan(&exists)
		if err == nil {
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("migration state check failed for %s: %w", file, err)
		}

		content, readErr := os.ReadFile(filepath.Join(migrationsDir, file))
		if readErr != nil {
			return fmt.Errorf("read migration %s: %w", file, readErr)
		}

		tx, txErr := pool.Begin(ctx)
		if txErr != nil {
			return txErr
		}

		if _, execErr := tx.Exec(ctx, string(content)); execErr != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", file, execErr)
		}

		if _, insErr := tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", file); insErr != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", file, insErr)
		}

		if commitErr := tx.Commit(ctx); commitErr != nil {
			return commitErr
		}
	}

	return nil
}
