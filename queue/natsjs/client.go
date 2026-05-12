package natsjs

import (
	"context"
	"strings"
	"time"

	"github.com/Xusk947/bubble/config"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Client struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
}

func Connect(ctx context.Context, cfg config.NATSConfig, logger *zap.Logger) (*Client, error) {
	url := strings.TrimSpace(cfg.URL)
	if url == "" {
		return nil, nil
	}

	opts := make([]nats.Option, 0, 3)
	opts = append(opts, nats.Timeout(5*time.Second))
	opts = append(opts, nats.Name("bubble"))

	if strings.TrimSpace(cfg.Creds) != "" {
		opts = append(opts, nats.UserCredentials(cfg.Creds))
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	if err := EnsureStream(ctx, js, cfg); err != nil {
		nc.Close()
		return nil, err
	}

	if logger != nil {
		logger.Info("nats connected", zap.String("nats_url", url))
	}

	return &Client{Conn: nc, JS: js}, nil
}

func (c *Client) Close() {
	if c == nil || c.Conn == nil {
		return
	}
	c.Conn.Close()
}
