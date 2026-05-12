package app

import "fmt"

type AlreadyInitializedError struct{}

func (e AlreadyInitializedError) Error() string {
	return "app already initialized"
}

type NotInitializedError struct{}

func (e NotInitializedError) Error() string {
	return "app not initialized"
}

type AlreadyStartedError struct{}

func (e AlreadyStartedError) Error() string {
	return "app already started"
}

type InitError struct {
	Step  string
	Cause error
}

func (e InitError) Error() string {
	return fmt.Sprintf("app init failed: step=%s: %v", e.Step, e.Cause)
}

func (e InitError) Unwrap() error {
	return e.Cause
}

type StartError struct {
	Step  string
	Cause error
}

func (e StartError) Error() string {
	return fmt.Sprintf("app start failed: step=%s: %v", e.Step, e.Cause)
}

func (e StartError) Unwrap() error {
	return e.Cause
}
