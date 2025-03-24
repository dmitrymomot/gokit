package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type storageClient struct {
	client  *s3.Client
	baseURL string
	config  Config
}

// New creates a new S3-compatible storage client with the given config and options.
func New(cfg Config, opts ...ClientOption) (Storage, error) {
	// Validate minimum required configuration
	if cfg.Key == "" || cfg.Secret == "" || cfg.Region == "" || cfg.Bucket == "" {
		return nil, ErrMissingConfig
	}

	// Initialize client options
	options := &clientOptions{}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	// If a pre-configured S3 client is provided, use it directly
	var client *s3.Client
	if options.s3Client != nil {
		client = options.s3Client
	} else {
		// Create a custom HTTP client with timeouts if configured and not already provided
		httpClient := options.httpClient
		if httpClient == nil && (cfg.RequestTimeout > 0 || cfg.ConnectTimeout > 0) {
			timeout := cfg.RequestTimeout
			if cfg.ConnectTimeout > 0 {
				timeout = cfg.ConnectTimeout
			}
			httpClient = &http.Client{
				Timeout: timeout,
			}
		}

		// Configure AWS SDK options
		s3Options := []func(*s3config.LoadOptions) error{
			s3config.WithRegion(cfg.Region),
			s3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.Key, cfg.Secret, "")),
		}

		// Add HTTP client with timeout if configured
		if httpClient != nil {
			s3Options = append(s3Options, s3config.WithHTTPClient(httpClient))
		}

		// Add retry configuration if needed
		if cfg.MaxRetries > 0 {
			s3Options = append(s3Options, s3config.WithRetryMaxAttempts(cfg.MaxRetries))
			if cfg.RetryBaseDelay > 0 {
				s3Options = append(s3Options, s3config.WithRetryMode(aws.RetryModeStandard))
			}
		}

		// Add any additional S3 config options
		s3Options = append(s3Options, options.s3ConfigOptions...)

		// Load AWS configuration
		awsCfg, err := s3config.LoadDefaultConfig(context.TODO(), s3Options...)
		if err != nil {
			return nil, errors.Join(ErrFailedToLoadConfig, err)
		}

		// Create the S3 client with endpoint configuration
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if cfg.Endpoint != "" {
				o.BaseEndpoint = aws.String(cfg.Endpoint)
			}
			o.UsePathStyle = cfg.ForcePathStyle

			// Apply any additional S3 client options
			for _, opt := range options.s3ClientOptions {
				opt(o)
			}
		})
	}

	// Derive BaseURL: if CDN is provided, use it; otherwise, construct it from bucket and endpoint
	var baseURL string
	if cfg.CDN != "" {
		baseURL = strings.TrimSuffix(cfg.CDN, "/") + "/"
	} else {
		u, err := url.Parse(cfg.Endpoint)
		if err != nil {
			return nil, errors.Join(ErrInvalidEndpoint, err)
		}
		baseURL = "https://" + cfg.Bucket + "." + u.Host + "/"
	}

	return &storageClient{
		client:  client,
		baseURL: baseURL,
		config:  cfg,
	}, nil
}

// GetFileURL returns the full URL for a file path
func (sc *storageClient) GetFileURL(path string) string {
	path = strings.TrimPrefix(path, "/")
	return sc.baseURL + path
}

// UploadFile uploads a file to the storage service
func (sc *storageClient) UploadFile(ctx context.Context, file []byte, opts UploadOptions) (File, error) {
	if len(file) > int(sc.config.MaxFileSize) {
		return File{}, ErrFileTooLarge
	}

	// Set default content type if not provided
	contentType := opts.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(file)
	}

	// Ensure the path is properly formatted
	filePath := opts.Path
	if filePath == "" {
		filePath = filepath.Join(sc.config.UploadBasePath, fmt.Sprintf("%d%s", time.Now().UnixNano(), getExtByContentType(contentType)))
	} else {
		filePath = filepath.Join(sc.config.UploadBasePath, filePath)
	}
	filePath = strings.TrimPrefix(filePath, "/")

	// Upload the file to S3
	_, err := sc.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(sc.config.Bucket),
		Key:         aws.String(filePath),
		Body:        bytes.NewReader(file),
		ContentType: aws.String(contentType),
		ACL: func() types.ObjectCannedACL {
			if opts.IsPublic {
				return types.ObjectCannedACLPublicRead
			}
			return types.ObjectCannedACLPrivate
		}(),
		Metadata: opts.Metadata,
	})
	if err != nil {
		return File{}, errors.Join(ErrFailedToUploadFile, err)
	}

	return File{
		Path:        filePath,
		URL:         sc.GetFileURL(filePath),
		Size:        int64(len(file)),
		ContentType: contentType,
	}, nil
}

// UploadFileFromRequest uploads a file from an HTTP request
func (sc *storageClient) UploadFileFromRequest(ctx context.Context, r *http.Request, opts UploadFromRequestOptions) (File, error) {
	// Default form field name is "file" if not specified
	field := opts.Field
	if field == "" {
		field = "file"
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(sc.config.MaxFileSize); err != nil {
		return File{}, errors.Join(ErrInvalidRequest, err)
	}

	// Get the file from the form
	formFile, header, err := r.FormFile(field)
	if err != nil {
		return File{}, errors.Join(ErrInvalidRequest, err)
	}
	defer formFile.Close()

	// Check file size
	if header.Size > sc.config.MaxFileSize {
		return File{}, ErrFileTooLarge
	}

	// Read file content
	fileBytes, err := io.ReadAll(formFile)
	if err != nil {
		return File{}, errors.Join(ErrInvalidRequest, err)
	}

	// Determine content type
	contentType := opts.ContentType
	if contentType == "" {
		contentType = header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(fileBytes)
		}
	}

	// Generate file path if not provided
	filePath := opts.Path
	if filePath == "" {
		// Use original filename or generate one
		filePath = header.Filename
		if filePath == "" {
			filePath = fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(header.Filename))
		}
	}

	// Upload the file
	return sc.UploadFile(ctx, fileBytes, UploadOptions{
		ContentType: contentType,
		Path:        filePath,
		IsPublic:    opts.IsPublic,
		Metadata:    opts.Metadata,
	})
}

// ListFiles lists files in a directory
func (sc *storageClient) ListFiles(ctx context.Context, dirPath string) ([]File, error) {
	// Ensure path is properly formatted
	if !strings.HasSuffix(dirPath, "/") && dirPath != "" {
		dirPath += "/"
	}

	prefix := filepath.Join(sc.config.UploadBasePath, dirPath)
	prefix = strings.TrimPrefix(prefix, "/")

	// List objects in the directory
	result, err := sc.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(sc.config.Bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, errors.Join(errors.New("failed to list files"), err)
	}

	files := make([]File, 0, len(result.Contents))
	for _, obj := range result.Contents {
		// Skip directories
		if strings.HasSuffix(*obj.Key, "/") {
			continue
		}

		files = append(files, File{
			Path: *obj.Key,
			URL:  sc.GetFileURL(*obj.Key),
			Size: *obj.Size,
			// ContentType is not available in ListObjectsV2 response
			ContentType: "",
		})
	}

	return files, nil
}

// DeleteFile deletes a file from storage
func (sc *storageClient) DeleteFile(ctx context.Context, filePath string) error {
	filePath = filepath.Join(sc.config.UploadBasePath, filePath)
	filePath = strings.TrimPrefix(filePath, "/")

	_, err := sc.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(sc.config.Bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return errors.Join(ErrFailedToDeleteFile, err)
	}

	return nil
}

// DeleteDirectory deletes all files in a directory
func (sc *storageClient) DeleteDirectory(ctx context.Context, dirPath string) error {
	// Ensure path is properly formatted
	if !strings.HasSuffix(dirPath, "/") && dirPath != "" {
		dirPath += "/"
	}

	prefix := filepath.Join(sc.config.UploadBasePath, dirPath)
	prefix = strings.TrimPrefix(prefix, "/")

	// List all objects to delete
	result, err := sc.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(sc.config.Bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return errors.Join(ErrFailedToDeleteDirectory, err)
	}

	if len(result.Contents) == 0 {
		return nil // No files to delete
	}

	// Prepare objects to delete
	toDelete := make([]types.ObjectIdentifier, 0, len(result.Contents))
	for _, obj := range result.Contents {
		toDelete = append(toDelete, types.ObjectIdentifier{
			Key: obj.Key,
		})
	}

	// Delete objects in batch
	_, err = sc.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(sc.config.Bucket),
		Delete: &types.Delete{
			Objects: toDelete,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return errors.Join(ErrFailedToDeleteDirectory, err)
	}

	return nil
}
