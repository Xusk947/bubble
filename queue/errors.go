package queue

import "fmt"

type InvalidConfigError struct {
	Field   string
	Message string
}

func (e InvalidConfigError) Error() string {
	return fmt.Sprintf("queue config invalid: field=%s: %s", e.Field, e.Message)
}

type OperationError struct {
	Op    string
	Cause error
}

func (e OperationError) Error() string {
	return fmt.Sprintf("queue operation failed: op=%s: %v", e.Op, e.Cause)
}

func (e OperationError) Unwrap() error {
	return e.Cause
}

