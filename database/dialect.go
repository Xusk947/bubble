package database

import (
	"strings"

	"entgo.io/ent/dialect"
)

func DialectFromDriver(driver string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(driver))
	switch normalized {
	case "postgres", "pgx":
		return dialect.Postgres, nil
	case "mysql":
		return dialect.MySQL, nil
	case "sqlite3", "sqlite":
		return dialect.SQLite, nil
	default:
		return "", UnsupportedDriverError{Driver: driver}
	}
}

