package config

import "fmt"

type EnvVar string

type DotenvLoadError struct {
	Path  string
	Cause error
}

func (e DotenvLoadError) Error() string {
	return fmt.Sprintf("dotenv load failed: path=%q: %v", e.Path, e.Cause)
}

func (e DotenvLoadError) Unwrap() error {
	return e.Cause
}

type InvalidEnvValueError struct {
	Name     EnvVar
	Value    string
	Expected string
}

func (e InvalidEnvValueError) Error() string {
	return fmt.Sprintf("invalid env value: %s=%q (expected %s)", e.Name, e.Value, e.Expected)
}

type Field int

const (
	FieldUnspecified Field = iota
	FieldLog
	FieldHTTP
	FieldDatabase
	FieldCache
	FieldS3
	FieldCron
	FieldQueue
)

type ValidationError struct {
	Field   Field
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("config validation failed: field=%s: %s", e.Field.String(), e.Message)
}

func (f Field) String() string {
	switch f {
	case FieldLog:
		return "log"
	case FieldHTTP:
		return "http"
	case FieldDatabase:
		return "database"
	case FieldCache:
		return "cache"
	case FieldS3:
		return "s3"
	case FieldCron:
		return "cron"
	case FieldQueue:
		return "queue"
	case FieldUnspecified:
		return "unspecified"
	default:
		return "unspecified"
	}
}
