package kafka

import (
	"context"
	"strings"
	"time"

	"github.com/Xusk947/bubble/config"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Client struct {
	Brokers []string
	Dialer  *kafka.Dialer
}

func NewClient(cfg config.KafkaConfig, logger *zap.Logger) (*Client, error) {
	brokers := splitCSV(cfg.Brokers)
	if len(brokers) == 0 {
		return nil, nil
	}
	dialer := &kafka.Dialer{
		Timeout: 5 * time.Second,
	}
	if logger != nil {
		logger.Info("kafka configured", zap.Int("kafka_brokers", len(brokers)))
	}
	return &Client{Brokers: brokers, Dialer: dialer}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || len(c.Brokers) == 0 {
		return nil
	}
	conn, err := kafka.DialContext(ctx, "tcp", c.Brokers[0])
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func splitCSV(value string) []string {
	raw := strings.Split(value, ",")
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		v := strings.TrimSpace(r)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}
