package natsjs

import (
	"context"
	"strings"

	"bubble/config"
	"bubble/queue"

	"github.com/nats-io/nats.go"
)

func EnsureStream(ctx context.Context, js nats.JetStreamContext, cfg config.NATSConfig) error {
	if !cfg.Ensure {
		return nil
	}
	stream := strings.TrimSpace(cfg.Stream)
	subject := strings.TrimSpace(cfg.Subject)
	if stream == "" {
		return queue.InvalidConfigError{Field: "nats.stream", Message: "empty"}
	}
	if subject == "" {
		return queue.InvalidConfigError{Field: "nats.subject", Message: "empty"}
	}

	if _, err := js.StreamInfo(stream); err == nil {
		return nil
	}

	_, err := js.AddStream(&nats.StreamConfig{
		Name:     stream,
		Subjects: []string{subject},
	})
	return err
}

