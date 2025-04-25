# OpenSearch Package

A lightweight wrapper for the official OpenSearch Go client with type-safe configuration.

## Installation

```bash
go get github.com/dmitrymomot/gokit/opensearch
```

## Overview

The `opensearch` package provides a simple, type-safe interface to OpenSearch with environment-based configuration, automatic health checking, and standardized error handling.

## Features

- Type-safe configuration with environment variable support
- Built-in health check on connection
- Comprehensive error handling with specific error types
- Simplified client initialization

## Usage

### Basic Connection

```go
import (
    "context"
    "github.com/dmitrymomot/gokit/opensearch"
)

// Create an OpenSearch client with direct configuration
client, err := opensearch.New(context.Background(), opensearch.Config{
    Addresses: []string{"https://localhost:9200"},
    Username:  "admin",
    Password:  "admin",
})
if err != nil {
    // Handle error
}

// Use the client
info, err := client.Info()
```

### Environment-Based Configuration

```go
import (
    "github.com/dmitrymomot/gokit/config"
    "github.com/dmitrymomot/gokit/opensearch"
)

// Load OpenSearch config from environment variables
cfg, err := config.Load[opensearch.Config]()
if err != nil {
    // Handle error
}

// Create client with loaded config
client, err := opensearch.New(context.Background(), cfg)
```

### Health Check Integration

```go
import (
    "net/http"
)

// Create a healthcheck function for an existing client
healthCheck := opensearch.Healthcheck(client)

// Use in HTTP handler
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := healthCheck(r.Context()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("OpenSearch unavailable"))
        return
    }
    w.Write([]byte("OpenSearch healthy"))
})
```

## Configuration

The `Config` struct provides options for OpenSearch connections with environment variable integration:

```go
type Config struct {
    Addresses    []string `env:"OPENSEARCH_ADDRESSES,required"`
    Username     string   `env:"OPENSEARCH_USERNAME,notEmpty"`
    Password     string   `env:"OPENSEARCH_PASSWORD,notEmpty"`
    MaxRetries   int      `env:"OPENSEARCH_MAX_RETRIES" default:"3"`
    DisableRetry bool     `env:"OPENSEARCH_DISABLE_RETRY" default:"false"`
}
```

## API Reference

### Client Creation

- `New(ctx context.Context, cfg Config) (*opensearch.Client, error)`: Create an OpenSearch client with built-in health check

### Health Monitoring

- `Healthcheck(client *opensearch.Client) func(context.Context) error`: Create a health check function for OpenSearch

### Error Handling

```go
// Check for specific errors
if errors.Is(err, opensearch.ErrConnectionFailed) {
    // Handle connection failure
}

if errors.Is(err, opensearch.ErrHealthcheckFailed) {
    // Handle health check failure
}
```

## Best Practices

1. **Secure credential storage**: Store OpenSearch credentials in environment variables
2. **Implement health checks**: Use the provided health check functionality in your application
3. **Configure retries appropriately**: Adjust MaxRetries based on your network reliability
4. **Handle errors specifically**: Use the provided error types for better error handling
5. **Use context for cancellation**: Pass appropriate context to control operation lifetimes
