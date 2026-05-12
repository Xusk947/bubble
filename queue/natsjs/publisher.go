package natsjs

import (
	"context"
	"strings"

	"bubble/config"
	"bubble/queue"

	"github.com/nats-io/nats.go"
)

type Publisher[T any] struct {
	js      nats.JetStreamContext
	subject string
	codec   queue.Codec[T]
}

func NewPublisher[T any](js nats.JetStreamContext, cfg config.NATSConfig, codec queue.Codec[T]) (*Publisher[T], error) {
	subject := strings.TrimSpace(cfg.Subject)
	if subject == "" {
		return nil, queue.InvalidConfigError{Field: "nats.subject", Message: "empty"}
	}
	if codec == nil {
		return nil, queue.InvalidConfigError{Field: "codec", Message: "nil"}
	}
	return &Publisher[T]{js: js, subject: subject, codec: codec}, nil
}

func (p *Publisher[T]) Publish(ctx context.Context, msg T, opts queue.PublishOptions) error {
	data, err := p.codec.Marshal(msg)
	if err != nil {
		return queue.OperationError{Op: "marshal", Cause: err}
	}

	m := nats.NewMsg(p.subject)
	m.Data = data
	if ct := strings.TrimSpace(p.codec.ContentType()); ct != "" {
		m.Header.Set("Content-Type", ct)
	}
	if strings.TrimSpace(opts.Key) != "" {
		m.Header.Set("X-Key", opts.Key)
	}
	for _, h := range opts.Headers {
		if strings.TrimSpace(h.Name) == "" {
			continue
		}
		m.Header.Add(h.Name, h.Value)
	}

	if _, err := p.js.PublishMsg(m, nats.Context(ctx)); err != nil {
		return queue.OperationError{Op: "publish", Cause: err}
	}
	return nil
}

var _ queue.Publisher[struct{}] = (*Publisher[struct{}])(nil)

