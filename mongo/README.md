# MongoDB Package

A lightweight wrapper around the MongoDB Go driver that simplifies connection management, configuration, and health checks for MongoDB in Go applications.

## Overview

This package provides:

- Simple connection management with retry capabilities
- Environment-based configuration
- Health check functionality for service monitoring
- Helper functions for common database operations

## Installation

```bash
go get github.com/dmitrymomot/gokit/mongo
```

Requires Go 1.24 or higher, and uses MongoDB Go Driver v2.

## Configuration

The package uses a `Config` struct that can be populated from environment variables:

```go
type Config struct {
    ConnectionURL   string        // MongoDB connection URI (env: MONGODB_URL)
    ConnectTimeout  time.Duration // Connection timeout (env: MONGODB_CONNECT_TIMEOUT, default: 10s)
    MaxPoolSize     uint64        // Maximum connections in the pool (env: MONGODB_MAX_POOL_SIZE, default: 100)
    MinPoolSize     uint64        // Minimum connections in the pool (env: MONGODB_MIN_POOL_SIZE, default: 1)
    MaxConnIdleTime time.Duration // Maximum idle connection time (env: MONGODB_MAX_CONN_IDLE_TIME, default: 300s)
    RetryWrites     bool          // Enable write operation retries (env: MONGODB_RETRY_WRITES, default: true)
    RetryReads      bool          // Enable read operation retries (env: MONGODB_RETRY_READS, default: true)
    RetryAttempts   int           // Connection retry attempts (env: MONGODB_RETRY_ATTEMPTS, default: 3)
    RetryInterval   time.Duration // Interval between retries (env: MONGODB_RETRY_INTERVAL, default: 5s)
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGODB_URL` | MongoDB connection string | *Required* |
| `MONGODB_CONNECT_TIMEOUT` | Connection timeout | `10s` |
| `MONGODB_MAX_POOL_SIZE` | Maximum connections | `100` |
| `MONGODB_MIN_POOL_SIZE` | Minimum connections | `1` |
| `MONGODB_MAX_CONN_IDLE_TIME` | Max idle time | `300s` |
| `MONGODB_RETRY_WRITES` | Retry write operations | `true` |
| `MONGODB_RETRY_READS` | Retry read operations | `true` |
| `MONGODB_RETRY_ATTEMPTS` | Connection retry attempts | `3` |
| `MONGODB_RETRY_INTERVAL` | Interval between retries | `5s` |

## Usage

### Creating a MongoDB Client

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/dmitrymomot/gokit/mongo"
)

func main() {
    ctx := context.Background()
    
    config := mongo.Config{
        ConnectionURL:   "mongodb://localhost:27017",
        ConnectTimeout:  10 * time.Second,
        MaxPoolSize:     100,
        MinPoolSize:     1,
        MaxConnIdleTime: 5 * time.Minute,
        RetryWrites:     true,
        RetryReads:      true,
        RetryAttempts:   3,
        RetryInterval:   5 * time.Second,
    }
    
    // Create a new MongoDB client
    client, err := mongo.New(ctx, config)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    
    // Use the client
    // ...
    
    // Remember to close the connection when done
    defer client.Disconnect(ctx)
}
```

### Connecting to a Specific Database

```go
// Connect to a specific database
db, err := mongo.NewWithDatabase(ctx, config, "my_database")
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Use the database
collection := db.Collection("my_collection")
// ...
```

### Health Check Implementation

```go
// Create a health check function
healthCheckFn := mongo.Healthcheck(client)

// Use in a health check endpoint
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    err := healthCheckFn(r.Context())
    if err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("MongoDB is not available"))
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("MongoDB is healthy"))
}
```

## Best Practices

1. **Always close connections**: Use `defer client.Disconnect(ctx)` to ensure connections are properly closed
2. **Use appropriate timeouts**: Set reasonable timeouts for your use case
3. **Configure pool size**: Adjust connection pool sizes based on your application's requirements
4. **Monitor connection health**: Implement regular health checks using the provided `Healthcheck` function
5. **Handle errors**: Always check and handle errors returned by the MongoDB functions

## Known Issues and Limitations

- The package currently uses MongoDB driver v2, which may not be stable in all environments
- The retry logic in connection function needs care when modifying