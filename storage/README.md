# Storage Package

The `storage` package provides a robust S3-compatible storage client with a clean, well-structured API. It offers a simple abstraction over Amazon S3 and other S3-compatible storage services.

## Features

- **Clean Interface Design**: Simple API that abstracts the underlying S3 operations
- **Comprehensive Error Handling**: Proper error wrapping with descriptive error messages
- **Flexible Configuration**: Option pattern for customization without breaking the main API
- **Safety Checks**: Validation for file size limits and configuration requirements
- **Content Type Detection**: Automatic content type detection with extensive mappings
- **Secure ACL Management**: Support for public and private file permissions

## Installation

```bash
go get github.com/dmitrymomot/gokit/storage
```

## Usage

### Basic Initialization

```go
package main

import (
	"context"
	"log"
	
	"github.com/dmitrymomot/gokit/storage"
)

func main() {
	cfg := storage.Config{
		Key:        "your-access-key",
		Secret:     "your-secret-key",
		Region:     "us-west-1",
		Bucket:     "your-bucket",
		Endpoint:   "https://s3.amazonaws.com",
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
	
	ctx := context.Background()
	client, err := storage.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}
	
	// Now you can use the client
}
```

### Custom Configuration with Options

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dmitrymomot/gokit/storage"
)

func main() {
	cfg := storage.Config{
		Key:      "your-access-key",
		Secret:   "your-secret-key",
		Region:   "us-west-1",
		Bucket:   "your-bucket",
		Endpoint: "https://s3.amazonaws.com",
	}
	
	// Custom HTTP client
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	ctx := context.Background()
	client, err := storage.New(
		ctx,
		cfg,
		storage.WithHTTPClient(httpClient),
		storage.WithRetryMaxAttempts(5),
		storage.WithRetryMode(aws.RetryModeAdaptive),
	)
	if err != nil {
		log.Fatalf("Failed to initialize storage client: %v", err)
	}
	
	// Now you can use the client
}
```

### Uploading a File

```go
func uploadExample(client storage.Storage) {
	// File content
	fileBytes := []byte("Hello, World!")
	
	ctx := context.Background()
	file, err := client.UploadFile(ctx, fileBytes, storage.UploadOptions{
		Path:        "examples/hello.txt",
		ContentType: "text/plain",
		IsPublic:    true,
		Metadata: map[string]string{
			"description": "Sample text file",
		},
	})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
	
	log.Printf("File uploaded successfully: %s", file.URL)
}
```

### Uploading From HTTP Request

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	file, err := client.UploadFileFromRequest(ctx, r, storage.UploadFromRequestOptions{
		Field:    "file", // Form field name (defaults to "file")
		Path:     "uploads/user-123/",
		IsPublic: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Return the file URL
	w.Write([]byte(file.URL))
}
```

### Listing Files

```go
func listFiles(client storage.Storage) {
	ctx := context.Background()
	files, err := client.ListFiles(ctx, "examples/")
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}
	
	for _, file := range files {
		log.Printf("Path: %s, Size: %d bytes, URL: %s, ContentType: %s", 
			file.Path, file.Size, file.URL, file.ContentType)
	}
}
```

### Deleting Files and Directories

```go
func deleteExample(client storage.Storage) {
	ctx := context.Background()
	
	// Delete a single file
	err := client.DeleteFile(ctx, "examples/hello.txt")
	if err != nil {
		log.Fatalf("Failed to delete file: %v", err)
	}
	
	// Delete an entire directory
	err = client.DeleteDirectory(ctx, "examples/")
	if err != nil {
		log.Fatalf("Failed to delete directory: %v", err)
	}
}
```

## API Reference

### Types

#### `Storage` Interface

```go
type Storage interface {
	GetFileURL(path string) string
	UploadFile(ctx context.Context, file []byte, opts UploadOptions) (File, error)
	UploadFileFromRequest(ctx context.Context, r *http.Request, opts UploadFromRequestOptions) (File, error)
	ListFiles(ctx context.Context, path string) ([]File, error)
	DeleteFile(ctx context.Context, path string) error
	DeleteDirectory(ctx context.Context, path string) error
}
```

#### `File` Struct

```go
type File struct {
	Path        string
	URL         string
	Size        int64
	ContentType string
}
```

#### `Config` Struct

```go
type Config struct {
	Key            string
	Secret         string
	Region         string
	Bucket         string
	Endpoint       string
	CDN            string
	MaxFileSize    int64
	UploadBasePath string
	ForcePathStyle bool
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	MaxRetries     int
	RetryBaseDelay time.Duration
}
```

### Configuration Options

- `WithHTTPClient(client *http.Client)`: Set a custom HTTP client
- `WithS3Client(client *s3.Client)`: Use a pre-configured S3 client
- `WithRetryMaxAttempts(attempts int)`: Set the maximum number of retry attempts
- `WithRetryMode(mode aws.RetryMode)`: Configure the retry behavior
- `WithS3ConfigOption(option func(*s3config.LoadOptions) error)`: Add a custom S3 config option
- `WithS3ClientOption(option func(*s3.Options))`: Add a custom S3 client option

## Error Handling

The package provides the following pre-defined errors:

- `ErrFailedToUploadFile`: Indicated when file upload fails
- `ErrFileTooLarge`: Indicates when a file exceeds the maximum size limit
- `ErrFailedToDeleteFile`: Indicates when file deletion fails
- `ErrFailedToDeleteDirectory`: Indicates when directory deletion fails
- `ErrFailedToListFiles`: Indicates when file listing fails
- `ErrInvalidRequest`: Indicates when the request is invalid or missing file data
- `ErrMissingConfig`: Indicates when required configuration is missing
- `ErrInvalidEndpoint`: Indicates when the provided endpoint URL is invalid
- `ErrFailedToLoadConfig`: Indicates when AWS configuration loading fails
