package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeProjectFiles(dir string, f Flags) error {
	appName := sanitizeName(f.Name)
	bubbleModule := strings.TrimSpace(f.BubbleModule)
	if bubbleModule == "" {
		bubbleModule = "bubble"
	}

	db := strings.TrimSpace(strings.ToLower(f.DB))
	cache := strings.TrimSpace(strings.ToLower(f.Cache))
	queue := strings.TrimSpace(strings.ToLower(f.Queue))
	driverImport := ""
	if db == dbSQLite {
		driverImport = "\n\t_ \"github.com/mattn/go-sqlite3\"\n"
	}
	if db == dbPostgres {
		driverImport = "\n\t_ \"github.com/jackc/pgx/v5/stdlib\"\n"
	}

	cmdDir := filepath.Join(dir, "cmd", appName)
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}
	internalDir := filepath.Join(dir, "internal", "api")
	if err := os.MkdirAll(internalDir, 0755); err != nil {
		return err
	}

	mainGo := fmt.Sprintf(`package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"%s/internal/api"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a := api.New()
	if err := a.Start(ctx); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	_ = a.Stop(context.Background())
}
`, f.Module)

	apiGo := fmt.Sprintf(`package api

import (
	"context"

	"%s/app"
	"%s/httpserver"
%s

	"github.com/gofiber/fiber/v3"
)

type API struct {
	App app.App
}

func New() *API {
	return &API{}
}

func (a *API) Start(ctx context.Context) error {
	if err := a.App.Init(ctx, app.WithHTTPSetup(func(s *httpserver.HTTPServer) {
		s.App.Get("/hello", func(c fiber.Ctx) error {
			return c.SendString("hello")
		})
	})); err != nil {
		return err
	}
	if err := a.App.Start(ctx); err != nil {
		return err
	}
	return nil
}
`, bubbleModule, bubbleModule, driverImport)

	stopGo := fmt.Sprintf(`package api

import "context"

func (a *API) Stop(ctx context.Context) error {
	return a.App.Stop(ctx)
}
`)

	env := buildEnv(db, cache, queue)
	envExample := buildEnvExample()

	readme := fmt.Sprintf(`# %s

Run:

    go run ./cmd/%s

Migrations (Ent+Atlas):

    bubble db diff --name init --dev-url "docker://postgres/16/dev" --to "ent://./ent/schema" --dir ./migrations
    bubble db apply --dir ./migrations --url "postgres://user:pass@localhost:5432/db?sslmode=disable"
`, appName, appName)

	files := []struct {
		path string
		data []byte
		perm os.FileMode
	}{
		{path: filepath.Join(cmdDir, "main.go"), data: []byte(mainGo), perm: 0644},
		{path: filepath.Join(dir, "internal", "api", "api.go"), data: []byte(apiGo), perm: 0644},
		{path: filepath.Join(dir, "internal", "api", "stop.go"), data: []byte(stopGo), perm: 0644},
		{path: filepath.Join(dir, ".env"), data: []byte(env), perm: 0600},
		{path: filepath.Join(dir, ".env.example"), data: []byte(envExample), perm: 0644},
		{path: filepath.Join(dir, "README.md"), data: []byte(readme), perm: 0644},
	}

	for _, file := range files {
		if err := os.WriteFile(file.path, file.data, file.perm); err != nil {
			return err
		}
	}

	migrationsDir := filepath.Join(dir, "migrations")
	return os.MkdirAll(migrationsDir, 0755)
}

func sanitizeName(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return "app"
	}
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func buildEnv(db string, cache string, queue string) string {
	cacheProvider := cacheLocal
	if cache == cacheRedis {
		cacheProvider = cacheRedis
	}

	base := "BUBBLE_HTTP_ADDRESS=:8080\nBUBBLE_CACHE_PROVIDER=" + cacheProvider + "\nBUBBLE_DB_AUTO_MIGRATE=false\n"
	switch db {
	case dbSQLite:
		base = base + "BUBBLE_DB_DRIVER=sqlite3\nBUBBLE_DB_DSN=file:./dev.db?_fk=1&cache=shared\n"
	case dbPostgres:
		base = base + "BUBBLE_DB_DRIVER=pgx\nBUBBLE_DB_DSN=postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable\n"
	default:
	}

	switch queue {
	case queueNATS:
		base = base + "BUBBLE_QUEUE_PROVIDER=nats\nBUBBLE_NATS_URL=nats://127.0.0.1:4222\nBUBBLE_NATS_STREAM=events\nBUBBLE_NATS_SUBJECT=events.*\nBUBBLE_NATS_DURABLE=api\n"
	case queueKafka:
		base = base + "BUBBLE_QUEUE_PROVIDER=kafka\nBUBBLE_KAFKA_BROKERS=127.0.0.1:9092\nBUBBLE_KAFKA_TOPIC=events\nBUBBLE_KAFKA_GROUP_ID=api\n"
	default:
	}
	return base
}

func buildEnvExample() string {
	return "BUBBLE_DOTENV=\n" +
		"\n" +
		"BUBBLE_LOG_LEVEL=info\n" +
		"BUBBLE_LOG_DEVELOPMENT=false\n" +
		"BUBBLE_LOG_ENCODING=json\n" +
		"BUBBLE_LOG_STDOUT=true\n" +
		"BUBBLE_LOG_STDERR=false\n" +
		"\n" +
		"BUBBLE_HTTP_ADDRESS=:8080\n" +
		"BUBBLE_HTTP_READ_TIMEOUT=15s\n" +
		"BUBBLE_HTTP_WRITE_TIMEOUT=15s\n" +
		"BUBBLE_HTTP_IDLE_TIMEOUT=60s\n" +
		"BUBBLE_HTTP_ENABLE_PPROF=false\n" +
		"BUBBLE_HTTP_ENABLE_CORS=false\n" +
		"\n" +
		"BUBBLE_DB_DRIVER=\n" +
		"BUBBLE_DB_DSN=\n" +
		"BUBBLE_DB_ATLAS_URL=\n" +
		"BUBBLE_DB_AUTO_MIGRATE=false\n" +
		"BUBBLE_DB_MIGRATIONS_DIR=./migrations\n" +
		"\n" +
		"BUBBLE_CACHE_PROVIDER=local\n" +
		"BUBBLE_REDIS_ADDRESS=127.0.0.1:6379\n" +
		"\n" +
		"BUBBLE_S3_ENDPOINT=\n" +
		"BUBBLE_S3_REGION=us-east-1\n" +
		"BUBBLE_S3_BUCKET=\n" +
		"BUBBLE_S3_ACCESS_KEY_ID=\n" +
		"BUBBLE_S3_SECRET_ACCESS_KEY=\n" +
		"\n" +
		"BUBBLE_QUEUE_PROVIDER=\n" +
		"BUBBLE_NATS_URL=nats://127.0.0.1:4222\n" +
		"BUBBLE_KAFKA_BROKERS=127.0.0.1:9092\n"
}
