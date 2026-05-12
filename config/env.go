package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func envString(name EnvVar, def string) string {
	value, ok := os.LookupEnv(string(name))
	if !ok {
		return def
	}
	return value
}

func envBool(name EnvVar, def bool) (bool, error) {
	value, ok := os.LookupEnv(string(name))
	if !ok {
		return def, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, InvalidEnvValueError{
			Name:     name,
			Value:    value,
			Expected: "bool",
		}
	}
	return parsed, nil
}

func envInt(name EnvVar, def int) (int, error) {
	value, ok := os.LookupEnv(string(name))
	if !ok {
		return def, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, InvalidEnvValueError{
			Name:     name,
			Value:    value,
			Expected: "int",
		}
	}
	return parsed, nil
}

func envInt64(name EnvVar, def int64) (int64, error) {
	value, ok := os.LookupEnv(string(name))
	if !ok {
		return def, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, InvalidEnvValueError{
			Name:     name,
			Value:    value,
			Expected: "int64",
		}
	}
	return parsed, nil
}

func envDuration(name EnvVar, def time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(string(name))
	if !ok {
		return def, nil
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, InvalidEnvValueError{
			Name:     name,
			Value:    value,
			Expected: "duration",
		}
	}
	return parsed, nil
}
