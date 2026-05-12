package kafka

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/queue"
)

type kafkaPayload struct {
	Value string
}

func TestKafka_PublishConsume(t *testing.T) {
	brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	topic := strings.TrimSpace(os.Getenv("KAFKA_TOPIC"))
	group := strings.TrimSpace(os.Getenv("KAFKA_GROUP_ID"))
	if group == "" {
		group = "bubble_test"
	}

	if brokers == "" || topic == "" {
		t.Skip("kafka env is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg := config.KafkaConfig{Brokers: brokers, Topic: topic, GroupID: group}
	client, err := NewClient(cfg, nil)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	codec := queue.JSONCodec[kafkaPayload]{}

	pub, err := NewPublisher(client, cfg, codec)
	if err != nil {
		t.Fatalf("publisher: %v", err)
	}
	defer pub.Close()

	cons, err := NewConsumer(client, cfg, codec)
	if err != nil {
		t.Fatalf("consumer: %v", err)
	}
	defer cons.Stop(context.Background())

	gotCh := make(chan kafkaPayload, 1)
	if err := cons.Start(ctx, func(ctx context.Context, msg kafkaPayload) error {
		_ = ctx
		select {
		case gotCh <- msg:
		default:
		}
		return nil
	}); err != nil {
		t.Fatalf("start: %v", err)
	}

	want := kafkaPayload{Value: "hello"}
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
