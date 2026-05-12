package natsjs

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/queue"
)

type natsPayload struct {
	Value string
}

func TestNATSJetStream_PublishConsume(t *testing.T) {
	url := strings.TrimSpace(os.Getenv("NATS_URL"))
	stream := strings.TrimSpace(os.Getenv("NATS_STREAM"))
	subject := strings.TrimSpace(os.Getenv("NATS_SUBJECT"))
	durable := strings.TrimSpace(os.Getenv("NATS_DURABLE"))
	if durable == "" {
		durable = "bubble_test"
	}

	if url == "" || stream == "" || subject == "" {
		t.Skip("nats env is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	c, err := Connect(ctx, config.NATSConfig{
		URL:        url,
		Stream:     stream,
		Subject:    subject,
		Durable:    durable,
		AckWait:    5 * time.Second,
		MaxDeliver: 3,
		Ensure:     true,
	}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()

	codec := queue.JSONCodec[natsPayload]{}

	pub, err := NewPublisher(c.JS, config.NATSConfig{Subject: subject}, codec)
	if err != nil {
		t.Fatalf("publisher: %v", err)
	}

	cons, err := NewConsumer(c.JS, config.NATSConfig{Subject: subject, Durable: durable, AckWait: 5 * time.Second}, codec)
	if err != nil {
		t.Fatalf("consumer: %v", err)
	}

	gotCh := make(chan natsPayload, 1)
	if err := cons.Start(ctx, func(ctx context.Context, msg natsPayload) error {
		_ = ctx
		select {
		case gotCh <- msg:
		default:
		}
		return nil
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer cons.Stop(context.Background())

	want := natsPayload{Value: "hello"}
	if err := pub.Publish(ctx, want, queue.PublishOptions{}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case got := <-gotCh:
		if got.Value != want.Value {
			t.Fatalf("unexpected payload: got=%q want=%q", got.Value, want.Value)
		}
	case <-ctx.Done():
		t.Fatalf("timeout waiting for message")
	}
}
