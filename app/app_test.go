package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"bubble/logging"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestAppInit_LoadsConfigAndLogsStartup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")

	content := []byte("BUBBLE_LOG_LEVEL=debug\nBUBBLE_HTTP_ADDRESS=:9998\nBUBBLE_CACHE_PROVIDER=local\n")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("write dotenv file: %v", err)
	}

	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	var a App
	if err := a.Init(context.Background(), WithDotenvPath(path), WithLogger(logger)); err != nil {
		t.Fatalf("init: %v", err)
	}

	if !a.Initialized {
		t.Fatalf("expected initialized=true")
	}
	if a.Config.HTTP.Address != ":9998" {
		t.Fatalf("unexpected http address: %q", a.Config.HTTP.Address)
	}
	if a.Config.Log.Level != logging.LevelDebug {
		t.Fatalf("unexpected log level: %v", a.Config.Log.Level)
	}

	entries := logs.All()
	if len(entries) == 0 {
		t.Fatalf("expected startup logs")
	}
	if entries[0].Message != "app initialized" {
		t.Fatalf("unexpected message: %q", entries[0].Message)
	}
}

func TestAppInit_AlreadyInitialized(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	var a App
	if err := a.Init(context.Background(), WithLogger(logger)); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := a.Init(context.Background(), WithLogger(logger)); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAppStart_NoDatabaseConfig(t *testing.T) {
	t.Setenv("BUBBLE_HTTP_ADDRESS", ":0")
	t.Setenv("BUBBLE_CACHE_PROVIDER", "local")

	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	var a App
	if err := a.Init(context.Background(), WithLogger(logger)); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := a.Stop(context.Background()); err != nil {
		t.Fatalf("stop: %v", err)
	}
}
