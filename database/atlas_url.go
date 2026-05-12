package database

import (
	"strings"

	"github.com/Xusk947/bubble/config"
)

func ResolveAtlasURL(cfg config.DatabaseConfig) (string, error) {
	url := strings.TrimSpace(cfg.AtlasURL)
	if url != "" {
		return url, nil
	}

	dsn := strings.TrimSpace(cfg.DSN)
	if strings.Contains(dsn, "://") {
		return dsn, nil
	}

	return "", MigrationError{Strategy: "atlas", Cause: config.ValidationError{Field: config.FieldDatabase, Message: "atlas url is empty"}}
}
