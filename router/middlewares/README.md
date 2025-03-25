# Middlewares

Collection of middlewares for different use cases.

## Request ID Middleware

The Request ID middleware automatically adds a unique request identifier to each incoming HTTP request. If a request already contains a request ID in the `X-Request-Id` header, that value is used; otherwise, a new UUID v4 is generated.

### Features

- Extracts request ID from the `X-Request-Id` header if present
- Generates a new UUID v4 if no request ID is found
- Stores the request ID in the request context
- Adds the request ID to the response headers
- Provides utility functions to get/set request ID from/to context

### Usage

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/dmitrymomot/gokit/router/middlewares"
)

func main() {
	// Create a new handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from context
		requestID := middlewares.GetRequestID(r.Context())
		fmt.Fprintf(w, "Request ID: %s", requestID)
	})

	// Wrap the handler with the RequestID middleware
	wrapped := middlewares.RequestID(handler)

	// Use in your HTTP server
	http.Handle("/", wrapped)
	http.ListenAndServe(":8080", nil)
}
```

### Context Utilities

```go
// Get the request ID from a context
requestID := middlewares.GetRequestID(ctx)

// Add a request ID to a context
ctx = middlewares.WithRequestID(ctx, "custom-request-id")
```

This middleware can be combined with other middleware to provide consistent request tracking throughout your application.

## Recover Middleware

The Recover middleware catches and handles panics that occur during HTTP request processing. It prevents your server from crashing when an unexpected panic occurs and provides a graceful error response to the client.

### Features

- Recovers from panics in HTTP handlers
- Logs the panic with a full stack trace using `log/slog`
- Returns a 500 Internal Server Error response by default
- Allows custom error handlers to be provided
- Does not interfere with non-panicking request flows
- Special handling for `http.ErrAbortHandler` to respect Go's internal abort mechanism

### Usage

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/dmitrymomot/gokit/router/middlewares"
)

func main() {
	// Create a new handler that might panic
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This will be safely caught by the middleware
		if r.URL.Path == "/panic" {
			panic("something went wrong")
		}
		fmt.Fprintf(w, "Hello, World!")
	})

	// Wrap the handler with the Recoverer middleware
	wrapped := middlewares.Recoverer(handler)

	// Use in your HTTP server
	http.Handle("/", wrapped)
	http.ListenAndServe(":8080", nil)
}
```

### Custom Error Handling

You can provide a custom error handler to customize the response:

```go
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/dmitrymomot/gokit/router/middlewares"
)

func main() {
	// Define a custom error handler
	customErrorHandler := func(w http.ResponseWriter, r *http.Request, err any) {
		// Log the error
		slog.ErrorContext(r.Context(), "panic recovered", "error", err, "stack", string(debug.Stack()))
		
		// Return a JSON error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Internal Server Error",
			"message": "The server encountered an unexpected condition",
		})
	}

	// Create a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This will be caught and handled by our custom handler
		panic("something went wrong")
	})

	// Use the RecovererWithHandler middleware with our custom handler
	wrapped := middlewares.RecovererWithHandler(handler, customErrorHandler)

	// Use in your HTTP server
	http.Handle("/", wrapped)
	http.ListenAndServe(":8080", nil)
}
```

The middleware is ideal for use with any HTTP server and can be combined with other middlewares to provide robust error handling.