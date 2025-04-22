# QR Code Package

## Overview

The QR Code package provides simple utilities for generating QR codes in Go applications. It offers functions to create QR codes as PNG images or as base64-encoded strings that can be directly embedded in HTML.

## Documentation

The package wraps the [github.com/skip2/go-qrcode](https://github.com/skip2/go-qrcode) library to provide a simplified and error-enhanced API for QR code generation.

### Functions

#### `Generate(content string, size int) ([]byte, error)`

Creates a QR code image in PNG format with the given content.

- **Parameters**:
  - `content`: The content to encode in the QR code (URL, text, etc.)
  - `size`: The size of the QR code image in pixels (defaults to 256 if value is <= 0)
- **Returns**:
  - `[]byte`: The QR code image as a PNG byte slice
  - `error`: An error if the generation fails

#### `GenerateBase64Image(content string, size int) (string, error)`

Creates a base64-encoded string representation of a QR code image that can be directly used in HTML.

- **Parameters**:
  - `content`: The content to encode in the QR code (URL, text, etc.)
  - `size`: The size of the QR code image in pixels (defaults to 256 if value is <= 0)
- **Returns**:
  - `string`: The base64-encoded string with the appropriate data URI prefix
  - `error`: An error if the generation fails

### Error Handling

The package defines a specific error:

- `ErrorFailedToGenerateQRCode`: Returned when the QR code generation fails

## Usage Examples

### Generating a QR Code as PNG Bytes

```go
package main

import (
	"log"
	"os"
	
	"github.com/dmitrymomot/gokit/qrcode"
)

func main() {
	// Generate a QR code for a URL with a size of 256 pixels
	png, err := qrcode.Generate("https://example.com", 256)
	if err != nil {
		log.Fatal(err)
	}
	
	// Save the QR code to a file
	if err := os.WriteFile("example-qrcode.png", png, 0644); err != nil {
		log.Fatal(err)
	}
	
	log.Println("QR code generated successfully!")
}
```

### Generating a QR Code for HTML Embedding

```go
package main

import (
	"fmt"
	"log"
	
	"github.com/dmitrymomot/gokit/qrcode"
)

func main() {
	// Generate a base64-encoded QR code image
	base64Image, err := qrcode.GenerateBase64Image("https://example.com", 256)
	if err != nil {
		log.Fatal(err)
	}
	
	// Create an HTML snippet that displays the QR code
	htmlSnippet := fmt.Sprintf(`<img src="%s" alt="QR Code">`, base64Image)
	
	// The HTML snippet can be used in your templates or served directly
	fmt.Println(htmlSnippet)
}
```
