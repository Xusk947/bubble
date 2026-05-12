package objectstore

import (
	"context"
	"io"
	"time"
)

type ObjectStore interface {
	Put(ctx context.Context, key string, body io.Reader, opts PutOptions) error
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectMeta, error)
	Delete(ctx context.Context, key string) error
	SignedURL(ctx context.Context, key string, opts SignedURLOptions) (string, error)
}

type PutOptions struct {
	ContentType  string
	CacheControl string
}

type SignedURLOptions struct {
	Expires time.Duration
}

type ObjectMeta struct {
	ContentLength int64
	ContentType   string
	ETag          string
	LastModified  time.Time
}
