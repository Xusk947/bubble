package kafka

import (
	"context"
	"strings"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/queue"

	"github.com/segmentio/kafka-go"
)

type Publisher[T any] struct {
	writer *kafka.Writer
	codec  queue.Codec[T]
}

func NewPublisher[T any](client *Client, cfg config.KafkaConfig, codec queue.Codec[T]) (*Publisher[T], error) {
	if client == nil || len(client.Brokers) == 0 {
		return nil, queue.InvalidConfigError{Field: "kafka.brokers", Message: "empty"}
	}
	topic := strings.TrimSpace(cfg.Topic)
	if topic == "" {
		return nil, queue.InvalidConfigError{Field: "kafka.topic", Message: "empty"}
	}
	if codec == nil {
		return nil, queue.InvalidConfigError{Field: "codec", Message: "nil"}
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(client.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Publisher[T]{writer: w, codec: codec}, nil
}

func (p *Publisher[T]) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func (p *Publisher[T]) Publish(ctx context.Context, msg T, opts queue.PublishOptions) error {
	data, err := p.codec.Marshal(msg)
	if err != nil {
		return queue.OperationError{Op: "marshal", Cause: err}
	}

	headers := make([]kafka.Header, 0, len(opts.Headers)+1)
	if ct := strings.TrimSpace(p.codec.ContentType()); ct != "" {
		headers = append(headers, kafka.Header{Key: "Content-Type", Value: []byte(ct)})
	}
	for _, h := range opts.Headers {
		if strings.TrimSpace(h.Name) == "" {
			continue
		}
		headers = append(headers, kafka.Header{Key: h.Name, Value: []byte(h.Value)})
	}

	m := kafka.Message{
		Time:    time.Now().UTC(),
		Value:   data,
		Headers: headers,
	}
	if strings.TrimSpace(opts.Key) != "" {
		m.Key = []byte(opts.Key)
	}

	if err := p.writer.WriteMessages(ctx, m); err != nil {
		return queue.OperationError{Op: "publish", Cause: err}
	}
	return nil
}

var _ queue.Publisher[struct{}] = (*Publisher[struct{}])(nil)
