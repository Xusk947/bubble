package kafka

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/queue"

	"github.com/segmentio/kafka-go"
)

type Consumer[T any] struct {
	reader *kafka.Reader
	codec  queue.Codec[T]

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewConsumer[T any](client *Client, cfg config.KafkaConfig, codec queue.Codec[T]) (*Consumer[T], error) {
	if client == nil || len(client.Brokers) == 0 {
		return nil, queue.InvalidConfigError{Field: "kafka.brokers", Message: "empty"}
	}
	topic := strings.TrimSpace(cfg.Topic)
	if topic == "" {
		return nil, queue.InvalidConfigError{Field: "kafka.topic", Message: "empty"}
	}
	group := strings.TrimSpace(cfg.GroupID)
	if group == "" {
		return nil, queue.InvalidConfigError{Field: "kafka.group_id", Message: "empty"}
	}
	if codec == nil {
		return nil, queue.InvalidConfigError{Field: "codec", Message: "nil"}
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  client.Brokers,
		Topic:    topic,
		GroupID:  group,
		MinBytes: 1,
		MaxBytes: 10e6,
		MaxWait:  2 * time.Second,
	})
	return &Consumer[T]{reader: r, codec: codec}, nil
}

func (c *Consumer[T]) Start(ctx context.Context, handler queue.Handler[T]) error {
	if handler == nil {
		return queue.InvalidConfigError{Field: "handler", Message: "nil"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return queue.OperationError{Op: "start", Cause: errors.New("already started")}
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.loop(runCtx, handler)
	}()
	return nil
}

func (c *Consumer[T]) Stop(ctx context.Context) error {
	c.mu.Lock()
	cancel := c.cancel
	c.cancel = nil
	c.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return queue.OperationError{Op: "stop", Cause: ctx.Err()}
	case <-done:
	}

	if c.reader != nil {
		if err := c.reader.Close(); err != nil {
			return queue.OperationError{Op: "close", Cause: err}
		}
	}
	return nil
}

func (c *Consumer[T]) loop(ctx context.Context, handler queue.Handler[T]) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			time.Sleep(250 * time.Millisecond)
			continue
		}

		value, err := c.codec.Unmarshal(m.Value)
		if err != nil {
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if err := handler(ctx, value); err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		_ = c.reader.CommitMessages(ctx, m)
	}
}

var _ queue.Consumer[struct{}] = (*Consumer[struct{}])(nil)
