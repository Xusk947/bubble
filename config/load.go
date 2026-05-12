package config

import (
	"os"
	"strings"

	"bubble/logging"

	"github.com/joho/godotenv"
)

const (
	envDotenvPath EnvVar = "BUBBLE_DOTENV"

	envLogLevel       EnvVar = "BUBBLE_LOG_LEVEL"
	envLogDevelopment EnvVar = "BUBBLE_LOG_DEVELOPMENT"
	envLogEncoding    EnvVar = "BUBBLE_LOG_ENCODING"
	envLogStdout      EnvVar = "BUBBLE_LOG_STDOUT"
	envLogStderr      EnvVar = "BUBBLE_LOG_STDERR"

	envDatabaseDriver      EnvVar = "BUBBLE_DB_DRIVER"
	envDatabaseDSN         EnvVar = "BUBBLE_DB_DSN"
	envDatabaseAtlasURL    EnvVar = "BUBBLE_DB_ATLAS_URL"
	envDatabaseMaxOpen     EnvVar = "BUBBLE_DB_MAX_OPEN_CONNS"
	envDatabaseMaxIdle     EnvVar = "BUBBLE_DB_MAX_IDLE_CONNS"
	envDatabaseMaxLifetime EnvVar = "BUBBLE_DB_CONN_MAX_LIFETIME"
	envDatabaseAutoMigrate EnvVar = "BUBBLE_DB_AUTO_MIGRATE"
	envDatabaseMigStrategy EnvVar = "BUBBLE_DB_MIGRATIONS_STRATEGY"
	envDatabaseMigDir      EnvVar = "BUBBLE_DB_MIGRATIONS_DIR"
	envDatabaseMigLock     EnvVar = "BUBBLE_DB_MIGRATIONS_LOCK_ENABLED"
	envDatabaseMigLockKey  EnvVar = "BUBBLE_DB_MIGRATIONS_LOCK_KEY"

	envCacheProvider EnvVar = "BUBBLE_CACHE_PROVIDER"
	envRedisAddress  EnvVar = "BUBBLE_REDIS_ADDRESS"
	envRedisUsername EnvVar = "BUBBLE_REDIS_USERNAME"
	envRedisPassword EnvVar = "BUBBLE_REDIS_PASSWORD"
	envRedisDB       EnvVar = "BUBBLE_REDIS_DB"
	envLocalMax      EnvVar = "BUBBLE_LOCAL_CACHE_MAX_ENTRIES"
	envLocalTTL      EnvVar = "BUBBLE_LOCAL_CACHE_DEFAULT_TTL"

	envS3Endpoint        EnvVar = "BUBBLE_S3_ENDPOINT"
	envS3Region          EnvVar = "BUBBLE_S3_REGION"
	envS3Bucket          EnvVar = "BUBBLE_S3_BUCKET"
	envS3AccessKeyID     EnvVar = "BUBBLE_S3_ACCESS_KEY_ID"
	envS3SecretAccessKey EnvVar = "BUBBLE_S3_SECRET_ACCESS_KEY"
	envS3SessionToken    EnvVar = "BUBBLE_S3_SESSION_TOKEN"
	envS3UsePathStyle    EnvVar = "BUBBLE_S3_USE_PATH_STYLE"
	envS3DisableTLS      EnvVar = "BUBBLE_S3_DISABLE_TLS"

	envHTTPAddress      EnvVar = "BUBBLE_HTTP_ADDRESS"
	envHTTPReadTimeout  EnvVar = "BUBBLE_HTTP_READ_TIMEOUT"
	envHTTPWriteTimeout EnvVar = "BUBBLE_HTTP_WRITE_TIMEOUT"
	envHTTPIdleTimeout  EnvVar = "BUBBLE_HTTP_IDLE_TIMEOUT"
	envHTTPEnablePprof  EnvVar = "BUBBLE_HTTP_ENABLE_PPROF"
	envHTTPEnableCORS   EnvVar = "BUBBLE_HTTP_ENABLE_CORS"

	envCronEnabled EnvVar = "BUBBLE_CRON_ENABLED"

	envQueueProvider  EnvVar = "BUBBLE_QUEUE_PROVIDER"
	envNATSURL        EnvVar = "BUBBLE_NATS_URL"
	envNATSCreds      EnvVar = "BUBBLE_NATS_CREDS"
	envNATSStream     EnvVar = "BUBBLE_NATS_STREAM"
	envNATSSubject    EnvVar = "BUBBLE_NATS_SUBJECT"
	envNATSDurable    EnvVar = "BUBBLE_NATS_DURABLE"
	envNATSAckWait    EnvVar = "BUBBLE_NATS_ACK_WAIT"
	envNATSMaxDeliver EnvVar = "BUBBLE_NATS_MAX_DELIVER"
	envNATSEnsure     EnvVar = "BUBBLE_NATS_ENSURE"
	envKafkaBrokers   EnvVar = "BUBBLE_KAFKA_BROKERS"
	envKafkaTopic     EnvVar = "BUBBLE_KAFKA_TOPIC"
	envKafkaGroupID   EnvVar = "BUBBLE_KAFKA_GROUP_ID"
)

type LoadOption func(*loadOptions)

type loadOptions struct {
	dotenvPath string
	overrides  []func(*Config)
}

func WithDotenvPath(path string) LoadOption {
	return func(o *loadOptions) {
		o.dotenvPath = path
	}
}

func WithLogConfig(value LogConfig) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.Log = value
		})
	}
}

func WithHTTPConfig(value HTTPConfig) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.HTTP = value
		})
	}
}

func WithDatabaseConfig(value DatabaseConfig) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.Database = value
		})
	}
}

func WithCacheConfig(value CacheConfig) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.Cache = value
		})
	}
}

func WithS3Config(value S3Config) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.S3 = value
		})
	}
}

func WithCronConfig(value CronConfig) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.Cron = value
		})
	}
}

func WithQueueConfig(value QueueConfig) LoadOption {
	return func(o *loadOptions) {
		o.overrides = append(o.overrides, func(c *Config) {
			c.Queue = value
		})
	}
}

func Load(opts ...LoadOption) (Config, error) {
	lo := loadOptions{}
	for _, opt := range opts {
		opt(&lo)
	}

	dotenvPath := strings.TrimSpace(lo.dotenvPath)
	if dotenvPath == "" {
		dotenvPath = strings.TrimSpace(envString(envDotenvPath, ""))
	}
	if dotenvPath != "" {
		if err := godotenv.Overload(dotenvPath); err != nil {
			return Config{}, DotenvLoadError{Path: dotenvPath, Cause: err}
		}
	}

	cfg, err := loadFromEnv()
	if err != nil {
		return Config{}, err
	}
	for _, apply := range lo.overrides {
		apply(&cfg)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadFromEnv() (Config, error) {
	cfg := Default()

	level, err := logging.ParseLevel(envString(envLogLevel, cfg.Log.Level.String()))
	if err != nil {
		return Config{}, InvalidEnvValueError{Name: envLogLevel, Value: os.Getenv(string(envLogLevel)), Expected: "debug|info|warn|error|dpanic|panic|fatal"}
	}
	cfg.Log.Level = level

	development, err := envBool(envLogDevelopment, cfg.Log.Development)
	if err != nil {
		return Config{}, err
	}
	cfg.Log.Development = development

	encodingRaw := strings.TrimSpace(strings.ToLower(envString(envLogEncoding, "")))
	switch encodingRaw {
	case "":
	case "json":
		cfg.Log.Encoding = logging.EncodingJSON
	case "console":
		cfg.Log.Encoding = logging.EncodingConsole
	default:
		return Config{}, InvalidEnvValueError{Name: envLogEncoding, Value: os.Getenv(string(envLogEncoding)), Expected: "json|console"}
	}

	stdout, err := envBool(envLogStdout, cfg.Log.Output.Stdout)
	if err != nil {
		return Config{}, err
	}
	stderr, err := envBool(envLogStderr, cfg.Log.Output.Stderr)
	if err != nil {
		return Config{}, err
	}
	cfg.Log.Output = logging.OutputConfig{Stdout: stdout, Stderr: stderr}

	cfg.Database.Driver = strings.TrimSpace(envString(envDatabaseDriver, cfg.Database.Driver))
	cfg.Database.DSN = strings.TrimSpace(envString(envDatabaseDSN, cfg.Database.DSN))
	cfg.Database.AtlasURL = strings.TrimSpace(envString(envDatabaseAtlasURL, cfg.Database.AtlasURL))

	maxOpen, err := envInt(envDatabaseMaxOpen, cfg.Database.MaxOpenConns)
	if err != nil {
		return Config{}, err
	}
	maxIdle, err := envInt(envDatabaseMaxIdle, cfg.Database.MaxIdleConns)
	if err != nil {
		return Config{}, err
	}
	maxLifetime, err := envDuration(envDatabaseMaxLifetime, cfg.Database.ConnMaxLifetime)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.MaxOpenConns = maxOpen
	cfg.Database.MaxIdleConns = maxIdle
	cfg.Database.ConnMaxLifetime = maxLifetime

	autoMigrate, err := envBool(envDatabaseAutoMigrate, cfg.Database.AutoMigrate)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.AutoMigrate = autoMigrate

	migStrategyRaw := strings.TrimSpace(strings.ToLower(envString(envDatabaseMigStrategy, "")))
	switch migStrategyRaw {
	case "":
	case "atlas":
		cfg.Database.Migrations.Strategy = MigrationStrategyAtlas
	default:
		return Config{}, InvalidEnvValueError{Name: envDatabaseMigStrategy, Value: os.Getenv(string(envDatabaseMigStrategy)), Expected: "atlas"}
	}

	cfg.Database.Migrations.Dir = strings.TrimSpace(envString(envDatabaseMigDir, cfg.Database.Migrations.Dir))

	lockEnabled, err := envBool(envDatabaseMigLock, cfg.Database.Migrations.Lock.Enabled)
	if err != nil {
		return Config{}, err
	}
	lockKey, err := envInt64(envDatabaseMigLockKey, cfg.Database.Migrations.Lock.Key)
	if err != nil {
		return Config{}, err
	}
	cfg.Database.Migrations.Lock = MigrationLockConfig{
		Enabled: lockEnabled,
		Key:     lockKey,
	}

	cacheProviderRaw := strings.TrimSpace(strings.ToLower(envString(envCacheProvider, "")))
	switch cacheProviderRaw {
	case "":
	case "local":
		cfg.Cache.Provider = CacheProviderLocal
	case "redis":
		cfg.Cache.Provider = CacheProviderRedis
	default:
		return Config{}, InvalidEnvValueError{Name: envCacheProvider, Value: os.Getenv(string(envCacheProvider)), Expected: "local|redis"}
	}

	cfg.Cache.Redis.Address = strings.TrimSpace(envString(envRedisAddress, cfg.Cache.Redis.Address))
	cfg.Cache.Redis.Username = strings.TrimSpace(envString(envRedisUsername, cfg.Cache.Redis.Username))
	cfg.Cache.Redis.Password = strings.TrimSpace(envString(envRedisPassword, cfg.Cache.Redis.Password))
	redisDB, err := envInt(envRedisDB, cfg.Cache.Redis.DB)
	if err != nil {
		return Config{}, err
	}
	cfg.Cache.Redis.DB = redisDB

	localMaxEntries, err := envInt(envLocalMax, cfg.Cache.Local.MaxEntries)
	if err != nil {
		return Config{}, err
	}
	localTTL, err := envDuration(envLocalTTL, cfg.Cache.Local.DefaultTTL)
	if err != nil {
		return Config{}, err
	}
	cfg.Cache.Local.MaxEntries = localMaxEntries
	cfg.Cache.Local.DefaultTTL = localTTL

	cfg.S3.Endpoint = strings.TrimSpace(envString(envS3Endpoint, cfg.S3.Endpoint))
	cfg.S3.Region = strings.TrimSpace(envString(envS3Region, cfg.S3.Region))
	cfg.S3.Bucket = strings.TrimSpace(envString(envS3Bucket, cfg.S3.Bucket))
	cfg.S3.AccessKeyID = strings.TrimSpace(envString(envS3AccessKeyID, cfg.S3.AccessKeyID))
	cfg.S3.SecretAccessKey = strings.TrimSpace(envString(envS3SecretAccessKey, cfg.S3.SecretAccessKey))
	cfg.S3.SessionToken = strings.TrimSpace(envString(envS3SessionToken, cfg.S3.SessionToken))
	usePathStyle, err := envBool(envS3UsePathStyle, cfg.S3.UsePathStyle)
	if err != nil {
		return Config{}, err
	}
	disableTLS, err := envBool(envS3DisableTLS, cfg.S3.DisableTLS)
	if err != nil {
		return Config{}, err
	}
	cfg.S3.UsePathStyle = usePathStyle
	cfg.S3.DisableTLS = disableTLS

	cfg.HTTP.Address = strings.TrimSpace(envString(envHTTPAddress, cfg.HTTP.Address))
	readTimeout, err := envDuration(envHTTPReadTimeout, cfg.HTTP.ReadTimeout)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := envDuration(envHTTPWriteTimeout, cfg.HTTP.WriteTimeout)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := envDuration(envHTTPIdleTimeout, cfg.HTTP.IdleTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.ReadTimeout = readTimeout
	cfg.HTTP.WriteTimeout = writeTimeout
	cfg.HTTP.IdleTimeout = idleTimeout

	enablePprof, err := envBool(envHTTPEnablePprof, cfg.HTTP.EnablePprof)
	if err != nil {
		return Config{}, err
	}
	enableCORS, err := envBool(envHTTPEnableCORS, cfg.HTTP.EnableCORS)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTP.EnablePprof = enablePprof
	cfg.HTTP.EnableCORS = enableCORS

	cronEnabled, err := envBool(envCronEnabled, cfg.Cron.Enabled)
	if err != nil {
		return Config{}, err
	}
	cfg.Cron.Enabled = cronEnabled

	queueProviderRaw := strings.TrimSpace(strings.ToLower(envString(envQueueProvider, "")))
	switch queueProviderRaw {
	case "":
	case "nats":
		cfg.Queue.Provider = QueueProviderNATS
	case "kafka":
		cfg.Queue.Provider = QueueProviderKafka
	default:
		return Config{}, InvalidEnvValueError{Name: envQueueProvider, Value: os.Getenv(string(envQueueProvider)), Expected: "nats|kafka"}
	}

	cfg.Queue.NATS.URL = strings.TrimSpace(envString(envNATSURL, cfg.Queue.NATS.URL))
	cfg.Queue.NATS.Creds = strings.TrimSpace(envString(envNATSCreds, cfg.Queue.NATS.Creds))
	cfg.Queue.NATS.Stream = strings.TrimSpace(envString(envNATSStream, cfg.Queue.NATS.Stream))
	cfg.Queue.NATS.Subject = strings.TrimSpace(envString(envNATSSubject, cfg.Queue.NATS.Subject))
	cfg.Queue.NATS.Durable = strings.TrimSpace(envString(envNATSDurable, cfg.Queue.NATS.Durable))
	ackWait, err := envDuration(envNATSAckWait, cfg.Queue.NATS.AckWait)
	if err != nil {
		return Config{}, err
	}
	maxDeliver, err := envInt(envNATSMaxDeliver, cfg.Queue.NATS.MaxDeliver)
	if err != nil {
		return Config{}, err
	}
	ensure, err := envBool(envNATSEnsure, cfg.Queue.NATS.Ensure)
	if err != nil {
		return Config{}, err
	}
	cfg.Queue.NATS.AckWait = ackWait
	cfg.Queue.NATS.MaxDeliver = maxDeliver
	cfg.Queue.NATS.Ensure = ensure

	cfg.Queue.Kafka.Brokers = strings.TrimSpace(envString(envKafkaBrokers, cfg.Queue.Kafka.Brokers))
	cfg.Queue.Kafka.Topic = strings.TrimSpace(envString(envKafkaTopic, cfg.Queue.Kafka.Topic))
	cfg.Queue.Kafka.GroupID = strings.TrimSpace(envString(envKafkaGroupID, cfg.Queue.Kafka.GroupID))

	return cfg, nil
}
