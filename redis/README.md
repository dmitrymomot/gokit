# Redis Package

A Go package that provides a simple and configurable Redis client with connection helpers, healthcheck functionality, and a Fiber storage interface implementation.

## Installation

```bash
go get github.com/dmitrymomot/gokit/redis
```

## Primary Features

### Redis Connection

This package provides simplified connection helpers for Redis with automatic retries and robust configuration:

```go
import "github.com/dmitrymomot/gokit/redis"

func main() {
    // Create a new Redis client
    client, err := redis.Connect(config.Redis{
        ConnectionURL: "redis://localhost:6379/0",
    })
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Your Redis operations here...
}
```

### Healthcheck Functionality

The package includes a built-in healthcheck function that can be used for service readiness probes:

```go
// Using the healthcheck
healthcheck := redis.Healthcheck(client)
if err := healthcheck(ctx); err != nil {
    log.Fatal("Redis healthcheck failed:", err)
}
```

### Configuration

The package uses environment variables for configuration and supports `.env` files through `godotenv/autoload`.

#### Environment Variables

The following environment variables configure the Redis client:

| Environment Variable    | Description                 | Default                    | Required |
| ----------------------- | --------------------------- | -------------------------- | -------- |
| `REDIS_URL`             | Redis connection URL        | "redis://localhost:6379/0" | Yes      |
| `REDIS_RETRY_ATTEMPTS`  | Connection retry attempts   | 3                          | No       |
| `REDIS_RETRY_INTERVAL`  | Interval between retries    | "5s"                       | No       |
| `REDIS_CONNECT_TIMEOUT` | Connection timeout duration | "30s"                      | No       |

## Fiber Storage Interface

The package also provides a Redis-based storage implementation that satisfies the Fiber storage interface. This makes it compatible with various Fiber features that require storage capabilities.

### Storage Usage

```go
import (
    "github.com/dmitrymomot/gokit/redis"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
)

func main() {
    // Create a Redis client
    client, err := redis.New()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Create a Redis-based storage for Fiber
    storage := redis.NewStorage(client)
    
    // Example: Use with session middleware
    store := session.New(session.Config{
        Storage: storage,
    })
    
    app := fiber.New()
    
    // Use the session middleware as an example
    app.Use(func(c *fiber.Ctx) error {
        sess, err := store.Get(c)
        if err != nil {
            return err
        }
        
        // Set a value
        sess.Set("user", "john")
        
        // Save the session
        if err := sess.Save(); err != nil {
            return err
        }
        
        return c.Next()
    })
    
    // Start server
    app.Listen(":3000")
}
```

### Storage Methods

The Redis storage implementation provides the following methods that fulfill the Fiber storage interface:

- `Get(key string) ([]byte, error)`: Retrieve a value by key
- `Set(key string, val []byte, exp time.Duration) error`: Store a value with optional expiration
- `Delete(key string) error`: Remove a key-value pair
- `Reset() error`: Clear all stored data
- `Close() error`: Close the Redis connection

Additionally, it provides these helper methods:

- `Conn() redis.UniversalClient`: Access the underlying Redis client
- `Keys() ([][]byte, error)`: Get all keys in the database

## Error Handling

The package provides several error types for better error handling:

- `ErrEmptyConnectionURL`: Redis connection URL is empty
- `ErrFailedToParseRedisConnString`: Invalid Redis connection string
- `ErrRedisNotReady`: Redis server not ready within timeout period
- `ErrHealthcheckFailed`: Redis healthcheck failed

## Dependencies

- [github.com/redis/go-redis/v9](https://github.com/redis/go-redis)
- [github.com/caarlos0/env/v11](https://github.com/caarlos0/env)
- [github.com/joho/godotenv](https://github.com/joho/godotenv)
