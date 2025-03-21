# Redis Package

A Go package that provides a simple and configurable Redis client with automatic connection retries and environment variable support.

## Installation

```bash
go get github.com/dmitrymomot/gokit/redis
```

## Quick Start

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

## Configuration

The package uses environment variables for configuration and supports `.env` files through `godotenv/autoload`.

### Environment Variables

The following environment variables are used to configure the Redis client:

| Environment Variable    | Description                 | Default                    | Required |
| ----------------------- | --------------------------- | -------------------------- | -------- |
| `REDIS_URL`             | Redis connection URL        | "redis://localhost:6379/0" | Yes      |
| `REDIS_RETRY_ATTEMPTS`  | Connection retry attempts   | 3                          | No       |
| `REDIS_RETRY_INTERVAL`  | Interval between retries    | "5s"                       | No       |
| `REDIS_CONNECT_TIMEOUT` | Connection timeout duration | "30s"                      | No       |

### Connection URL Format

The Redis connection URL should be in the following format:

```
redis://:password@localhost:6379/0
```

## Error Handling

The package provides several error types for better error handling:

- `ErrEmptyConnectionURL`: Redis connection URL is empty
- `ErrFailedToParseRedisConnString`: Invalid Redis connection string
- `ErrRedisNotReady`: Redis server not ready within timeout period
- `ErrHealthcheckFailed`: Redis healthcheck failed

## Features

- Automatic connection retries
- Environment variable configuration
- Built-in healthcheck functionality
- Connection timeout handling
- `.env` file support

## Example Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/dmitrymomot/gokit/redis"
)

func main() {
    // Create a new Redis client
    client, err := redis.New()
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Use the client
    ctx := context.Background()
    err = client.Set(ctx, "key", "value", time.Hour).Err()
    if err != nil {
        log.Fatal(err)
    }

    val, err := client.Get(ctx, "key").Result()
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Value: %s", val)

    // Using the healthcheck
    healthcheck := redis.Healthcheck(client)
    if err := healthcheck(ctx); err != nil {
        log.Fatal("Redis healthcheck failed:", err)
    }
}
```

## Dependencies

- [github.com/redis/go-redis/v9](https://github.com/redis/go-redis)
- [github.com/caarlos0/env/v11](https://github.com/caarlos0/env)
- [github.com/joho/godotenv](https://github.com/joho/godotenv)
