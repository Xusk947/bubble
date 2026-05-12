package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"bubble/logging"
)

func TestLoad_UsesDotenvWhenProvided(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := []byte("BUBBLE_LOG_LEVEL=debug\nBUBBLE_HTTP_ADDRESS=:9999\nBUBBLE_CACHE_PROVIDER=local\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	cfg, err := Load(WithDotenvPath(path))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Log.Level != logging.LevelDebug {
		t.Fatalf("unexpected log level: %v", cfg.Log.Level)
	}
	if cfg.HTTP.Address != ":9999" {
		t.Fatalf("unexpected http address: %q", cfg.HTTP.Address)
	}
}

func TestLoad_ParsesDurations(t *testing.T) {
	t.Setenv("BUBBLE_HTTP_ADDRESS", ":8081")
	t.Setenv("BUBBLE_HTTP_READ_TIMEOUT", "1s")
	t.Setenv("BUBBLE_HTTP_WRITE_TIMEOUT", "2s")
	t.Setenv("BUBBLE_HTTP_IDLE_TIMEOUT", "3s")
	t.Setenv("BUBBLE_CACHE_PROVIDER", "local")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.HTTP.ReadTimeout != time.Second {
		t.Fatalf("unexpected read timeout: %v", cfg.HTTP.ReadTimeout)
	}
	if cfg.HTTP.WriteTimeout != 2*time.Second {
		t.Fatalf("unexpected write timeout: %v", cfg.HTTP.WriteTimeout)
	}
	if cfg.HTTP.IdleTimeout != 3*time.Second {
		t.Fatalf("unexpected idle timeout: %v", cfg.HTTP.IdleTimeout)
	}
}

func TestLoad_OverrideOptionWins(t *testing.T) {
	t.Setenv("BUBBLE_HTTP_ADDRESS", ":8082")
	t.Setenv("BUBBLE_CACHE_PROVIDER", "local")

	override := HTTPConfig{
		Address:      ":7777",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 11 * time.Second,
		IdleTimeout:  12 * time.Second,
	}

	cfg, err := Load(WithHTTPConfig(override))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.HTTP.Address != ":7777" {
		t.Fatalf("override not applied: %q", cfg.HTTP.Address)
	}
}

