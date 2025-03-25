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