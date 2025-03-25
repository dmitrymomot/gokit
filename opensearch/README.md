# OpenSearch

A lightweight wrapper for the official OpenSearch Go client that provides:

- Simple configuration with environment variable binding
- Client creation with automatic healthcheck
- Error handling and custom error types

## Installation

```go
go get github.com/dmitrymomot/gokit/opensearch
```

## Usage

### Basic Setup

```go
package main

import (
	"context"
	"log"

	"github.com/dmitrymomot/gokit/opensearch"
)

func main() {
	// Create configuration
	cfg := opensearch.Config{
		Addresses:    []string{"https://localhost:9200"},
		Username:     "admin",
		Password:     "admin",
		MaxRetries:   3,
		DisableRetry: false,
	}

	// Create a new client
	ctx := context.Background()
	client, err := opensearch.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create OpenSearch client: %v", err)
	}

	// Use the client for OpenSearch operations
	info, err := client.Info()
	if err != nil {
		log.Fatalf("Failed to get cluster info: %v", err)
	}
	log.Printf("Connected to OpenSearch cluster: %s", info.ClusterName)
}
```

### Environment Variables

The configuration can be loaded from environment variables:

```
OPENSEARCH_ADDRESSES=https://localhost:9200
OPENSEARCH_USERNAME=admin
OPENSEARCH_PASSWORD=admin
OPENSEARCH_MAX_RETRIES=3
OPENSEARCH_DISABLE_RETRY=false
```

### Healthcheck

The built-in healthcheck function can be used independently:

```go
// Run a healthcheck on an existing client
err := opensearch.Healthcheck(client)(ctx)
if err != nil {
    log.Fatalf("OpenSearch cluster is not healthy: %v", err)
}
```

## API Reference

### Config Struct

```go
type Config struct {
	Addresses    []string // OpenSearch node addresses
	Username     string   // Basic auth username
	Password     string   // Basic auth password
	MaxRetries   int      // Maximum number of retries
	DisableRetry bool     // Disable retry mechanism
}
```

### Functions

- `New(ctx context.Context, cfg Config) (*opensearch.Client, error)`: Creates a new OpenSearch client with the provided configuration and performs a health check
- `Healthcheck(client *opensearch.Client) func(context.Context) error`: Returns a function that checks the health of the OpenSearch cluster
