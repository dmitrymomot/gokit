package storage

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Client defines the interface for S3 client methods used by the storage package
type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

// ClientOption defines a function that configures the storage client
type ClientOption func(*clientOptions)

// clientOptions contains additional configurable options beyond Config
type clientOptions struct {
	httpClient      *http.Client
	s3Client        S3Client
	s3ConfigOptions []func(*s3config.LoadOptions) error
	s3ClientOptions []func(*s3.Options)
}

// WithHTTPClient sets a custom HTTP client for S3 requests
func WithHTTPClient(client *http.Client) ClientOption {
	return func(o *clientOptions) {
		o.httpClient = client
	}
}

// WithS3Client sets a custom pre-configured S3 client
func WithS3Client(client S3Client) ClientOption {
	return func(o *clientOptions) {
		o.s3Client = client
	}
}

// WithS3ConfigOption adds a custom S3 config option
func WithS3ConfigOption(option func(*s3config.LoadOptions) error) ClientOption {
	return func(o *clientOptions) {
		o.s3ConfigOptions = append(o.s3ConfigOptions, option)
	}
}

// WithS3ClientOption adds a custom S3 client option
func WithS3ClientOption(option func(*s3.Options)) ClientOption {
	return func(o *clientOptions) {
		o.s3ClientOptions = append(o.s3ClientOptions, option)
	}
}

// WithRetryMaxAttempts sets the maximum number of retry attempts
func WithRetryMaxAttempts(attempts int) ClientOption {
	return func(o *clientOptions) {
		o.s3ConfigOptions = append(o.s3ConfigOptions, s3config.WithRetryMaxAttempts(attempts))
	}
}

// WithRetryMode sets the retry mode for AWS requests
func WithRetryMode(mode aws.RetryMode) ClientOption {
	return func(o *clientOptions) {
		o.s3ConfigOptions = append(o.s3ConfigOptions, s3config.WithRetryMode(mode))
	}
}
