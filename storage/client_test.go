package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dmitrymomot/gokit/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S3ClientInterface defines the interface for S3 client methods
type S3ClientInterface interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

// MockS3Client is a mock implementation of the S3 client interface
type MockS3Client struct {
	PutObjectFunc     func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	HeadObjectFunc    func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	ListObjectsV2Func func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
	DeleteObjectFunc  func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjectsFunc func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

// PutObject implements the PutObject method for the S3 client interface
func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.PutObjectFunc != nil {
		return m.PutObjectFunc(ctx, params, optFns...)
	}
	return &s3.PutObjectOutput{}, nil
}

// HeadObject implements the HeadObject method for the S3 client interface
func (m *MockS3Client) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	if m.HeadObjectFunc != nil {
		return m.HeadObjectFunc(ctx, params, optFns...)
	}
	return &s3.HeadObjectOutput{}, nil
}

// ListObjectsV2 implements the ListObjectsV2 method for the S3 client interface
func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	if m.ListObjectsV2Func != nil {
		return m.ListObjectsV2Func(ctx, params, optFns...)
	}
	return &s3.ListObjectsV2Output{}, nil
}

// DeleteObject implements the DeleteObject method for the S3 client interface
func (m *MockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	if m.DeleteObjectFunc != nil {
		return m.DeleteObjectFunc(ctx, params, optFns...)
	}
	return &s3.DeleteObjectOutput{}, nil
}

// DeleteObjects implements the DeleteObjects method for the S3 client interface
func (m *MockS3Client) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	if m.DeleteObjectsFunc != nil {
		return m.DeleteObjectsFunc(ctx, params, optFns...)
	}
	return &s3.DeleteObjectsOutput{}, nil
}

// setupTest sets up a mock S3 client and a storage client for testing
func setupTest(t *testing.T) (*MockS3Client, storage.Storage) {
	t.Helper()
	
	// Create the mock S3 client
	mockClient := &MockS3Client{}
	
	// Create storage configuration with generous max file size
	cfg := storage.Config{
		Key:            "test-key",
		Secret:         "test-secret",
		Region:         "us-east-1",
		Bucket:         "test-bucket",
		Endpoint:       "https://s3.amazonaws.com",
		MaxFileSize:    10 * 1024 * 1024, // 10MB for tests
		UploadBasePath: "uploads",         // Set the upload base path for consistent testing
		CDN:            "https://cdn.example.com",
	}

	// Pass the mock client using the WithS3Client option
	client, err := storage.New(context.Background(), cfg, storage.WithS3Client(mockClient))
	require.NoError(t, err)
	require.NotNil(t, client)
	
	return mockClient, client
}

func TestNew(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := storage.Config{
			Key:         "test-key",
			Secret:      "test-secret",
			Region:      "us-east-1",
			Bucket:      "test-bucket",
			Endpoint:    "https://s3.amazonaws.com",
			MaxFileSize: 1024 * 1024,
		}
		
		mockClient := &MockS3Client{}
		client, err := storage.New(context.Background(), cfg, storage.WithS3Client(mockClient))
		
		require.NoError(t, err)
		require.NotNil(t, client)
	})
	
	t.Run("invalid config", func(t *testing.T) {
		cfg := storage.Config{}
		
		client, err := storage.New(context.Background(), cfg)
		
		require.Error(t, err)
		require.Nil(t, client)
		assert.ErrorIs(t, err, storage.ErrMissingConfig)
	})
}

func TestGetFileURL(t *testing.T) {
	t.Run("with standard endpoint", func(t *testing.T) {
		mockClient, client := setupTest(t)
		_ = mockClient
		
		url := client.GetFileURL("test.jpg")
		assert.Equal(t, "https://cdn.example.com/test.jpg", url)
	})
	
	t.Run("strip leading slash", func(t *testing.T) {
		mockClient, client := setupTest(t)
		_ = mockClient
		
		url := client.GetFileURL("/test.jpg")
		assert.Equal(t, "https://cdn.example.com/test.jpg", url)
	})
}

func TestUploadFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		fileContent := []byte("test file content")
		
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Equal(t, "uploads/test.jpg", *params.Key)
			assert.Equal(t, "image/jpeg", *params.ContentType)
			assert.Equal(t, types.ObjectCannedACLPublicRead, params.ACL)
			
			// Verify the file body
			buf := new(bytes.Buffer)
			_, err := io.Copy(buf, params.Body)
			assert.NoError(t, err)
			assert.Equal(t, fileContent, buf.Bytes())
			
			return &s3.PutObjectOutput{}, nil
		}
		
		result, err := client.UploadFile(context.Background(), fileContent, storage.UploadOptions{
			ContentType: "image/jpeg",
			Path:        "test.jpg",
			IsPublic:    true,
		})
		
		require.NoError(t, err)
		assert.Equal(t, "uploads/test.jpg", result.Path)
		assert.Equal(t, "https://cdn.example.com/uploads/test.jpg", result.URL)
		assert.Equal(t, int64(len(fileContent)), result.Size)
		assert.Equal(t, "image/jpeg", result.ContentType)
	})
	
	t.Run("detect content type", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// A PNG file signature
		fileContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
		
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Equal(t, "image/png", *params.ContentType)
			return &s3.PutObjectOutput{}, nil
		}
		
		result, err := client.UploadFile(context.Background(), fileContent, storage.UploadOptions{
			Path: "test.png",
		})
		
		require.NoError(t, err)
		assert.Equal(t, "image/png", result.ContentType)
	})
	
	t.Run("file too large", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Large file content
		fileContent := make([]byte, storage.DefaultMaxFileSize+1)
		
		_, err := client.UploadFile(context.Background(), fileContent, storage.UploadOptions{
			ContentType: "image/jpeg",
			Path:        "test.jpg",
		})
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrFileTooLarge))
		
		// Verify that PutObject was never called
		assert.Nil(t, mockClient.PutObjectFunc)
	})
	
	t.Run("upload error", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		fileContent := []byte("test file content")
		mockError := errors.New("upload failed")
		
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			return &s3.PutObjectOutput{}, mockError
		}
		
		_, err := client.UploadFile(context.Background(), fileContent, storage.UploadOptions{
			ContentType: "image/jpeg",
			Path:        "test.jpg",
		})
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrFailedToUploadFile))
	})
}
