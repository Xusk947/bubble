package config

import "strings"

func looksLikeURL(value string) bool {
	return strings.Contains(value, "://")
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.HTTP.Address) == "" {
		return ValidationError{Field: FieldHTTP, Message: "http address is empty"}
	}
	if c.HTTP.ReadTimeout < 0 {
		return ValidationError{Field: FieldHTTP, Message: "http read timeout is negative"}
	}
	if c.HTTP.WriteTimeout < 0 {
		return ValidationError{Field: FieldHTTP, Message: "http write timeout is negative"}
	}
	if c.HTTP.IdleTimeout < 0 {
		return ValidationError{Field: FieldHTTP, Message: "http idle timeout is negative"}
	}

	switch c.Cache.Provider {
	case CacheProviderLocal:
	case CacheProviderRedis:
		if strings.TrimSpace(c.Cache.Redis.Address) == "" {
			return ValidationError{Field: FieldCache, Message: "redis address is empty"}
		}
	default:
		return ValidationError{Field: FieldCache, Message: "cache provider is unspecified"}
	}

	if c.Database.Driver != "" && strings.TrimSpace(c.Database.DSN) == "" {
		return ValidationError{Field: FieldDatabase, Message: "database dsn is empty"}
	}

	if c.Database.MaxOpenConns < 0 {
		return ValidationError{Field: FieldDatabase, Message: "max open conns is negative"}
	}
	if c.Database.MaxIdleConns < 0 {
		return ValidationError{Field: FieldDatabase, Message: "max idle conns is negative"}
	}
	if c.Database.AutoMigrate {
		switch c.Database.Migrations.Strategy {
		case MigrationStrategyAtlas:
			if strings.TrimSpace(c.Database.Migrations.Dir) == "" {
				return ValidationError{Field: FieldDatabase, Message: "migrations dir is empty"}
			}
			atlasURL := strings.TrimSpace(c.Database.AtlasURL)
			if atlasURL == "" {
				if !looksLikeURL(c.Database.DSN) {
					return ValidationError{Field: FieldDatabase, Message: "atlas url is empty"}
				}
			}
		default:
			return ValidationError{Field: FieldDatabase, Message: "migrations strategy is unspecified"}
		}
	}

	switch c.Queue.Provider {
	case QueueProviderUnspecified:
	case QueueProviderNATS:
		if strings.TrimSpace(c.Queue.NATS.URL) == "" {
			return ValidationError{Field: FieldQueue, Message: "nats url is empty"}
		}
		if c.Queue.NATS.AckWait < 0 {
			return ValidationError{Field: FieldQueue, Message: "nats ack wait is negative"}
		}
		if c.Queue.NATS.MaxDeliver < 0 {
			return ValidationError{Field: FieldQueue, Message: "nats max deliver is negative"}
		}
	case QueueProviderKafka:
		if strings.TrimSpace(c.Queue.Kafka.Brokers) == "" {
			return ValidationError{Field: FieldQueue, Message: "kafka brokers is empty"}
		}
	default:
		return ValidationError{Field: FieldQueue, Message: "queue provider is unspecified"}
	}

	return nil
}
