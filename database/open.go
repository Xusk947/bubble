package database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"bubble/config"

	entsql "entgo.io/ent/dialect/sql"
	"go.uber.org/zap"
)

type Handles struct {
	SQL       *sql.DB
	EntDriver *entsql.Driver
}

func OpenEntDriver(ctx context.Context, cfg config.DatabaseConfig, logger *zap.Logger) (*entsql.Driver, *sql.DB, error) {
	handles, err := Open(ctx, cfg, logger)
	if err != nil {
		return nil, nil, err
	}
	return handles.EntDriver, handles.SQL, nil
}

func Open(ctx context.Context, cfg config.DatabaseConfig, logger *zap.Logger) (Handles, error) {
	if strings.TrimSpace(cfg.Driver) == "" || strings.TrimSpace(cfg.DSN) == "" {
		return Handles{}, nil
	}

	ddlDialect, err := DialectFromDriver(cfg.Driver)
	if err != nil {
		return Handles{}, err
	}

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return Handles{}, OpenError{Cause: err}
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return Handles{}, OpenError{Cause: err}
	}

	if logger != nil {
		logger.Info(
			"database connected",
			zap.String("db_driver", cfg.Driver),
			zap.Int("db_max_open_conns", cfg.MaxOpenConns),
			zap.Int("db_max_idle_conns", cfg.MaxIdleConns),
			zap.Duration("db_conn_max_lifetime", cfg.ConnMaxLifetime),
		)
	}

	return Handles{
		SQL:       db,
		EntDriver: entsql.OpenDB(ddlDialect, db),
	}, nil
}

func Close(handles Handles) error {
	if handles.SQL == nil {
		return nil
	}
	return handles.SQL.Close()
}
