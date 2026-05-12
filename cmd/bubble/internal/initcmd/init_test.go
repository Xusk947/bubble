package initcmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInferBubbleModuleFromAppModule(t *testing.T) {
	got := inferBubbleModuleFromAppModule("github.com/xusk947/hello-world")
	if got != "github.com/Xusk947/bubble" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestRun_NonInteractive_SkipGo_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "hello")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := Run([]string{
		"--name", "hello",
		"--module", "github.com/xusk947/hello",
		"--db", "sqlite",
		"--cache", "local",
		"--queue", "none",
		"--s3=false",
		"--dir", target,
		"--skip-go",
		"--non-interactive",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run: %v (stderr=%s)", err, stderr.String())
	}

	envPath := filepath.Join(target, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("read .env: %v", err)
	}
	env := string(data)
	if !strings.Contains(env, "BUBBLE_DB_DRIVER=sqlite3") {
		t.Fatalf("expected sqlite config in .env, got:\n%s", env)
	}
	if !strings.Contains(env, "BUBBLE_CACHE_PROVIDER=local") {
		t.Fatalf("expected cache local in .env, got:\n%s", env)
	}
}
