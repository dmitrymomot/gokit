# Storage Package

A robust S3-compatible storage client with a clean, abstract API for file operations.

## Installation

```bash
go get github.com/dmitrymomot/gokit/storage
```

## Overview

The storage package provides a simple abstraction over Amazon S3 and other S3-compatible storage services. It offers a clean interface for common operations like file uploads, downloads, listing, and deletion with proper error handling and configuration options. The package is thread-safe and can be safely used in concurrent applications.

## Features

- Clean interface abstracting underlying S3 operations
- Automatic content type detection with extensive MIME mappings
- Comprehensive error handling with specific error types
- Support for both public and private file permissions
- Multi-part upload support for large files
- Directory-style operations (list, delete recursively)
- HTTP request integration for direct file uploads
- Thread-safe implementation for concurrent usage

## Usage

### Basic Initialization

```go
import (
	"context"
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
		// Handle error
	}
	
	// Use the client
}
```

### Uploading Files

```go
// From byte slice
fileBytes := []byte("Hello, World!")
file, err := client.UploadFile(ctx, fileBytes, storage.UploadOptions{
	Path:        "examples/hello.txt",
	ContentType: "text/plain",
	IsPublic:    true,
	Metadata: map[string]string{
		"description": "Sample text file",
	},
})
if err != nil {
	// Handle error
}
// Returns: File with Path, URL, Size, and ContentType

// From HTTP request
func handleUpload(w http.ResponseWriter, r *http.Request) {
	file, err := client.UploadFileFromRequest(ctx, r, storage.UploadFromRequestOptions{
		Field:    "file", // Form field name
		Path:     "uploads/user-123/",
		IsPublic: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Returns: File with Path, URL, Size, and ContentType
}
```

### Managing Files

```go
// List files in a directory
files, err := client.ListFiles(ctx, "examples/")
if err != nil {
	// Handle error
}
// Returns: Slice of File structs with details of each file

// Get a file's URL
url := client.GetFileURL("examples/hello.txt")
// Returns: Complete URL to access the file

// Delete a single file
err = client.DeleteFile(ctx, "examples/hello.txt")
if err != nil {
	// Handle error
}

// Delete an entire directory
err = client.DeleteDirectory(ctx, "examples/")
if err != nil {
	// Handle error
}
```

### Error Handling

```go
file, err := client.UploadFile(ctx, largeFileBytes, opts)
if err != nil {
	switch {
	case errors.Is(err, storage.ErrFileTooLarge):
		// Handle file size exceeds maximum
		fmt.Println("File is too large, maximum size is", cfg.MaxFileSize)
	case errors.Is(err, storage.ErrFailedToUploadFile):
		// Handle upload failure
		fmt.Println("Upload failed, please try again")
	default:
		// Handle general error
		fmt.Println("An error occurred:", err)
	}
}
```

### Advanced Configuration

```go
// Custom HTTP client and retry options
httpClient := &http.Client{
	Timeout: 30 * time.Second,
}

client, err := storage.New(
	ctx,
	cfg,
	storage.WithHTTPClient(httpClient),
	storage.WithRetryMaxAttempts(5),
	storage.WithRetryMode(aws.RetryModeAdaptive),
)
if err != nil {
	// Handle error
}

// Using with pre-configured S3 client
customS3Client := s3.NewFromConfig(awsConfig)
client, err := storage.New(
	ctx, 
	cfg, 
	storage.WithS3Client(customS3Client),
)
```

## Best Practices

1. **Security Considerations**:
   - Never embed AWS keys directly in your code
   - Use IAM roles or environment variables for credentials
   - Set appropriate file permissions (public vs. private)
   - Implement file type validation for uploads

2. **Performance Optimization**:
   - Reuse the client to avoid repeated initialization
   - Set appropriate timeouts based on your application needs
   - Configure retries for transient errors
   - Use a CDN for frequently accessed public files

3. **Error Handling**:
   - Always check for and handle errors from all operations
   - Use `errors.Is()` to check for specific error types
   - Implement proper logging for storage operations
   - Consider implementing retry logic for important operations

4. **Path Management**:
   - Use consistent path prefixes for better organization
   - Include user or tenant identifiers in paths for multi-tenant applications
   - Use directory-style paths with trailing slashes for directories

## API Reference

### Core Interface

```go
// Storage defines the interface for file operations
type Storage interface {
	// GetFileURL returns the URL for a file path
	GetFileURL(path string) string
	
	// UploadFile uploads a file from a byte slice
	UploadFile(ctx context.Context, file []byte, opts UploadOptions) (File, error)
	
	// UploadFileFromRequest uploads a file from an HTTP request
	UploadFileFromRequest(ctx context.Context, r *http.Request, opts UploadFromRequestOptions) (File, error)
	
	// ListFiles lists all files in a directory
	ListFiles(ctx context.Context, path string) ([]File, error)
	
	// DeleteFile deletes a single file
	DeleteFile(ctx context.Context, path string) error
	
	// DeleteDirectory deletes a directory and all its contents
	DeleteDirectory(ctx context.Context, path string) error
}
```

### Configuration

```go
// Config holds the storage configuration
type Config struct {
	Key            string        // AWS/S3 access key
	Secret         string        // AWS/S3 secret key
	Region         string        // AWS region
	Bucket         string        // S3 bucket name
	Endpoint       string        // S3 endpoint URL (optional for non-AWS S3 providers)
	CDN            string        // CDN URL prefix (optional)
	MaxFileSize    int64         // Maximum allowed file size in bytes
	UploadBasePath string        // Base path prefix for all uploads
	ForcePathStyle bool          // Use path-style addressing instead of virtual hosted-style
	ConnectTimeout time.Duration // Connection timeout
	RequestTimeout time.Duration // Request timeout
	MaxRetries     int           // Maximum number of retries for failed operations
	RetryBaseDelay time.Duration // Base delay between retries
}
```

### Models and Types

```go
// File represents information about a stored file
type File struct {
	Path        string // File path in the storage
	URL         string // Full URL to access the file
	Size        int64  // File size in bytes
	ContentType string // MIME content type
}

// UploadOptions configures file upload behavior
type UploadOptions struct {
	Path        string            // Storage path (including filename)
	ContentType string            // MIME type (auto-detected if empty)
	IsPublic    bool              // If true, file is publicly accessible
	Metadata    map[string]string // Optional metadata key-value pairs
}

// UploadFromRequestOptions configures HTTP request uploads
type UploadFromRequestOptions struct {
	Field       string            // Form field name (defaults to "file")
	Path        string            // Storage directory path (filename is preserved)
	IsPublic    bool              // If true, file is publicly accessible
	Metadata    map[string]string // Optional metadata key-value pairs
	ContentType string            // Override content type (auto-detected if empty)
}
```

### Configuration Options

```go
// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption

// WithS3Client uses a pre-configured S3 client
func WithS3Client(client *s3.Client) ClientOption

// WithRetryMaxAttempts sets the maximum retry attempts
func WithRetryMaxAttempts(attempts int) ClientOption

// WithRetryMode configures the retry behavior
func WithRetryMode(mode aws.RetryMode) ClientOption

// WithS3ConfigOption adds a custom S3 config option
func WithS3ConfigOption(option func(*s3config.LoadOptions) error) ClientOption

// WithS3ClientOption adds a custom S3 client option
func WithS3ClientOption(option func(*s3.Options)) ClientOption
```

### Error Types

```go
// Package-level error variables
var ErrFailedToUploadFile = errors.New("failed to upload file")
var ErrFileTooLarge = errors.New("file size exceeds the maximum allowed limit")
var ErrFailedToDeleteFile = errors.New("failed to delete file")
var ErrFailedToDeleteDirectory = errors.New("failed to delete directory")
var ErrFailedToListFiles = errors.New("failed to list files")
var ErrInvalidRequest = errors.New("invalid request or missing file data")
var ErrMissingConfig = errors.New("missing required configuration")
var ErrInvalidEndpoint = errors.New("invalid endpoint URL")
var ErrFailedToLoadConfig = errors.New("failed to load AWS configuration")
```
