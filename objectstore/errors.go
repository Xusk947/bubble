package objectstore

import "fmt"

type InvalidConfigError struct {
	Field   string
	Message string
}

func (e InvalidConfigError) Error() string {
	return fmt.Sprintf("objectstore config invalid: field=%s: %s", e.Field, e.Message)
}

type OperationError struct {
	Op    string
	Key   string
	Cause error
}

func (e OperationError) Error() string {
	return fmt.Sprintf("objectstore operation failed: op=%s key=%q: %v", e.Op, e.Key, e.Cause)
}

func (e OperationError) Unwrap() error {
	return e.Cause
}
