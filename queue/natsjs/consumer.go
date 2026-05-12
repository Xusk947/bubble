package natsjs

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/queue"

	"github.com/nats-io/nats.go"
)

type Consumer[T any] struct {
	js      nats.JetStreamContext
	subject string
	durable string
	ackWait time.Duration
	codec   queue.Codec[T]

	mu  sync.Mutex
	sub *nats.Subscription
}

func NewConsumer[T any](js nats.JetStreamContext, cfg config.NATSConfig, codec queue.Codec[T]) (*Consumer[T], error) {
	subject := strings.TrimSpace(cfg.Subject)
	if subject == "" {
		return nil, queue.InvalidConfigError{Field: "nats.subject", Message: "empty"}
	}
	durable := strings.TrimSpace(cfg.Durable)
	if durable == "" {
		return nil, queue.InvalidConfigError{Field: "nats.durable", Message: "empty"}
	}
	if codec == nil {
		return nil, queue.InvalidConfigError{Field: "codec", Message: "nil"}
	}
	ackWait := cfg.AckWait
	if ackWait <= 0 {
		ackWait = 30 * time.Second
	}
	return &Consumer[T]{
		js:      js,
		subject: subject,
		durable: durable,
		ackWait: ackWait,
		codec:   codec,
	}, nil
}

func (c *Consumer[T]) Start(ctx context.Context, handler queue.Handler[T]) error {
	if handler == nil {
		return queue.InvalidConfigError{Field: "handler", Message: "nil"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sub != nil {
		return queue.OperationError{Op: "start", Cause: errors.New("already started")}
	}

	sub, err := c.js.Subscribe(
		c.subject,
		func(m *nats.Msg) {
			c.handleMsg(ctx, m, handler)
		},
		nats.Durable(c.durable),
		nats.ManualAck(),
		nats.AckWait(c.ackWait),
	)
	if err != nil {
		return queue.OperationError{Op: "subscribe", Cause: err}
	}
	c.sub = sub
	return nil
}

func (c *Consumer[T]) Stop(ctx context.Context) error {
	_ = ctx
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sub == nil {
		return nil
	}
	err := c.sub.Unsubscribe()
	c.sub = nil
	if err != nil {
		return queue.OperationError{Op: "unsubscribe", Cause: err}
	}
	return nil
}

func (c *Consumer[T]) handleMsg(ctx context.Context, m *nats.Msg, handler queue.Handler[T]) {
	value, err := c.codec.Unmarshal(m.Data)
	if err != nil {
		_ = m.Term()
		return
	}
	if err := handler(ctx, value); err != nil {
		_ = m.Nak()
		return
	}
	_ = m.Ack()
}

var _ queue.Consumer[struct{}] = (*Consumer[struct{}])(nil)
