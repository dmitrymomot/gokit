package storage

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ClientOption defines a function that configures the storage client
type ClientOption func(*clientOptions)

// clientOptions contains additional configurable options beyond Config
type clientOptions struct {
	httpClient      *http.Client
	s3Client        *s3.Client
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
func WithS3Client(client *s3.Client) ClientOption {
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
	return WithS3ConfigOption(s3config.WithRetryMaxAttempts(attempts))
}

// WithRetryMode sets the retry mode for AWS requests
func WithRetryMode(mode aws.RetryMode) ClientOption {
	return WithS3ConfigOption(s3config.WithRetryMode(mode))
}
