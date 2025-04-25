# Storage Package

A robust S3-compatible storage client with a clean, abstract API for file operations.

## Installation

```bash
go get github.com/dmitrymomot/gokit/storage
```

## Overview

The `storage` package provides a simple abstraction over Amazon S3 and other S3-compatible storage services. It offers a clean interface for common operations like file uploads, downloads, listing, and deletion with proper error handling and configuration options.

## Features

- Clean interface abstracting underlying S3 operations
- Flexible configuration using functional options pattern
- Automatic content type detection with extensive MIME mappings
- Comprehensive error handling with clear error types
- Support for both public and private file permissions
- Size limits and validation for uploads
- Multi-part upload support for large files
- Directory-style operations (list, delete recursively)
- HTTP request integration for direct uploads

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

fmt.Printf("Uploaded to: %s\n", file.URL)

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
	
	// Return the file information
	json.NewEncoder(w).Encode(map[string]string{
		"url": file.URL,
		"path": file.Path,
	})
}
```

### Managing Files

```go
// List files in a directory
files, err := client.ListFiles(ctx, "examples/")
if err != nil {
	// Handle error
}

for _, file := range files {
	fmt.Printf("Path: %s, Size: %d bytes, URL: %s\n", file.Path, file.Size, file.URL)
}

// Get a file's URL
url := client.GetFileURL("examples/hello.txt")

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

### File Model

```go
// File represents information about a stored file
type File struct {
	Path        string // File path in the storage
	URL         string // Full URL to access the file
	Size        int64  // File size in bytes
	ContentType string // MIME content type
}
```

### Upload Options

```go
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
	MaxFileSize int64             // Override default max file size
}
```

### Configuration Options

```go
// WithHTTPClient sets a custom HTTP client
WithHTTPClient(client *http.Client) ClientOption

// WithS3Client uses a pre-configured S3 client
WithS3Client(client *s3.Client) ClientOption

// WithRetryMaxAttempts sets the maximum retry attempts
WithRetryMaxAttempts(attempts int) ClientOption

// WithRetryMode configures the retry behavior
WithRetryMode(mode aws.RetryMode) ClientOption

// WithS3ConfigOption adds a custom S3 config option
WithS3ConfigOption(option func(*s3config.LoadOptions) error) ClientOption

// WithS3ClientOption adds a custom S3 client option
WithS3ClientOption(option func(*s3.Options)) ClientOption
```

## Error Handling

The package defines several specific error types for better error handling:

```go
// Check for specific error types
if errors.Is(err, storage.ErrFileTooLarge) {
    // Handle file size exceeds maximum
}

if errors.Is(err, storage.ErrFailedToUploadFile) {
    // Handle upload failure
}
```

Available error types:
- `ErrFailedToUploadFile`: File upload failed
- `ErrFileTooLarge`: File exceeds the maximum size limit
- `ErrFailedToDeleteFile`: File deletion failed
- `ErrFailedToDeleteDirectory`: Directory deletion failed
- `ErrFailedToListFiles`: File listing failed
- `ErrInvalidRequest`: Request is invalid or missing file data
- `ErrMissingConfig`: Required configuration is missing
- `ErrInvalidEndpoint`: Provided endpoint URL is invalid
- `ErrFailedToLoadConfig`: AWS configuration loading failed

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
