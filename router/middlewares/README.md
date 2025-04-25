# HTTP Middlewares Package

A collection of essential HTTP middlewares for robust web services.

## Installation

```bash
go get github.com/dmitrymomot/gokit/router/middlewares
```

## Overview

The `middlewares` package provides a collection of reusable HTTP middleware components that help build reliable, traceable, and resilient web services. These middlewares follow standard Go patterns and can be used with any HTTP router or directly with Go's `net/http` package.

## Features

- Request ID middleware for consistent request tracking
- Panic recovery middleware with customizable error handling
- Context-based utilities for middleware communication
- Transparent integration with standard Go interfaces
- Compatible with any HTTP router respecting the `http.Handler` interface

## Usage

### Request ID Middleware

Automatically adds a unique identifier to every request for tracking and debugging:

```go
import (
    "fmt"
    "net/http"
    
    "github.com/dmitrymomot/gokit/router/middlewares"
)

func main() {
    // Create a handler that uses the request ID
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get request ID from context
        requestID := middlewares.GetRequestID(r.Context())
        fmt.Fprintf(w, "Request ID: %s", requestID)
    })
    
    // Apply the RequestID middleware
    wrappedHandler := middlewares.RequestID(handler)
    
    http.Handle("/", wrappedHandler)
    http.ListenAndServe(":8080", nil)
}
```

### Panic Recovery Middleware

Catches panics in handlers and prevents server crashes:

```go
import (
    "net/http"
    
    "github.com/dmitrymomot/gokit/router/middlewares"
)

func main() {
    // Handler that might panic
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/panic" {
            panic("something went wrong")
        }
        w.Write([]byte("Hello World"))
    })
    
    // Apply the Recoverer middleware
    safeHandler := middlewares.Recoverer(handler)
    
    http.Handle("/", safeHandler)
    http.ListenAndServe(":8080", nil)
}
```

### Custom Error Handling

Customize panic recovery responses:

```go
import (
    "encoding/json"
    "log/slog"
    "net/http"
    "runtime/debug"
    
    "github.com/dmitrymomot/gokit/router/middlewares"
)

func main() {
    // Custom error handler for panics
    customErrorHandler := func(w http.ResponseWriter, r *http.Request, err any) {
        // Log the error with stack trace
        slog.ErrorContext(r.Context(), "panic recovered", 
            "error", err, 
            "stack", string(debug.Stack()),
            "request_id", middlewares.GetRequestID(r.Context()),
        )
        
        // Return JSON error response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{
            "error": "Internal Server Error",
            "request_id": middlewares.GetRequestID(r.Context()),
        })
    }
    
    // Apply middleware with custom handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your handler logic
    })
    
    safeHandler := middlewares.RecovererWithHandler(handler, customErrorHandler)
    http.Handle("/", safeHandler)
    http.ListenAndServe(":8080", nil)
}
```

### Working with Request Context

Get and set values in the request context:

```go
// Get request ID from context
requestID := middlewares.GetRequestID(ctx)

// Add request ID to context
ctx = middlewares.WithRequestID(ctx, "custom-request-id")
```

### Combining Middlewares

Chain middlewares together:

```go
func main() {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your handler logic
    })
    
    // Apply middlewares in order (outside to inside)
    // 1. First recover from panics
    // 2. Then add request ID
    wrappedHandler := middlewares.Recoverer(
        middlewares.RequestID(handler),
    )
    
    http.Handle("/", wrappedHandler)
    http.ListenAndServe(":8080", nil)
}
```

## API Reference

### Request ID Middleware

```go
// Add a request ID to requests
func RequestID(next http.Handler) http.Handler

// Get the request ID from context
func GetRequestID(ctx context.Context) string

// Add a request ID to context
func WithRequestID(ctx context.Context, reqID string) context.Context

// The standard header used for request IDs
const RequestIDHeader = "X-Request-Id"
```

### Recover Middleware

```go
// Basic panic recovery
func Recoverer(next http.Handler) http.Handler

// Panic recovery with custom error handler
func RecovererWithHandler(next http.Handler, handler RecovererErrorHandlerFunc) http.Handler

// Default error handler function
func DefaultRecovererErrorHandler(w http.ResponseWriter, r *http.Request, err any)

// Type definition for error handler function
type RecovererErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err any)
```

## Best Practices

1. **Apply middlewares in the correct order**: Typically recovery middleware should be applied first (outermost), followed by request ID and other middlewares
2. **Use request IDs for tracing**: Include request IDs in logs and error responses to help with debugging and tracing request flows
3. **Customize error responses**: Use `RecovererWithHandler` to provide user-friendly error responses while still logging complete details
4. **Add context values as needed**: The pattern shown with request IDs can be extended for other values that need to be accessible throughout the request lifecycle
5. **Combine with logging middleware**: These middlewares work well with structured logging to create comprehensive request tracing