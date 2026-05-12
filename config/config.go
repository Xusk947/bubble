package config

import (
	"time"

	"github.com/Xusk947/bubble/logging"
)

type Config struct {
	Log      LogConfig
	Database DatabaseConfig
	Cache    CacheConfig
	S3       S3Config
	HTTP     HTTPConfig
	Cron     CronConfig
	Queue    QueueConfig
}

type LogConfig struct {
	Level       logging.Level
	Development bool
	Encoding    logging.Encoding
	Output      logging.OutputConfig
}

type DatabaseConfig struct {
	Driver          string
	DSN             string
	AtlasURL        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	AutoMigrate     bool
	Migrations      MigrationsConfig
}

type MigrationStrategy int

const (
	MigrationStrategyUnspecified MigrationStrategy = iota
	MigrationStrategyAtlas
)

const defaultMigrationLockKey int64 = 424242

type MigrationsConfig struct {
	Strategy MigrationStrategy
	Dir      string
	Lock     MigrationLockConfig
}

type MigrationLockConfig struct {
	Enabled bool
	Key     int64
}

type CacheProvider int

const (
	CacheProviderUnspecified CacheProvider = iota
	CacheProviderLocal
	CacheProviderRedis
)

type CacheConfig struct {
	Provider CacheProvider
	Local    LocalCacheConfig
	Redis    RedisConfig
}

type LocalCacheConfig struct {
	MaxEntries int
	DefaultTTL time.Duration
}

type RedisConfig struct {
	Address  string
	Username string
	Password string
	DB       int
}

type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	UsePathStyle    bool
	DisableTLS      bool
}

type HTTPConfig struct {
	Address      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	EnablePprof  bool
	EnableCORS   bool
}

type CronConfig struct {
	Enabled bool
}

type QueueProvider int

const (
	QueueProviderUnspecified QueueProvider = iota
	QueueProviderNATS
	QueueProviderKafka
)

type QueueConfig struct {
	Provider QueueProvider
	NATS     NATSConfig
	Kafka    KafkaConfig
}

type NATSConfig struct {
	URL        string
	Creds      string
	Stream     string
	Subject    string
	Durable    string
	AckWait    time.Duration
	MaxDeliver int
	Ensure     bool
}

type KafkaConfig struct {
	Brokers string
	Topic   string
	GroupID string
}

func Default() Config {
	return Config{
		Log: LogConfig{
			Level:       logging.LevelInfo,
			Development: false,
			Encoding:    logging.EncodingJSON,
			Output: logging.OutputConfig{
				Stdout: true,
				Stderr: false,
			},
		},
		Database: DatabaseConfig{
			Driver:          "",
			DSN:             "",
			AtlasURL:        "",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 30 * time.Minute,
			AutoMigrate:     false,
			Migrations: MigrationsConfig{
				Strategy: MigrationStrategyAtlas,
				Dir:      "",
				Lock: MigrationLockConfig{
					Enabled: true,
					Key:     defaultMigrationLockKey,
				},
			},
		},
		Cache: CacheConfig{
			Provider: CacheProviderLocal,
			Local: LocalCacheConfig{
				MaxEntries: 1024,
				DefaultTTL: 5 * time.Minute,
			},
			Redis: RedisConfig{
				Address:  "127.0.0.1:6379",
				Username: "",
				Password: "",
				DB:       0,
			},
		},
		S3: S3Config{
			Endpoint:        "",
			Region:          "us-east-1",
			Bucket:          "",
			AccessKeyID:     "",
			SecretAccessKey: "",
			SessionToken:    "",
			UsePathStyle:    false,
			DisableTLS:      false,
		},
		HTTP: HTTPConfig{
			Address:      ":8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
			EnablePprof:  false,
			EnableCORS:   false,
		},
		Cron: CronConfig{
			Enabled: true,
		},
		Queue: QueueConfig{
			Provider: QueueProviderUnspecified,
			NATS: NATSConfig{
				URL:        "nats://127.0.0.1:4222",
				Creds:      "",
				Stream:     "",
				Subject:    "",
				Durable:    "",
				AckWait:    30 * time.Second,
				MaxDeliver: 10,
				Ensure:     true,
			},
			Kafka: KafkaConfig{
				Brokers: "127.0.0.1:9092",
				Topic:   "",
				GroupID: "",
			},
		},
	}
}
