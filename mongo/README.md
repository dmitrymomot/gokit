# MongoDB Package

A lightweight wrapper for MongoDB with connection management, type-safe configuration, and built-in health checks.

## Installation

```bash
go get github.com/dmitrymomot/gokit/mongo
```

## Overview

The `mongo` package provides a simple, robust interface to MongoDB in Go applications with a focus on connection management, environment-based configuration, and health monitoring.

## Features

- Type-safe configuration with environment variable support
- Automatic connection management with retry capabilities
- Connection pooling with configurable settings
- Built-in health check functionality
- Support for MongoDB Driver v2
- Simple database and collection access

## Usage

### Basic Connection

```go
import (
    "context"
    "log"
    "github.com/dmitrymomot/gokit/mongo"
)

// Create a MongoDB client with direct configuration
client, err := mongo.New(context.Background(), mongo.Config{
    ConnectionURL: "mongodb://localhost:27017",
})
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer client.Disconnect(context.Background())

// Use the client
collection := client.Database("mydb").Collection("mycollection")
```

### Environment-Based Configuration

```go
import (
    "github.com/dmitrymomot/gokit/config"
    "github.com/dmitrymomot/gokit/mongo"
)

// Load MongoDB config from environment variables
cfg, err := config.Load[mongo.Config]()
if err != nil {
    log.Fatalf("Failed to load config: %v", err)
}

// Create client with loaded config
client, err := mongo.New(context.Background(), cfg)
```

### Connect to a Specific Database

```go
// Direct connection to a specific database
db, err := mongo.NewWithDatabase(context.Background(), cfg, "mydb")
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Use the database
collection := db.Collection("users")
```

### Health Check Integration

```go
import (
    "net/http"
)

// Create a MongoDB client
client, _ := mongo.New(context.Background(), cfg)

// Create a healthcheck function
healthCheck := mongo.Healthcheck(client)

// Use in HTTP handler
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := healthCheck(r.Context()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("MongoDB unavailable"))
        return
    }
    w.Write([]byte("MongoDB healthy"))
})
```

## Configuration

The `Config` struct provides comprehensive options for MongoDB connections:

```go
type Config struct {
    ConnectionURL   string        `env:"MONGODB_URL,required"`
    ConnectTimeout  time.Duration `env:"MONGODB_CONNECT_TIMEOUT" envDefault:"10s"`
    MaxPoolSize     uint64        `env:"MONGODB_MAX_POOL_SIZE" envDefault:"100"`
    MinPoolSize     uint64        `env:"MONGODB_MIN_POOL_SIZE" envDefault:"1"`
    MaxConnIdleTime time.Duration `env:"MONGODB_MAX_CONN_IDLE_TIME" envDefault:"300s"`
    RetryWrites     bool          `env:"MONGODB_RETRY_WRITES" envDefault:"true"`
    RetryReads      bool          `env:"MONGODB_RETRY_READS" envDefault:"true"`
    RetryAttempts   int           `env:"MONGODB_RETRY_ATTEMPTS" envDefault:"3"`
    RetryInterval   time.Duration `env:"MONGODB_RETRY_INTERVAL" envDefault:"5s"`
}
```

## API Reference

### Client Creation

- `New(ctx context.Context, cfg Config) (*mongo.Client, error)`: Create a MongoDB client
- `NewWithDatabase(ctx context.Context, cfg Config, database string) (*mongo.Database, error)`: Create a client and return a specific database

### Health Monitoring

- `Healthcheck(client *mongo.Client) func(context.Context) error`: Create a health check function for MongoDB

### Error Handling

```go
// Check for specific errors
if errors.Is(err, mongo.ErrFailedToConnectToMongo) {
    // Handle connection failure
}

if errors.Is(err, mongo.ErrHealthcheckFailed) {
    // Handle health check failure
}
```

## Best Practices

1. **Always close connections**: Use `defer client.Disconnect(ctx)` to ensure proper cleanup
2. **Configure appropriate timeouts**: Set timeout values based on your network environment
3. **Monitor connection health**: Implement regular health checks in your application
4. **Adjust pool sizes**: Configure connection pooling based on your load requirements
5. **Use context for cancellation**: Pass appropriate context to control operation lifetimes
6. **Handle transient failures**: Configure retry options for resilience