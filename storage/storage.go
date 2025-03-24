// storage.go
package storage

import (
	"context"
	"net/http"
)

// Storage interface and related types.
type (
	Storage interface {
		GetFileURL(path string) string
		UploadFile(ctx context.Context, file []byte, opts UploadOptions) (File, error)
		UploadFileFromRequest(ctx context.Context, r *http.Request, opts UploadFromRequestOptions) (File, error)
		ListFiles(ctx context.Context, path string) ([]File, error)
		DeleteFile(ctx context.Context, path string) error
		DeleteDirectory(ctx context.Context, path string) error
	}

	File struct {
		Path        string
		URL         string
		Size        int64
		ContentType string
	}

	UploadOptions struct {
		ContentType string
		Path        string
		IsPublic    bool
		Metadata    map[string]string
	}

	UploadFromRequestOptions struct {
		ContentType string
		Path        string
		Field       string
		IsPublic    bool
		Metadata    map[string]string
	}
)
