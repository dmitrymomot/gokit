# Webhook

A flexible, configurable webhook client for sending HTTP requests with support for custom methods, headers, and request parameters.

## Overview

The webhook package provides a simple but powerful client for sending webhook requests to external services. It supports:

- Configurable HTTP methods
- Custom headers for each request
- Default and per-request configuration options
- Automatic JSON marshaling of request parameters
- Automatic conversion of parameters to query string for GET, HEAD, DELETE methods
- Retry mechanism for failed requests
- Response processing with status code validation
- Context support for cancellation and timeouts

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/webhook"
)

func main() {
	// Create a new webhook sender with default options
	sender := webhook.NewWebhookSender()

	// Send a POST request with JSON payload
	params := map[string]string{
		"event": "user.created",
		"user_id": "123456",
	}

	ctx := context.Background()
	resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
	if err != nil {
		log.Fatalf("Failed to send webhook: %v", err)
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response body: %s\n", resp.Body)
}
```

### Custom HTTP Method and Headers

```go
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

For HTTP methods that don't support request bodies (GET, HEAD, DELETE), parameters are automatically converted to query string parameters:

```go
// Parameters will be automatically converted to query string
// This will make a request to: https://api.example.com/search?term=golang&limit=10
params := map[string]string{
	"term": "golang",
	"limit": "10",
}
resp, err := sender.Send(
	ctx, 
	"https://api.example.com/search", 
	params,
	webhook.WithMethod("GET"),
)
```

### Using Structs as Parameters

You can also use structs as parameters, which will be automatically marshaled to JSON for POST/PUT requests, or converted to query string parameters for GET/HEAD/DELETE requests:

```go
// Define a struct with JSON tags
type SearchParams struct {
	Term  string `json:"term"`
	Limit int    `json:"limit"`
}

// For GET request, this will be converted to: ?term=golang&limit=10
// For POST request, this will be marshaled to JSON: {"term":"golang","limit":10}
params := SearchParams{
	Term: "golang",
	Limit: 10,
}

resp, err := sender.Send(ctx, "https://api.example.com/search", params, webhook.WithMethod("GET"))
```

### Creating a Configured Sender

```go
// Create a webhook sender with custom configuration
sender := webhook.NewWebhookSender(
	webhook.WithDefaultMethod("POST"),
	webhook.WithDefaultHeaders(map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Version": "1.0",
	}),
	webhook.WithDefaultTimeout(10 * time.Second),
	webhook.WithMaxRetries(3),
	webhook.WithRetryInterval(500 * time.Millisecond),
)
```

### Using Request-Specific Timeout

```go
resp, err := sender.Send(
	ctx, 
	"https://api.example.com/webhook", 
	params,
	webhook.WithRequestTimeout(5 * time.Second),
)
```

### Checking Response Status

```go
resp, err := sender.Send(ctx, "https://api.example.com/webhook", params)
if err != nil {
	log.Fatalf("Failed to send webhook: %v", err)
}

if !resp.IsSuccessful() {
	log.Printf("Webhook request failed with status %d: %s", resp.StatusCode, resp.Body)
} else {
	log.Printf("Webhook sent successfully!")
}
```

## API Reference

### Types

#### `WebhookSender`

```go
type WebhookSender interface {
	Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error)
}
```

#### `Response`

```go
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
	Request    *Request
}

func (r *Response) IsSuccessful() bool
```

### Functions

#### `NewWebhookSender`

```go
func NewWebhookSender(opts ...SenderOption) WebhookSender
```

### Sender Options

- `WithHTTPClient(client *http.Client) SenderOption`
- `WithDefaultTimeout(timeout time.Duration) SenderOption`
- `WithDefaultHeaders(headers map[string]string) SenderOption`
- `WithDefaultMethod(method string) SenderOption`
- `WithMaxRetries(retries int) SenderOption`
- `WithRetryInterval(interval time.Duration) SenderOption`

### Request Options

- `WithMethod(method string) RequestOption`
- `WithHeader(key, value string) RequestOption`
- `WithHeaders(headers map[string]string) RequestOption`
- `WithRequestTimeout(timeout time.Duration) RequestOption`
