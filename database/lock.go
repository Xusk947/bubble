package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"bubble/config"
)

const mysqlLockTimeoutSeconds = 60

type ReleaseFunc func()

func AcquireLock(ctx context.Context, conn *sql.Conn, cfg config.DatabaseConfig) (ReleaseFunc, error) {
	if !cfg.Migrations.Lock.Enabled {
		return func() {}, nil
	}

	key := cfg.Migrations.Lock.Key
	driver := strings.TrimSpace(strings.ToLower(cfg.Driver))
	switch driver {
	case "postgres", "pgx":
		if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", key); err != nil {
			return nil, err
		}
		return func() {
			_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", key)
		}, nil
	case "mysql":
		lockName := fmt.Sprintf("bubble_migrations_%d", key)
		var acquired sql.NullInt64
		if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", lockName, mysqlLockTimeoutSeconds).Scan(&acquired); err != nil {
			return nil, err
		}
		if !acquired.Valid || acquired.Int64 != 1 {
			return nil, fmt.Errorf("mysql migration lock not acquired: %s", lockName)
		}
		return func() {
			_, _ = conn.ExecContext(context.Background(), "SELECT RELEASE_LOCK(?)", lockName)
		}, nil
	default:
		return func() {}, nil
	}
}

