# QR Code Package

Simple QR code generation for Go applications with PNG and base64 output formats.

## Installation

```bash
go get github.com/dmitrymomot/gokit/qrcode
```

## Overview

The `qrcode` package provides a simplified interface for generating QR codes in Go applications. It offers straightforward functions to create QR codes as PNG images or as base64-encoded data URIs ready for HTML embedding.

## Features

- Simple API with just two functions
- Generate QR codes as PNG byte slices
- Generate QR codes as base64 data URIs for direct HTML embedding
- Consistent error handling
- Zero dependencies beyond the high-performance go-qrcode library

## Usage

### Generate QR Code as PNG

```go
import (
    "os"
    "github.com/dmitrymomot/gokit/qrcode"
)

// Generate a QR code for a URL (default size is 256px if size <= 0)
png, err := qrcode.Generate("https://example.com", 256)
if err != nil {
    // Handle error
}

// Save to file
err = os.WriteFile("example.png", png, 0644)
```

### Generate QR Code for HTML

```go
import (
    "fmt"
    "github.com/dmitrymomot/gokit/qrcode"
)

// Generate QR code as base64 data URI
dataURI, err := qrcode.GenerateBase64Image("https://example.com", 256)
if err != nil {
    // Handle error
}

// Use in HTML
html := fmt.Sprintf(`<img src="%s" alt="QR Code">`, dataURI)
```

### Using with HTTP Handlers

```go
http.HandleFunc("/qrcode", func(w http.ResponseWriter, r *http.Request) {
    content := r.URL.Query().Get("content")
    if content == "" {
        http.Error(w, "Missing content parameter", http.StatusBadRequest)
        return
    }
    
    // Generate QR code
    png, err := qrcode.Generate(content, 256)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Serve the image
    w.Header().Set("Content-Type", "image/png")
    w.Header().Set("Cache-Control", "public, max-age=86400")
    w.Write(png)
})
```

## API Reference

### Generate QR Code

```go
// Generate creates a QR code image in PNG format with the given content.
// Returns the image as a byte slice or an error if generation fails.
func Generate(content string, size int) ([]byte, error)
```

- **Parameters**:
  - `content`: Text/URL to encode in the QR code
  - `size`: Size of the QR code in pixels (defaults to 256 if <= 0)
- **Returns**:
  - `[]byte`: PNG image as byte slice
  - `error`: Error if generation fails

### Generate Base64 Image

```go
// GenerateBase64Image creates a base64 encoded string representation of a QR code
// Returns the base64 encoded data URI or an error if generation fails.
func GenerateBase64Image(content string, size int) (string, error)
```

- **Parameters**:
  - `content`: Text/URL to encode in the QR code
  - `size`: Size of the QR code in pixels (defaults to 256 if <= 0)
- **Returns**:
  - `string`: Base64-encoded data URI (format: `data:image/png;base64,...`)
  - `error`: Error if generation fails

## Error Handling

```go
// Check for specific error
if errors.Is(err, qrcode.ErrorFailedToGenerateQRCode) {
    // Handle QR code generation failure
}
```

## Best Practices

1. **Validate input content**: Ensure the content to be encoded is valid
2. **Choose appropriate size**: 256px is a good default for most use cases
3. **Add error handling**: Always check for errors during generation
4. **Consider caching**: For static QR codes, generate once and cache
5. **Set proper content type**: When serving QR codes via HTTP, use `image/png`
