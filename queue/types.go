package queue

import "context"

type Publisher[T any] interface {
	Publish(ctx context.Context, msg T, opts PublishOptions) error
}

type Consumer[T any] interface {
	Start(ctx context.Context, handler Handler[T]) error
	Stop(ctx context.Context) error
}

type Handler[T any] func(context.Context, T) error

type Header struct {
	Name  string
	Value string
}

type PublishOptions struct {
	Key     string
	Headers []Header
}

