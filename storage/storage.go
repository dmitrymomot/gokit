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
)

// Config holds all necessary fields.
type Config struct {
	Key            string `env:"STORAGE_KEY,required"`
	Secret         string `env:"STORAGE_SECRET,required"`
	Region         string `env:"STORAGE_REGION,required"`
	Bucket         string `env:"STORAGE_BUCKET,required"`
	Endpoint       string `env:"STORAGE_ENDPOINT,required"`
	CDN            string `env:"STORAGE_CDN"`
	MaxFileSize    int64  `env:"STORAGE_MAX_FILE_SIZE" envDefault:"10485760"`
	UploadBasePath string `env:"STORAGE_BASE_PATH" envDefault:"uploads"`
	ForcePathStyle bool   `env:"STORAGE_FORCE_PATH_STYLE" envDefault:"false"`
}

type File struct {
	Path        string
	URL         string
	Size        int64
	ContentType string
}

type UploadOptions struct {
	ContentType string
	Path        string
}

type UploadFromRequestOptions struct {
	ContentType string
	Path        string
	Field       string
}

// Storage is the public interface.
type Storage interface {
	GetFileURL(path string) string
	UploadFile(ctx context.Context, file []byte, opts UploadOptions) (File, error)
	UploadFileFromRequest(ctx context.Context, r *http.Request, opts UploadFromRequestOptions) (File, error)
	ListFiles(ctx context.Context, path string) ([]File, error)
	DeleteFile(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, path string) error
}
