# Redis Package

A lightweight Redis client wrapper with connection management, healthchecks, and Fiber storage implementation.

## Installation

```bash
go get github.com/dmitrymomot/gokit/redis
```

## Overview

The `redis` package provides a simplified, production-ready Redis client interface with robust connection management, environment-based configuration, health monitoring, and a ready-to-use Fiber storage implementation.

## Features

- Type-safe configuration with environment variable support
- Automatic connection retry with configurable attempts and intervals
- Built-in health check function for service monitoring
- Context-aware operations with proper timeout handling
- Full Fiber storage interface implementation for sessions and caching
- Clear error types for improved error handling

## Usage

### Basic Connection

```go
import (
    "context"
    "github.com/dmitrymomot/gokit/redis"
)

func main() {
    // Create a context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()
    
    // Connect with configuration
    client, err := redis.Connect(ctx, redis.Config{
        ConnectionURL: "redis://localhost:6379/0",
        RetryAttempts: 3,
        RetryInterval: 5 * time.Second,
        ConnectTimeout: 30 * time.Second,
    })
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Use the Redis client
    err = client.Set(ctx, "key", "value", time.Hour).Err()
    // ...
}
```

### Loading Config from Environment

```go
import (
    "context"
    "github.com/dmitrymomot/gokit/config"
    "github.com/dmitrymomot/gokit/redis"
)

func main() {
    // Load from environment variables
    cfg, err := config.Load[redis.Config]()
    if err != nil {
        panic(err)
    }
    
    // Connect using the loaded config
    client, err := redis.Connect(context.Background(), cfg)
    if err != nil {
        panic(err)
    }
    defer client.Close()
}
```

### Health Checking

```go
import (
    "context"
    "net/http"
    "github.com/dmitrymomot/gokit/redis"
)

func setupHealthCheck(client *redis.Client) {
    // Create a health check function
    healthCheck := redis.Healthcheck(client)
    
    // Use in HTTP health endpoint
    http.HandleFunc("/health/redis", func(w http.ResponseWriter, r *http.Request) {
        if err := healthCheck(r.Context()); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte("Redis unhealthy: " + err.Error()))
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Redis healthy"))
    })
}
```

### Using with Fiber

```go
import (
    "context"
    "github.com/dmitrymomot/gokit/redis"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
)

func setupFiberApp(ctx context.Context) *fiber.App {
    // Connect to Redis
    client, err := redis.Connect(ctx, redis.Config{
        ConnectionURL: "redis://localhost:6379/0",
    })
    if err != nil {
        panic(err)
    }
    
    // Create Redis storage for Fiber
    storage := redis.NewStorage(client)
    
    // Create session store with Redis storage
    store := session.New(session.Config{
        Storage: storage,
    })
    
    // Create Fiber app
    app := fiber.New()
    
    // Example route with session
    app.Get("/", func(c *fiber.Ctx) error {
        sess, err := store.Get(c)
        if err != nil {
            return err
        }
        
        // Get or set session values
        visits := sess.Get("visits")
        if visits == nil {
            sess.Set("visits", 1)
        } else {
            sess.Set("visits", visits.(int)+1)
        }
        
        if err := sess.Save(); err != nil {
            return err
        }
        
        return c.SendString(fmt.Sprintf("You have visited %d times", sess.Get("visits")))
    })
    
    return app
}
```

## Configuration

The `Config` struct supports the following fields, all configurable via environment variables:

| Field | Env Variable | Default | Description |
|-------|--------------|---------|-------------|
| `ConnectionURL` | `REDIS_URL` | `redis://localhost:6379/0` | Redis connection URL |
| `RetryAttempts` | `REDIS_RETRY_ATTEMPTS` | `3` | Number of connection retry attempts |
| `RetryInterval` | `REDIS_RETRY_INTERVAL` | `5s` | Interval between retry attempts |
| `ConnectTimeout` | `REDIS_CONNECT_TIMEOUT` | `30s` | Connection timeout |

## API Reference

### Connection Management

```go
// Connect establishes a connection to Redis with retry logic
func Connect(ctx context.Context, cfg Config) (*redis.Client, error)

// Healthcheck creates a health check function for the Redis connection
func Healthcheck(client redis.UniversalClient) func(context.Context) error
```

### Storage Interface

```go
// Create a new storage instance
storage := redis.NewStorage(client)

// Available methods
storage.Get(key string) ([]byte, error)
storage.Set(key string, val []byte, exp time.Duration) error
storage.Delete(key string) error
storage.Reset() error
storage.Close() error
storage.Conn() redis.UniversalClient
storage.Keys() ([][]byte, error)
```

## Error Handling

The package defines several error types for specific error conditions:

```go
if errors.Is(err, redis.ErrRedisNotReady) {
    // Handle Redis not ready error
}

if errors.Is(err, redis.ErrFailedToParseRedisConnString) {
    // Handle invalid connection string error
}

if errors.Is(err, redis.ErrHealthcheckFailed) {
    // Handle healthcheck failure
}
```

## Best Practices

1. **Always use context with timeouts** for Redis operations to prevent blocked goroutines
2. **Close Redis clients** when they're no longer needed to prevent resource leaks
3. **Implement health checks** in your service readiness/liveness probes
4. **Use environment-based configuration** for different deployment environments
5. **Consider connection pooling** for high-throughput applications
6. **Add error handling** for Redis operations, particularly for distributed systems
7. **Set appropriate timeouts** based on your application's needs and network conditions
