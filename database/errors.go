package database

import "fmt"

type UnsupportedDriverError struct {
	Driver string
}

func (e UnsupportedDriverError) Error() string {
	return fmt.Sprintf("unsupported database driver: %q", e.Driver)
}

type OpenError struct {
	Cause error
}

func (e OpenError) Error() string {
	return fmt.Sprintf("database open failed: %v", e.Cause)
}

func (e OpenError) Unwrap() error {
	return e.Cause
}

type MigrationError struct {
	Strategy string
	Cause    error
}

func (e MigrationError) Error() string {
	return fmt.Sprintf("database migration failed: strategy=%s: %v", e.Strategy, e.Cause)
}

func (e MigrationError) Unwrap() error {
	return e.Cause
}
