# Webhook Package

A flexible, thread-safe client for sending webhook HTTP requests with retry and logging capabilities.

## Installation

```bash
go get github.com/dmitrymomot/gokit/webhook
```

## Overview

The `webhook` package provides a robust client for sending HTTP webhook requests to external services with configurable retry logic and structured logging. It follows a clean decorator pattern, allowing you to add capabilities like automatic retries and comprehensive logging to a base webhook sender. All components are designed to be thread-safe for concurrent usage.

## Features

- Flexible HTTP webhook sending with support for all common HTTP methods
- Automatic parameter handling (JSON body or query parameters based on HTTP method)
- Configurable retry mechanism with exponential backoff and custom retry conditions
- Structured logging with privacy controls for sensitive data
- Thread-safe implementation for concurrent webhook requests
- Decorator pattern for modular extension of functionality
- Context support for timeouts and cancellation
- Zero external dependencies beyond the Go standard library

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
	
	// Send a POST request with JSON payload
	params := map[string]any{
		"event": "user.created",
		"user_id": 123456,
		"timestamp": time.Now().Unix(),
	}
	
	// Send the webhook request
	resp, err := sender.Send(context.Background(), "https://api.example.com/webhook", params)
	if err != nil {
		// Handle error
		fmt.Printf("Failed to send webhook: %v\n", err)
		return
	}
	
	// Check response status
	if !resp.IsSuccessful() {
		fmt.Printf("Request failed with status %d\n", resp.StatusCode)
		return
	}
	
	// Response body is available as resp.Body ([]byte)
	// Headers are available as resp.Headers (http.Header)
	// Request duration is available as resp.Duration (time.Duration)
}
```

### Using Different HTTP Methods

```go
// GET request with query parameters
// Parameters are automatically converted to a query string
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
// Sends request to: https://api.example.com/search?term=golang&limit=10

// PUT request with custom headers
resp, err := sender.Send(
	ctx, 
	"https://api.example.com/users/123", 
	userDataMap,
	webhook.WithMethod("PUT"),
	webhook.WithHeader("Authorization", "Bearer token123"),
	webhook.WithHeader("X-Custom-Header", "value"),
)
```

### Retry Decorator

```go
// Create a base webhook sender
baseSender := webhook.NewWebhookSender()

// Add retry capabilities
retrySender := webhook.NewRetryDecorator(
	baseSender,
	webhook.WithRetryCount(3),                // Retry up to 3 times
	webhook.WithRetryDelay(1 * time.Second),  // Start with 1 second delay
	webhook.WithRetryBackoff(),               // Use exponential backoff
	webhook.WithRetryOnServerErrors(),        // Retry on 5xx errors
	webhook.WithRetryOnNetworkErrors(),       // Retry on network failures
)

// Use the retry-enabled sender
resp, err := retrySender.Send(ctx, "https://api.example.com/webhook", params)
```

### Logging Decorator

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
loggedSender := webhook.NewLoggerDecorator(
	baseSender, 
	logger,
	webhook.WithMaskedFields("api_key", "password", "token"),
)

// Use the logging-enabled sender
resp, err := loggedSender.Send(ctx, "https://api.example.com/webhook", params)
// Logs request details (with sensitive fields masked) and response details
```

### Combined Decorators

```go
// Create a base webhook sender
baseSender := webhook.NewWebhookSender()

// Add logging capability first (inner decorator)
loggedSender := webhook.NewLoggerDecorator(
	baseSender,
	logger,
	webhook.WithMaskedFields("password", "token"),
)

// Add retry capability (outer decorator)
sender := webhook.NewRetryDecorator(
	loggedSender,
	webhook.WithRetryCount(3),
	webhook.WithRetryBackoff(),
	webhook.WithRetryOnNetworkErrors(),
)

// Use the fully decorated sender
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
// The request will be logged both before sending and after receiving a response
// If the request fails, it will be retried up to 3 times with logging for each attempt
```

### Error Handling

```go
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
if err != nil {
	switch {
	case errors.Is(err, webhook.ErrInvalidURL):
		// Handle invalid URL error
	case errors.Is(err, webhook.ErrMarshalParams):
		// Handle parameter marshaling error
	case errors.Is(err, webhook.ErrSendRequest):
		// Handle network or connection error
	case errors.Is(err, webhook.ErrResponseTimeout):
		// Handle timeout error
	default:
		// Handle other errors
	}
	return
}

// Check response status
if !resp.IsSuccessful() {
	// Handle unsuccessful HTTP status code (non 2xx)
	fmt.Printf("Request failed with status %d: %s\n", resp.StatusCode, resp.Body)
	return
}
```

## Best Practices

1. **Error Handling**:
   - Always check both the error return value and the response status code
   - Use `errors.Is()` to check for specific error types
   - Implement proper logging for failed requests for troubleshooting

2. **Security**:
   - Use `WithMaskedFields()` to protect sensitive data in logs
   - Always use HTTPS URLs for webhook endpoints
   - Avoid hardcoding authentication tokens in your code

3. **Reliability**:
   - Use the retry decorator for important webhooks or unreliable endpoints
   - Configure reasonable timeout values based on the expected response time
   - Consider implementing circuit breakers for consistently failing endpoints

4. **Performance**:
   - Reuse webhook sender instances instead of creating new ones for each request
   - For high-volume webhook sending, consider using goroutines for concurrent requests
   - Use context timeouts to avoid waiting indefinitely for slow endpoints

## API Reference

### Core Interfaces and Types

```go
// WebhookSender defines the core interface for sending webhooks
type WebhookSender interface {
	Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error)
}

// Request represents a webhook request
type Request struct {
	URL     string
	Method  string
	Headers map[string]string
	Params  any
	Timeout time.Duration
}

// Response contains the HTTP response details
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
	Request    *Request
}
```

### Functions

```go
func NewWebhookSender(opts ...SenderOption) WebhookSender
```
Creates a new webhook sender with the specified options.

```go
func NewRetryDecorator(sender WebhookSender, opts ...RetryOption) WebhookSender
```
Wraps a webhook sender with retry capabilities.

```go
func NewLoggerDecorator(sender WebhookSender, logger *slog.Logger, opts ...LoggerOption) WebhookSender
```
Wraps a webhook sender with logging capabilities.

### Methods

```go
func (r *Response) IsSuccessful() bool
```
Returns true if the response status code is in the 2xx range.

### Configuration Options

#### Sender Options
```go
func WithHTTPClient(client *http.Client) SenderOption
func WithDefaultTimeout(timeout time.Duration) SenderOption
func WithDefaultHeaders(headers map[string]string) SenderOption
func WithDefaultMethod(method string) SenderOption
func WithMaxRetries(retries int) SenderOption
func WithRetryInterval(interval time.Duration) SenderOption
```

#### Request Options
```go
func WithMethod(method string) RequestOption
func WithHeader(key, value string) RequestOption
func WithHeaders(headers map[string]string) RequestOption
func WithRequestTimeout(timeout time.Duration) RequestOption
```

#### Retry Options
```go
func WithRetryCount(max int) RetryOption
func WithRetryDelay(interval time.Duration) RetryOption
func WithRetryBackoff() RetryOption
func WithRetryOnStatus(statusCodes ...int) RetryOption
func WithRetryOnServerErrors() RetryOption
func WithRetryOnNetworkErrors() RetryOption
func WithRetryLogger(logger *slog.Logger) RetryOption
```

#### Logger Options
```go
func WithHideParams() LoggerOption
func WithMaskedFields(fields ...string) LoggerOption
```

### Error Types
```go
var ErrInvalidURL = errors.New("invalid webhook URL")
var ErrInvalidMethod = errors.New("invalid HTTP method")
var ErrMarshalParams = errors.New("failed to marshal request parameters")
var ErrCreateRequest = errors.New("failed to create HTTP request")
var ErrSendRequest = errors.New("failed to send HTTP request")
var ErrReadResponse = errors.New("failed to read HTTP response")
var ErrResponseTimeout = errors.New("webhook request timed out")
```
