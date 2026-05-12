package s3store

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	bubbleconfig "github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/objectstore"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestS3Store_MinIO_PutGetDelete(t *testing.T) {
	endpoint := strings.TrimSpace(os.Getenv("MINIO_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("MINIO_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("MINIO_SECRET_KEY"))
	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	region := strings.TrimSpace(os.Getenv("MINIO_REGION"))
	if region == "" {
		region = "us-east-1"
	}

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		t.Skip("minio env is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, r string, options ...interface{}) (aws.Endpoint, error) {
				_ = r
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               endpoint,
						SigningRegion:     region,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)),
	)
	if err != nil {
		t.Fatalf("load aws config: %v", err)
	}

	s3client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	_, _ = s3client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	store, err := New(ctx, bubbleconfig.S3Config{
		Endpoint:        endpoint,
		Region:          region,
		Bucket:          bucket,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		UsePathStyle:    true,
		DisableTLS:      strings.HasPrefix(endpoint, "http://"),
	}, nil)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	key := "test/" + time.Now().UTC().Format("20060102T150405.000000000Z07:00")
	payload := []byte("hello")

	if err := store.Put(ctx, key, bytes.NewReader(payload), objectstore.PutOptions{ContentType: "text/plain"}); err != nil {
		t.Fatalf("put: %v", err)
	}

	rc, meta, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected body: %q", string(got))
	}
	if meta.ContentLength == 0 {
		t.Fatalf("expected content length")
	}

	url, err := store.SignedURL(ctx, key, objectstore.SignedURLOptions{Expires: time.Minute})
	if err != nil {
		t.Fatalf("signed url: %v", err)
	}
	if strings.TrimSpace(url) == "" {
		t.Fatalf("expected url")
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
