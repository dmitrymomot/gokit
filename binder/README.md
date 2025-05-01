# Binder Package

A simple and efficient HTTP request data binding package for Go structs.

## Installation

```bash
go get github.com/dmitrymomot/gokit/binder
```

## Overview

The `binder` package provides utilities for binding HTTP request data to Go structs. It automatically handles different content types including JSON, form data, and query parameters with minimal configuration. This package is thread-safe and designed for high-performance web applications.

## Features

- Content-type based automatic binding
- Support for JSON, form data, and query parameters
- Tag-based field mapping with `json` or `form` tags
- Support for nested structs with dot notation
- Support for slice types and maps
- Automatic time.Time parsing with multiple formats
- Comprehensive error handling

## Usage

### Basic Example

```go
package main

import (
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/binder"
)

type User struct {
	Name  string   `json:"name" form:"name"`
	Email string   `json:"email" form:"email"`
	Age   int      `json:"age" form:"age"`
	Roles []string `json:"roles" form:"roles"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	
	// Bind request data to the struct
	if err := binder.Bind(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Process the user data...
	log.Printf("User: %+v", user)
	
	// Respond with success
	w.WriteHeader(http.StatusCreated)
}
```

### Specific Binding Methods

```go
// JSON binding
var jsonData User
if err := binder.BindJSON(r, &jsonData); err != nil {
    // Handle error
}
// jsonData contains the parsed JSON data

// Form data binding
var formData User
if err := binder.BindForm(r, &formData); err != nil {
    // Handle error
}
// formData contains the parsed form data

// Query parameter binding
var queryParams User
if err := binder.BindQuery(r, &queryParams); err != nil {
    // Handle error
}
// queryParams contains the parsed query parameters
```

### Nested Struct Binding

```go
type Address struct {
    Street  string `json:"street" form:"street"`
    City    string `json:"city" form:"city"`
    ZipCode string `json:"zip_code" form:"zip_code"`
}

type User struct {
    Name    string  `json:"name" form:"name"`
    Email   string  `json:"email" form:"email"`
    Address Address `json:"address" form:"address"` // Nested struct
}

// For JSON requests:
// {"name":"John","email":"john@example.com","address":{"street":"123 Main St","city":"San Francisco","zip_code":"94105"}}

// For form data or query parameters:
// name=John&email=john@example.com&address.street=123+Main+St&address.city=San+Francisco&address.zip_code=94105
```

### Working with Time Fields

```go
type Event struct {
    Title     string    `json:"title" form:"title"`
    StartDate time.Time `json:"start_date" form:"start_date"`
}

// The binder will automatically parse time fields using multiple formats:
// - RFC3339: "2023-04-25T14:30:00Z"
// - RFC3339Nano: "2023-04-25T14:30:00.123456789Z"
// - "2023-04-25T14:30:00"
// - "2023-04-25 14:30:00"
// - "2023-04-25"
```

### Error Handling

```go
import (
    "errors"
    "fmt"
    "net/http"
    
    "github.com/dmitrymomot/gokit/binder"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    var data User
    
    err := binder.Bind(r, &data)
    if err != nil {
        switch {
        case errors.Is(err, binder.ErrInvalidContentType):
            http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
            return
            
        case errors.Is(err, binder.ErrEmptyBody):
            http.Error(w, "Request body is empty", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrInvalidJSON):
            http.Error(w, "Invalid JSON format", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrInvalidFormData):
            http.Error(w, "Invalid form data", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrUnsupportedType):
            http.Error(w, "Target type not supported", http.StatusInternalServerError)
            return
            
        default:
            http.Error(w, "Bad request", http.StatusBadRequest)
            return
        }
    }
    
    // Process valid data...
}
```

## Best Practices

1. **Struct Design**:
   - Use consistent field naming conventions
   - Always include both `json` and `form` tags to support multiple content types
   - Add validation tags if using with a validation package

2. **Error Handling**:
   - Check for specific error types to provide meaningful error messages
   - Consider wrapping errors with context information
   - Return appropriate HTTP status codes based on error types

3. **Performance**:
   - Pre-define your struct types rather than using map[string]interface{}
   - For very large requests, consider binding only the needed fields

4. **Security**:
   - Always validate and sanitize bound data before use
   - Set appropriate size limits for request bodies

## API Reference

### Functions

```go
func Bind(r *http.Request, v any) error
```
Automatically selects the binding method based on the request's Content-Type header. For GET requests, it binds from query parameters.

```go
func BindJSON(r *http.Request, v any) error
```
Binds JSON request body to the provided struct.

```go
func BindForm(r *http.Request, v any) error
```
Binds form data from the request to the provided struct. Handles both regular and multipart forms.

```go
func BindQuery(r *http.Request, v any) error
```
Binds query parameters from the request to the provided struct.

### Error Types

```go
var ErrInvalidRequest = errors.New("invalid request")
var ErrInvalidContentType = errors.New("invalid content type")
var ErrEmptyBody = errors.New("empty body")
var ErrInvalidJSON = errors.New("invalid JSON")
var ErrInvalidFormData = errors.New("invalid form data")
var ErrUnsupportedType = errors.New("unsupported type")
