package app

import (
	"context"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/health"
)

const (
	checkAppInitialized = "app.initialized"
	checkAppStarted     = "app.started"
	checkDatabase       = "database"
	checkNATS           = "nats"
	checkKafka          = "kafka"
	checkS3             = "s3"
)

func (a *App) Health(ctx context.Context) health.Status {
	status := health.Status{
		Live:  a != nil && a.Initialized,
		Ready: a != nil && a.Initialized && a.Started,
	}

	checks := make([]health.Check, 0, 6)
	checks = append(checks, health.Check{Name: checkAppInitialized, Healthy: a != nil && a.Initialized})
	checks = append(checks, health.Check{Name: checkAppStarted, Healthy: a != nil && a.Started})

	if a == nil {
		status.Checks = checks
		status.Ready = false
		return status
	}

	if a.Config.Database.Driver != "" && a.DB() != nil {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err := a.DB().PingContext(pingCtx)
		checks = append(checks, health.Check{Name: checkDatabase, Healthy: err == nil, Message: errorMessage(err)})
		if err != nil {
			status.Ready = false
		}
	}

	if stringsTrim(a.Config.S3.Bucket) != "" {
		healthy := a.ObjectStore != nil
		checks = append(checks, health.Check{Name: checkS3, Healthy: healthy})
		if !healthy {
			status.Ready = false
		}
	}

	switch a.Config.Queue.Provider {
	case config.QueueProviderNATS:
		healthy := a.NATS != nil && a.NATS.Conn != nil && a.NATS.Conn.IsConnected()
		checks = append(checks, health.Check{Name: checkNATS, Healthy: healthy})
		if !healthy {
			status.Ready = false
		}
	case config.QueueProviderKafka:
		if a.Kafka != nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			err := a.Kafka.Ping(pingCtx)
			checks = append(checks, health.Check{Name: checkKafka, Healthy: err == nil, Message: errorMessage(err)})
			if err != nil {
				status.Ready = false
			}
		} else {
			checks = append(checks, health.Check{Name: checkKafka, Healthy: false})
			status.Ready = false
		}
	default:
	}

	status.Checks = checks
	return status
}

func stringsTrim(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\n' || value[0] == '\t' || value[0] == '\r') {
		value = value[1:]
	}
	for len(value) > 0 {
		last := value[len(value)-1]
		if last != ' ' && last != '\n' && last != '\t' && last != '\r' {
			break
		}
		value = value[:len(value)-1]
	}
	return value
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
