package app

import (
	"context"
	"database/sql"
	"runtime"
	"strings"
	"time"

	"bubble/config"
	"bubble/database"
	"bubble/httpserver"
	"bubble/logging"
	"bubble/objectstore"
	"bubble/objectstore/s3store"
	"bubble/queue/kafka"
	"bubble/queue/natsjs"

	entsql "entgo.io/ent/dialect/sql"
	"go.uber.org/zap"
)

type App struct {
	Config      config.Config
	Logger      *zap.Logger
	Initialized bool
	StartedAt   time.Time
	SQLDB       *sql.DB
	EntDriver   *entsql.Driver
	ObjectStore objectstore.ObjectStore
	NATS        *natsjs.Client
	Kafka       *kafka.Client
	HTTP        *httpserver.HTTPServer
	HTTPSetup   []func(*httpserver.HTTPServer)
	Started     bool
}

func (a *App) Init(ctx context.Context, opts ...Option) error {
	if a.Initialized {
		return AlreadyInitializedError{}
	}

	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	cfg, err := a.resolveConfig(o)
	if err != nil {
		return err
	}

	logger, err := a.resolveLogger(cfg, o)
	if err != nil {
		return err
	}

	a.Config = cfg
	a.Logger = logger
	a.HTTPSetup = o.httpSetup
	a.StartedAt = time.Now().UTC()
	a.Initialized = true

	a.logStartup(ctx)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.Stop(ctx)
}

func (a *App) Start(ctx context.Context) error {
	if !a.Initialized {
		return NotInitializedError{}
	}
	if a.Started {
		return AlreadyStartedError{}
	}

	handles, err := database.Open(ctx, a.Config.Database, a.Logger)
	if err != nil {
		return StartError{Step: "database.open", Cause: err}
	}
	a.SQLDB = handles.SQL
	a.EntDriver = handles.EntDriver

	if a.Config.Database.AutoMigrate && a.SQLDB != nil {
		switch a.Config.Database.Migrations.Strategy {
		case config.MigrationStrategyAtlas:
			m := database.AtlasMigrator{
				DB:     a.SQLDB,
				Config: a.Config.Database,
				Logger: a.Logger,
			}
			if _, err := m.Up(ctx); err != nil {
				return StartError{Step: "database.migrate", Cause: err}
			}
		default:
			return StartError{Step: "database.migrate", Cause: config.ValidationError{Field: config.FieldDatabase, Message: "migrations strategy is unspecified"}}
		}
	}

	if strings.TrimSpace(a.Config.S3.Bucket) != "" {
		store, err := s3store.New(ctx, a.Config.S3, a.Logger)
		if err != nil {
			return StartError{Step: "s3.init", Cause: err}
		}
		a.ObjectStore = store
	}

	switch a.Config.Queue.Provider {
	case config.QueueProviderNATS:
		c, err := natsjs.Connect(ctx, a.Config.Queue.NATS, a.Logger)
		if err != nil {
			return StartError{Step: "queue.nats.connect", Cause: err}
		}
		a.NATS = c
	case config.QueueProviderKafka:
		c, err := kafka.NewClient(a.Config.Queue.Kafka, a.Logger)
		if err != nil {
			return StartError{Step: "queue.kafka.init", Cause: err}
		}
		if err := c.Ping(ctx); err != nil {
			return StartError{Step: "queue.kafka.ping", Cause: err}
		}
		a.Kafka = c
	case config.QueueProviderUnspecified:
	default:
	}

	srv, err := httpserver.NewHTTPServer(a.Config.HTTP, httpserver.Deps{
		Logger: a.Logger,
		Health: a,
	})
	if err != nil {
		return StartError{Step: "http.init", Cause: err}
	}
	for _, setup := range a.HTTPSetup {
		setup(srv)
	}
	if err := srv.Start(ctx); err != nil {
		return StartError{Step: "http.start", Cause: err}
	}
	a.HTTP = srv

	a.Started = true
	return nil
}

func (a *App) Stop(ctx context.Context) error {
	if a.HTTP != nil {
		_ = a.HTTP.Stop(ctx)
		a.HTTP = nil
	}
	if a.SQLDB != nil {
		_ = a.SQLDB.Close()
		a.SQLDB = nil
		a.EntDriver = nil
	}
	a.ObjectStore = nil
	if a.NATS != nil {
		a.NATS.Close()
		a.NATS = nil
	}
	a.Kafka = nil
	a.Started = false

	if a.Logger == nil {
		return nil
	}
	return a.Logger.Sync()
}

func (a *App) DB() *sql.DB {
	return a.SQLDB
}

func (a *App) Ent() *entsql.Driver {
	return a.EntDriver
}

func (a *App) S3() objectstore.ObjectStore {
	return a.ObjectStore
}

func (a *App) NATSClient() *natsjs.Client {
	return a.NATS
}

func (a *App) KafkaClient() *kafka.Client {
	return a.Kafka
}

func (a *App) HTTPServer() *httpserver.HTTPServer {
	return a.HTTP
}

func (a *App) resolveConfig(o options) (config.Config, error) {
	if o.config != nil {
		if err := o.config.Validate(); err != nil {
			return config.Config{}, InitError{Step: "config.validate", Cause: err}
		}
		return *o.config, nil
	}

	if o.dotenvPath != "" {
		cfg, err := config.Load(config.WithDotenvPath(o.dotenvPath))
		if err != nil {
			return config.Config{}, InitError{Step: "config.load", Cause: err}
		}
		return cfg, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return config.Config{}, InitError{Step: "config.load", Cause: err}
	}
	return cfg, nil
}

func (a *App) resolveLogger(cfg config.Config, o options) (*zap.Logger, error) {
	if o.logger != nil {
		return o.logger, nil
	}
	logger, err := logging.New(cfg.Log.AsLoggingConfig())
	if err != nil {
		return nil, InitError{Step: "logger.init", Cause: err}
	}
	return logger, nil
}

func (a *App) logStartup(ctx context.Context) {
	if a.Logger == nil {
		return
	}
	_ = ctx

	l := a.Logger.With(
		zap.String("module", "bubble"),
		zap.String("go_version", runtime.Version()),
		zap.Time("started_at", a.StartedAt),
	)

	l.Info(
		"app initialized",
		zap.String("log_level", a.Config.Log.Level.String()),
		zap.Bool("log_development", a.Config.Log.Development),
		zap.String("http_address", a.Config.HTTP.Address),
		zap.String("db_driver", a.Config.Database.Driver),
		zap.Int("db_max_open_conns", a.Config.Database.MaxOpenConns),
		zap.Int("db_max_idle_conns", a.Config.Database.MaxIdleConns),
		zap.Duration("db_conn_max_lifetime", a.Config.Database.ConnMaxLifetime),
		zap.Bool("db_auto_migrate", a.Config.Database.AutoMigrate),
		zap.Int("db_migrations_strategy", int(a.Config.Database.Migrations.Strategy)),
		zap.String("db_migrations_dir", a.Config.Database.Migrations.Dir),
		zap.Bool("db_migrations_lock_enabled", a.Config.Database.Migrations.Lock.Enabled),
		zap.Int("cache_provider", int(a.Config.Cache.Provider)),
		zap.Bool("cron_enabled", a.Config.Cron.Enabled),
		zap.Int("queue_provider", int(a.Config.Queue.Provider)),
		zap.String("s3_endpoint", a.Config.S3.Endpoint),
		zap.String("s3_region", a.Config.S3.Region),
		zap.String("s3_bucket", a.Config.S3.Bucket),
	)
}
