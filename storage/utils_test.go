package storage_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dmitrymomot/gokit/storage"
	"github.com/stretchr/testify/require"
)

// TestContentTypeToExtension tests the internal getExtByContentType function indirectly
func TestContentTypeToExtension(t *testing.T) {
	t.Run("common content types", func(t *testing.T) {
		mockClient, client := setupTest(t)

		// Test with known content types
		testCases := []struct {
			contentType string
			extension   string
		}{
			{"image/jpeg", ".jpg"},
			{"image/png", ".png"},
			{"application/pdf", ".pdf"},
			{"text/plain", ".txt"},
			{"application/octet-stream", ".bin"},
		}

		for _, tc := range testCases {
			// Create a new mock function for each test case
			mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				// Verify the key has the correct extension
				require.True(t, strings.HasSuffix(*params.Key, tc.extension), "Key should end with %s, got %s", tc.extension, *params.Key)
				return &s3.PutObjectOutput{}, nil
			}

			// Call UploadFile with no path (which will generate one with timestamp and extension)
			_, err := client.UploadFile(context.Background(), []byte("test"), storage.UploadOptions{
				ContentType: tc.contentType,
			})

			require.NoError(t, err)
		}
	})

	t.Run("content type with parameters", func(t *testing.T) {
		mockClient, client := setupTest(t)

		// Set the PutObject mock function
		keySuffix := ""
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			keySuffix = (*params.Key)[strings.LastIndex(*params.Key, "."):]
			return &s3.PutObjectOutput{}, nil
		}

		// Call UploadFile with a content type that includes parameters
		_, err := client.UploadFile(context.Background(), []byte("test"), storage.UploadOptions{
			ContentType: "text/plain; charset=utf-8",
		})

		require.NoError(t, err)
		require.Equal(t, ".txt", keySuffix)
	})

	t.Run("unknown content type", func(t *testing.T) {
		mockClient, client := setupTest(t)

		// Set the PutObject mock function
		putObjectCalled := false
		mockClient.PutObjectFunc = func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
			putObjectCalled = true
			return &s3.PutObjectOutput{}, nil
		}

		// Call UploadFile with an unknown content type
		_, err := client.UploadFile(context.Background(), []byte("test"), storage.UploadOptions{
			ContentType: "application/x-unknown",
		})

		require.NoError(t, err)
		require.True(t, putObjectCalled)
	})
}
