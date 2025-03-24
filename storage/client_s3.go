package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// s3Client implements Storage.
type s3Client struct {
	client    *s3.Client
	presigner *s3.PresignClient
	config    Config
}

// New creates an S3-compatible client using the provided Config.
func New(cfg Config) (Storage, error) {
	loadOptions := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.Key, cfg.Secret, "")),
	}
	// No custom endpoint resolver needed - we'll set endpoint directly on the S3 client options
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.ForcePathStyle
		if cfg.Endpoint != "" {
			o.EndpointOptions.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})
	presigner := s3.NewPresignClient(client)

	return &s3Client{
		client:    client,
		presigner: presigner,
		config:    cfg,
	}, nil
}

func (s *s3Client) GetFileURL(path string) string {
	// Prepend UploadBasePath to key.
	key := s.config.UploadBasePath + "/" + path

	// If CDN is configured, use it
	if s.config.CDN != "" {
		// Ensure CDN URL doesn't have trailing slash
		cdnBase := strings.TrimSuffix(s.config.CDN, "/")
		return fmt.Sprintf("%s/%s", cdnBase, key)
	}

	// Generate a presigned URL (with a 15-minute expiry).
	presignResult, err := s.presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &s.config.Bucket,
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))

	if err != nil {
		return ""
	}
	return presignResult.URL
}

func (s *s3Client) UploadFile(ctx context.Context, file []byte, opts UploadOptions) (File, error) {
	if int64(len(file)) > s.config.MaxFileSize {
		return File{}, errors.Join(ErrFailedToUploadFile, fmt.Errorf("file size %d exceeds maximum allowed", len(file)))
	}
	key := s.config.UploadBasePath + "/" + opts.Path
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.config.Bucket,
		Key:         aws.String(key),
		Body:        bytes.NewReader(file),
		ContentType: &opts.ContentType,
	})
	if err != nil {
		return File{}, errors.Join(ErrFailedToUploadFile, fmt.Errorf("upload failed: %w", err))
	}
	return File{
		Path:        opts.Path,
		URL:         s.GetFileURL(opts.Path),
		Size:        int64(len(file)),
		ContentType: opts.ContentType,
	}, nil
}

func (s *s3Client) UploadFileFromRequest(ctx context.Context, r *http.Request, opts UploadFromRequestOptions) (File, error) {
	filePart, header, err := r.FormFile(opts.Field)
	if err != nil {
		return File{}, errors.Join(ErrFailedToUploadFile, fmt.Errorf("unable to retrieve file from request: %w", err))
	}
	defer filePart.Close()

	if header.Size > s.config.MaxFileSize {
		return File{}, errors.Join(ErrFailedToUploadFile, fmt.Errorf("file %s exceeds the maximum allowed size", header.Filename))
	}
	data, err := io.ReadAll(filePart)
	if err != nil {
		return File{}, errors.Join(ErrFailedToUploadFile, fmt.Errorf("failed to read file: %w", err))
	}
	uploadOpts := UploadOptions{
		ContentType: opts.ContentType,
		Path:        opts.Path,
	}
	return s.UploadFile(ctx, data, uploadOpts)
}

func (s *s3Client) ListFiles(ctx context.Context, path string) ([]File, error) {
	prefix := s.config.UploadBasePath + "/" + path
	var files []File
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &s.config.Bucket,
		Prefix: aws.String(prefix),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, errors.Join(ErrFailedToListFiles, fmt.Errorf("listing failed: %w", err))
		}
		for _, obj := range page.Contents {
			trimmedPath := strings.TrimPrefix(*obj.Key, s.config.UploadBasePath+"/")
			files = append(files, File{
				Path: trimmedPath,
				URL:  s.GetFileURL(trimmedPath),
				Size: aws.ToInt64(obj.Size),
			})
		}
	}
	return files, nil
}

func (s *s3Client) DeleteFile(ctx context.Context, path string) error {
	key := s.config.UploadBasePath + "/" + path
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.config.Bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return errors.Join(ErrFailedToDeleteFile, fmt.Errorf("deletion failed: %w", err))
	}
	return nil
}

func (s *s3Client) DeleteDirectory(ctx context.Context, path string) error {
	prefix := s.config.UploadBasePath + "/" + path

	// First list all objects with the given prefix
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: &s.config.Bucket,
		Prefix: aws.String(prefix),
	})

	// Collect objects to delete
	var objectsToDelete []types.ObjectIdentifier
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return errors.Join(ErrFailedToDeleteDirectory, fmt.Errorf("retrieving objects failed: %w", err))
		}

		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		// If we have a batch of objects, delete them
		if len(objectsToDelete) > 0 {
			if err := s.deleteObjects(ctx, objectsToDelete); err != nil {
				return err
			}
			objectsToDelete = nil // Reset for next batch
		}
	}

	return nil
}

// deleteObjects deletes a batch of objects in a single API call
func (s *s3Client) deleteObjects(ctx context.Context, objects []types.ObjectIdentifier) error {
	if len(objects) == 0 {
		return nil
	}

	_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: &s.config.Bucket,
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	})

	if err != nil {
		return errors.Join(ErrFailedToDeleteDirectory, fmt.Errorf("batch deletion failed: %w", err))
	}

	return nil
}
