# Webhook Package

A flexible, configurable client for sending webhook HTTP requests with retry and logging support.

## Installation

```bash
go get github.com/dmitrymomot/gokit/webhook
```

## Overview

The `webhook` package provides a robust client for sending HTTP webhook requests to external services with support for various configuration options, automatic retries, and comprehensive logging capabilities.

## Features

- Flexible request configuration with sensible defaults
- Multiple HTTP methods (GET, POST, PUT, DELETE, etc.)
- Automatic parameter handling (JSON for request bodies, query strings for GET/DELETE)
- Configurable retry mechanism with exponential backoff
- Detailed request/response logging with privacy controls
- Context support for cancellation and timeouts
- Decorator pattern for extensibility
- Thread-safe implementation

## Usage

### Basic Request

```go
import (
	"context"
	"fmt"
	"github.com/dmitrymomot/gokit/webhook"
)

func main() {
	// Create a new webhook sender
	sender := webhook.NewWebhookSender()
	
	// Create a context
	ctx := context.Background()
	
	// Send a POST request with JSON payload
	params := map[string]any{
		"event": "user.created",
		"user_id": 123456,
		"timestamp": time.Now().Unix(),
	}
	
	// Send the webhook request
	resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
	if err != nil {
		// Handle error
		fmt.Printf("Failed to send webhook: %v\n", err)
		return
	}
	
	// Check response status
	if !resp.IsSuccessful() {
		fmt.Printf("Request failed with status %d: %s\n", resp.StatusCode, resp.Body)
		return
	}
	
	fmt.Println("Webhook sent successfully!")
}
```

### Custom HTTP Method and Headers

```go
// Send with custom HTTP method and headers
resp, err := sender.Send(
	ctx, 
	"https://api.example.com/webhook", 
	params,
	webhook.WithMethod("PUT"),
	webhook.WithHeader("Authorization", "Bearer token123"),
	webhook.WithHeader("X-Custom-Header", "value"),
)
```

### GET Request with Query Parameters

```go
// Parameters are automatically converted to query string for GET requests
// This will make a request to: https://api.example.com/search?term=golang&limit=10
params := map[string]any{
	"term": "golang",
	"limit": 10,
}

resp, err := sender.Send(
	ctx, 
	"https://api.example.com/search", 
	params,
	webhook.WithMethod("GET"),
)
```

### Using Structs as Parameters

```go
// Define a struct with JSON tags
type SearchParams struct {
	Term     string `json:"term"`
	Limit    int    `json:"limit"`
	Page     int    `json:"page"`
	SortBy   string `json:"sort_by,omitempty"`
}

// Create search parameters
params := SearchParams{
	Term:   "golang",
	Limit:  10,
	Page:   1,
	SortBy: "relevance",
}

// For GET, parameters are converted to query string
// For POST/PUT, parameters are marshaled to JSON
resp, err := sender.Send(ctx, "https://api.example.com/search", params, webhook.WithMethod("GET"))
```

### Configuring a Sender

```go
// Create a sender with custom configuration
sender := webhook.NewWebhookSender(
	webhook.WithDefaultMethod("POST"),
	webhook.WithDefaultHeaders(map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Version": "1.0",
		"User-Agent": "GoKit/1.0",
	}),
	webhook.WithDefaultTimeout(10 * time.Second),
	webhook.WithMaxRetries(3),
	webhook.WithRetryInterval(500 * time.Millisecond),
)

// Use the configured sender
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
```

### Adding Retry Capability

```go
// Create a base webhook sender
baseSender := webhook.NewWebhookSender()

// Add retry capabilities with configuration
sender := webhook.NewRetryDecorator(
	baseSender,
	webhook.WithRetryCount(5),            // Retry up to 5 times
	webhook.WithRetryDelay(1 * time.Second), // Wait 1 second between retries
	webhook.WithRetryBackoff(),           // Use exponential backoff
	webhook.WithRetryOnServerErrors(),    // Retry on 5xx errors
)

// Send with automatic retries for failures
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
```

### Adding Logging Capability

```go
import (
	"log/slog"
	"os"
)

// Create a logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// Create a base webhook sender
baseSender := webhook.NewWebhookSender()

// Add logging capability with masked sensitive fields
sender := webhook.NewLoggerDecorator(
	baseSender, 
	logger,
	webhook.WithMaskedFields("api_key", "password", "token"),
)

// Use the logging-enabled sender
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
```

### Combining Multiple Decorators

```go
// Create a base webhook sender
baseSender := webhook.NewWebhookSender()

// Add logging capability first
loggedSender := webhook.NewLoggerDecorator(
	baseSender,
	logger,
	webhook.WithMaskedFields("password", "token"),
)

// Add retry capability on top of logging
sender := webhook.NewRetryDecorator(
	loggedSender,
	webhook.WithRetryCount(3),
	webhook.WithRetryBackoff(),
	webhook.WithRetryOnNetworkErrors(),
)

// Use the fully decorated sender
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
```

## API Reference

### Core Interface

```go
// WebhookSender defines the interface for sending webhook requests
type WebhookSender interface {
	Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error)
}
```

### Response Type

```go
// Response contains the HTTP response details
type Response struct {
	StatusCode int           // HTTP status code
	Body       []byte        // Response body
	Headers    http.Header   // Response headers
	Duration   time.Duration // Request duration
	Request    *Request      // Original request details
}

// Check if response status code indicates success (2xx)
func (r *Response) IsSuccessful() bool
```

### Configuration Options

#### Sender Options

```go
// Configure the base webhook sender
WithHTTPClient(client *http.Client) SenderOption
WithDefaultTimeout(timeout time.Duration) SenderOption
WithDefaultHeaders(headers map[string]string) SenderOption
WithDefaultMethod(method string) SenderOption
WithMaxRetries(retries int) SenderOption
WithRetryInterval(interval time.Duration) SenderOption
```

#### Request Options

```go
// Configure individual requests
WithMethod(method string) RequestOption
WithHeader(key, value string) RequestOption
WithHeaders(headers map[string]string) RequestOption
WithRequestTimeout(timeout time.Duration) RequestOption
```

#### Retry Options

```go
// Configure retry behavior
WithRetryCount(max int) RetryOption
WithRetryDelay(interval time.Duration) RetryOption
WithRetryBackoff() RetryOption
WithRetryOnStatus(statusCodes ...int) RetryOption
WithRetryOnServerErrors() RetryOption
WithRetryOnNetworkErrors() RetryOption
WithRetryLogger(logger *slog.Logger) RetryOption
```

#### Logger Options

```go
// Configure logging behavior
WithHideParams() LoggerOption
WithMaskedFields(fields ...string) LoggerOption
```

## Best Practices

1. **Error Handling**:
   - Always check for errors returned from `Send`
   - Verify response success with `resp.IsSuccessful()`
   - Log failed requests and response bodies for debugging

2. **Security**:
   - Use `WithMaskedFields` to protect sensitive data in logs
   - Store authentication tokens securely, not in code
   - Use HTTPS URLs for all webhook endpoints

3. **Performance**:
   - Configure appropriate timeouts based on endpoint responsiveness
   - Use retry with backoff for unreliable endpoints
   - Consider using context with timeout for long-running operations

4. **Reliability**:
   - Set reasonable retry counts for important webhooks
   - Use exponential backoff for retry intervals
   - Implement circuit breaker patterns for consistently failing endpoints
