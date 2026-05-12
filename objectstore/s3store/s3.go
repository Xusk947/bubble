package s3store

import (
	"context"
	"io"
	"net/url"
	"strings"
	"time"

	bubbleconfig "github.com/Xusk947/bubble/config"
	"github.com/Xusk947/bubble/objectstore"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type Store struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
}

func New(ctx context.Context, cfg bubbleconfig.S3Config, logger *zap.Logger) (*Store, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, objectstore.InvalidConfigError{Field: "bucket", Message: "empty"}
	}
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, objectstore.InvalidConfigError{Field: "region", Message: "empty"}
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" && strings.TrimSpace(cfg.SecretAccessKey) != "" {
		return nil, objectstore.InvalidConfigError{Field: "secret_access_key", Message: "set without access_key_id"}
	}
	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, objectstore.InvalidConfigError{Field: "secret_access_key", Message: "empty"}
	}

	endpoint, err := normalizeEndpoint(cfg.Endpoint, cfg.DisableTLS)
	if err != nil {
		return nil, objectstore.InvalidConfigError{Field: "endpoint", Message: err.Error()}
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		withEndpoint(endpoint, cfg.Region),
		withCredentials(cfg),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	if logger != nil {
		logger.Info(
			"s3 initialized",
			zap.String("s3_endpoint", endpoint),
			zap.String("s3_region", cfg.Region),
			zap.String("s3_bucket", cfg.Bucket),
			zap.Bool("s3_use_path_style", cfg.UsePathStyle),
			zap.Bool("s3_disable_tls", cfg.DisableTLS),
		)
	}

	return &Store{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
	}, nil
}

func (s *Store) Put(ctx context.Context, key string, body io.Reader, opts objectstore.PutOptions) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         body,
		ContentType:  optionalString(opts.ContentType),
		CacheControl: optionalString(opts.CacheControl),
	})
	if err != nil {
		return objectstore.OperationError{Op: "put", Key: key, Cause: err}
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, objectstore.ObjectMeta, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, objectstore.ObjectMeta{}, objectstore.OperationError{Op: "get", Key: key, Cause: err}
	}

	meta := objectstore.ObjectMeta{
		ContentLength: valueOrInt64(out.ContentLength),
		ContentType:   valueOrEmpty(out.ContentType),
		ETag:          valueOrEmpty(out.ETag),
		LastModified:  valueOrTime(out.LastModified),
	}
	return out.Body, meta, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return objectstore.OperationError{Op: "delete", Key: key, Cause: err}
	}
	return nil
}

func (s *Store) SignedURL(ctx context.Context, key string, opts objectstore.SignedURLOptions) (string, error) {
	expires := opts.Expires
	if expires <= 0 {
		expires = 15 * time.Minute
	}
	out, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(p *s3.PresignOptions) {
		p.Expires = expires
	})
	if err != nil {
		return "", objectstore.OperationError{Op: "signed_url", Key: key, Cause: err}
	}
	return out.URL, nil
}

func normalizeEndpoint(value string, disableTLS bool) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", nil
	}
	if strings.Contains(v, "://") {
		u, err := url.Parse(v)
		if err != nil {
			return "", err
		}
		if u.Host == "" {
			return "", &url.Error{Op: "parse", URL: v, Err: url.InvalidHostError("")}
		}
		return u.String(), nil
	}
	scheme := "https"
	if disableTLS {
		scheme = "http"
	}
	u, err := url.Parse(scheme + "://" + v)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", &url.Error{Op: "parse", URL: v, Err: url.InvalidHostError("")}
	}
	return u.String(), nil
}

func withCredentials(cfg bubbleconfig.S3Config) func(*awsconfig.LoadOptions) error {
	if strings.TrimSpace(cfg.AccessKeyID) == "" && strings.TrimSpace(cfg.SecretAccessKey) == "" && strings.TrimSpace(cfg.SessionToken) == "" {
		return func(*awsconfig.LoadOptions) error { return nil }
	}
	return awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken))
}

func withEndpoint(endpoint string, signingRegion string) func(*awsconfig.LoadOptions) error {
	if strings.TrimSpace(endpoint) == "" {
		return func(*awsconfig.LoadOptions) error { return nil }
	}
	return awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			_ = region
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               endpoint,
					PartitionID:       "aws",
					SigningRegion:     signingRegion,
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	))
}

func optionalString(value string) *string {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return aws.String(v)
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func valueOrTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func valueOrInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

var _ objectstore.ObjectStore = (*Store)(nil)
