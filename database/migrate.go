package database

import (
	"context"
	"database/sql"
	"os"
	"strings"

	"bubble/config"

	"ariga.io/atlas/atlasexec"
	"go.uber.org/zap"
)

const atlasBinaryName = "atlas"

type Migrator interface {
	Up(ctx context.Context) (int, error)
}

type AtlasMigrator struct {
	DB     *sql.DB
	Config config.DatabaseConfig
	Logger *zap.Logger
}

func (m AtlasMigrator) Up(ctx context.Context) (int, error) {
	if m.DB == nil {
		return 0, MigrationError{Strategy: "atlas", Cause: OpenError{Cause: sql.ErrConnDone}}
	}

	if strings.TrimSpace(m.Config.Migrations.Dir) == "" {
		return 0, MigrationError{Strategy: "atlas", Cause: config.ValidationError{Field: config.FieldDatabase, Message: "migrations dir is empty"}}
	}

	conn, err := m.DB.Conn(ctx)
	if err != nil {
		return 0, MigrationError{Strategy: "atlas", Cause: err}
	}
	defer conn.Close()

	release, err := AcquireLock(ctx, conn, m.Config)
	if err != nil {
		return 0, MigrationError{Strategy: "atlas", Cause: err}
	}
	defer release()

	workdir, err := atlasexec.NewWorkingDir(
		atlasexec.WithMigrations(os.DirFS(m.Config.Migrations.Dir)),
	)
	if err != nil {
		return 0, MigrationError{Strategy: "atlas", Cause: err}
	}
	defer workdir.Close()

	client, err := atlasexec.NewClient(workdir.Path(), atlasBinaryName)
	if err != nil {
		return 0, MigrationError{Strategy: "atlas", Cause: err}
	}

	url, err := ResolveAtlasURL(m.Config)
	if err != nil {
		return 0, err
	}

	res, err := client.MigrateApply(ctx, &atlasexec.MigrateApplyParams{
		URL: url,
	})
	if err != nil {
		return 0, MigrationError{Strategy: "atlas", Cause: err}
	}

	if m.Logger != nil {
		m.Logger.Info(
			"database migrations applied",
			zap.String("strategy", "atlas"),
			zap.Int("applied", len(res.Applied)),
		)
	}

	return len(res.Applied), nil
}
