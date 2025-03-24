package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/dmitrymomot/gokit/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadFileFromRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Create a test HTTP request with a file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		
		// Create a form file
		fileContents := "test file contents"
		part, err := writer.CreateFormFile("file", "testfile.jpg")
		require.NoError(t, err)
		_, err = io.Copy(part, strings.NewReader(fileContents))
		require.NoError(t, err)
		
		err = writer.Close()
		require.NoError(t, err)
		
		// Create a test request
		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Set the PutObject mock function
		putObjectCalled := false
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			putObjectCalled = true
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Equal(t, "uploads/testfile.jpg", *params.Key)
			return &s3.PutObjectOutput{}, nil
		}
		
		result, err := client.UploadFileFromRequest(context.Background(), req, storage.UploadFromRequestOptions{
			IsPublic: true,
		})
		
		require.NoError(t, err)
		assert.Equal(t, "uploads/testfile.jpg", result.Path)
		assert.Contains(t, result.URL, "testfile.jpg")
		assert.True(t, putObjectCalled)
	})
	
	t.Run("nil request", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		_, err := client.UploadFileFromRequest(context.Background(), nil, storage.UploadFromRequestOptions{})
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrInvalidRequest))
		
		// PutObject should not be called
		assert.Nil(t, mockClient.PutObjectFunc)
	})
	
	t.Run("missing file", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Create a test HTTP request without a file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		
		err := writer.Close()
		require.NoError(t, err)
		
		// Create a test request
		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		_, err = client.UploadFileFromRequest(context.Background(), req, storage.UploadFromRequestOptions{})
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrInvalidRequest))
		
		// PutObject should not be called
		assert.Nil(t, mockClient.PutObjectFunc)
	})
	
	t.Run("custom field name", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Create a test HTTP request with a file
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		
		// Create a form file with custom field name
		fileContents := "test file contents"
		part, err := writer.CreateFormFile("custom_field", "testfile.jpg")
		require.NoError(t, err)
		_, err = io.Copy(part, strings.NewReader(fileContents))
		require.NoError(t, err)
		
		err = writer.Close()
		require.NoError(t, err)
		
		// Create a test request
		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		// Set the PutObject mock function
		putObjectCalled := false
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			putObjectCalled = true
			return &s3.PutObjectOutput{}, nil
		}
		
		result, err := client.UploadFileFromRequest(context.Background(), req, storage.UploadFromRequestOptions{
			Field: "custom_field",
		})
		
		require.NoError(t, err)
		assert.Contains(t, result.Path, "testfile.jpg")
		assert.True(t, putObjectCalled)
	})
}

func TestListFiles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return files
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Equal(t, "uploads/test/", *params.Prefix)
			
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{
						Key:  aws.String("uploads/test/file1.jpg"),
						Size: aws.Int64(100),
					},
					{
						Key:  aws.String("uploads/test/file2.png"),
						Size: aws.Int64(200),
					},
					// Directory object (ends with /)
					{
						Key:  aws.String("uploads/test/subdir/"),
						Size: aws.Int64(0),
					},
				},
			}, nil
		}
		
		// Mock HeadObject to return content types
		headObjectCalled := 0
		mockClient.HeadObjectFunc = func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			headObjectCalled++
			
			// Return content type based on file extension
			var contentType string
			if strings.HasSuffix(*params.Key, ".jpg") {
				contentType = "image/jpeg"
			} else if strings.HasSuffix(*params.Key, ".png") {
				contentType = "image/png"
			}
			
			return &s3.HeadObjectOutput{
				ContentType: aws.String(contentType),
			}, nil
		}
		
		files, err := client.ListFiles(context.Background(), "test")
		
		require.NoError(t, err)
		assert.Len(t, files, 2) // Should exclude the directory object
		assert.True(t, listObjectsV2Called)
		assert.Equal(t, 2, headObjectCalled) // Called for each non-directory file
		
		// Verify first file
		assert.Equal(t, "uploads/test/file1.jpg", files[0].Path)
		assert.Equal(t, "https://cdn.example.com/uploads/test/file1.jpg", files[0].URL)
		assert.Equal(t, int64(100), files[0].Size)
		assert.Equal(t, "image/jpeg", files[0].ContentType)
		
		// Verify second file
		assert.Equal(t, "uploads/test/file2.png", files[1].Path)
		assert.Equal(t, "https://cdn.example.com/uploads/test/file2.png", files[1].URL)
		assert.Equal(t, int64(200), files[1].Size)
		assert.Equal(t, "image/png", files[1].ContentType)
	})
	
	t.Run("list error", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return an error
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			return &s3.ListObjectsV2Output{}, errors.New("list error")
		}
		
		files, err := client.ListFiles(context.Background(), "test")
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrFailedToListFiles))
		assert.Nil(t, files)
		
		assert.True(t, listObjectsV2Called)
	})
	
	t.Run("empty list", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return empty list
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{},
			}, nil
		}
		
		files, err := client.ListFiles(context.Background(), "test")
		
		require.NoError(t, err)
		assert.Empty(t, files)
		
		assert.True(t, listObjectsV2Called)
	})
	
	t.Run("directory objects excluded", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return a mix of files and directories
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{
						Key:  aws.String("uploads/test/file1.jpg"),
						Size: aws.Int64(1024),
					},
					{
						Key:  aws.String("uploads/test/subdir/"), // This is a directory
						Size: aws.Int64(0),
					},
				},
			}, nil
		}
		
		// Mock HeadObject for content type
		headObjectCalled := false
		mockClient.HeadObjectFunc = func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
			headObjectCalled = true
			if *params.Key == "uploads/test/file1.jpg" {
				return &s3.HeadObjectOutput{
					ContentType: aws.String("image/jpeg"),
				}, nil
			}
			return nil, errors.New("unexpected key")
		}
		
		files, err := client.ListFiles(context.Background(), "test")
		
		require.NoError(t, err)
		require.Len(t, files, 1) // Only one file, not the directory
		assert.Equal(t, "uploads/test/file1.jpg", files[0].Path)
		
		assert.True(t, listObjectsV2Called)
		assert.True(t, headObjectCalled)
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock DeleteObject to succeed
		deleteObjectCalled := false
		mockClient.DeleteObjectFunc = func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
			deleteObjectCalled = true
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Equal(t, "uploads/test.jpg", *params.Key)
			return &s3.DeleteObjectOutput{}, nil
		}
		
		err := client.DeleteFile(context.Background(), "test.jpg")
		
		require.NoError(t, err)
		assert.True(t, deleteObjectCalled)
	})
	
	t.Run("delete error", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock DeleteObject to return an error
		deleteObjectCalled := false
		mockClient.DeleteObjectFunc = func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
			deleteObjectCalled = true
			return &s3.DeleteObjectOutput{}, errors.New("delete error")
		}
		
		err := client.DeleteFile(context.Background(), "test.jpg")
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrFailedToDeleteFile))
		assert.True(t, deleteObjectCalled)
	})
}

func TestDeleteDirectory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return files
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Equal(t, "uploads/test/", *params.Prefix)
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{
						Key: aws.String("uploads/test/file1.jpg"),
					},
					{
						Key: aws.String("uploads/test/file2.png"),
					},
				},
			}, nil
		}
		
		// Mock DeleteObjects to succeed
		deleteObjectsCalled := false
		mockClient.DeleteObjectsFunc = func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
			deleteObjectsCalled = true
			assert.Equal(t, "test-bucket", *params.Bucket)
			assert.Len(t, params.Delete.Objects, 2)
			assert.Equal(t, "uploads/test/file1.jpg", *params.Delete.Objects[0].Key)
			assert.Equal(t, "uploads/test/file2.png", *params.Delete.Objects[1].Key)
			return &s3.DeleteObjectsOutput{}, nil
		}
		
		err := client.DeleteDirectory(context.Background(), "test")
		
		require.NoError(t, err)
		assert.True(t, listObjectsV2Called)
		assert.True(t, deleteObjectsCalled)
	})
	
	t.Run("list error", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return an error
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			return &s3.ListObjectsV2Output{}, errors.New("list error")
		}
		
		err := client.DeleteDirectory(context.Background(), "test")
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrFailedToDeleteDirectory))
		assert.True(t, listObjectsV2Called)
	})
	
	t.Run("empty directory", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return empty list
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{},
			}, nil
		}
		
		err := client.DeleteDirectory(context.Background(), "test")
		
		require.NoError(t, err)
		assert.True(t, listObjectsV2Called)
	})
	
	t.Run("delete error", func(t *testing.T) {
		mockClient, client := setupTest(t)
		
		// Mock ListObjectsV2 to return files
		listObjectsV2Called := false
		mockClient.ListObjectsV2Func = func(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
			listObjectsV2Called = true
			return &s3.ListObjectsV2Output{
				Contents: []types.Object{
					{
						Key: aws.String("uploads/test/file1.jpg"),
					},
				},
			}, nil
		}
		
		// Mock DeleteObjects to return an error
		deleteObjectsCalled := false
		mockClient.DeleteObjectsFunc = func(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
			deleteObjectsCalled = true
			return &s3.DeleteObjectsOutput{}, errors.New("delete error")
		}
		
		err := client.DeleteDirectory(context.Background(), "test")
		
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrFailedToDeleteDirectory))
		assert.True(t, listObjectsV2Called)
		assert.True(t, deleteObjectsCalled)
	})
}
