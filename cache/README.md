# Cache Package

A generic, multi-level caching system with Redis and LRU implementations.

## Installation

```bash
go get github.com/dmitrymomot/gokit/cache
```

## Overview

The `cache` package provides a unified caching interface with multiple implementations, including Redis, in-memory LRU, and a layered combination of both. It supports operations like get, set, delete with a consistent API across all implementations.

## Features

- Generic `Cache` interface for consistent API
- Redis adapter using `github.com/redis/go-redis/v9`
- In-memory LRU adapter using `github.com/hashicorp/golang-lru/v2`
- Layered cache for multi-level caching (LRU + Redis)
- Context support for all operations
- Comprehensive error handling
- Thread-safe implementations

## Usage

### Cache Interface

```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) (bool, error)
    Exists(ctx context.Context, key string) (bool, error)
    Flush(ctx context.Context) error
    Close() error
}
```

### Redis Cache

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
    "github.com/redis/go-redis/v9"
)

// Create a Redis client
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Create a Redis cache adapter
redisCache, err := cache.NewRedisAdapter(redisClient)
if err != nil {
    // Handle error
}
defer redisCache.Close()

ctx := context.Background()

// Use the cache
err = redisCache.Set(ctx, "user:123", []byte(`{"name":"John"}`), time.Hour)
value, found, err := redisCache.Get(ctx, "user:123")
```

### LRU Cache

```go
// Create an in-memory LRU cache with capacity of 1000 items
lruCache, err := cache.NewLRUAdapter(1000)
if err != nil {
    // Handle error
}

// Use the cache (same API as Redis cache)
err = lruCache.Set(ctx, "session:xyz", []byte("session-data"), 5*time.Minute)
value, found, err := lruCache.Get(ctx, "session:xyz")
```

### Layered Cache

```go
// Create LRU cache (L1)
lruCache, err := cache.NewLRUAdapter(10000)
if err != nil {
    // Handle error
}

// Create Redis cache (L2)
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})
redisCache, err := cache.NewRedisAdapter(redisClient)
if err != nil {
    // Handle error
}

// Create layered cache
layeredCache, err := cache.NewLayeredCache(lruCache, redisCache)
if err != nil {
    // Handle error
}
defer layeredCache.Close()

// Use the layered cache (same API)
// Gets check L1 first, then L2 if not found
// Sets write to both L1 and L2 concurrently
err = layeredCache.Set(ctx, "product:123", []byte(`{"name":"Widget"}`), time.Hour)
value, found, err := layeredCache.Get(ctx, "product:123")
```

## Error Handling

The package defines standard errors for consistent error handling:

```go
var (
    // Key not found in the cache
    ErrNotFound = errors.New("cache: key not found")
    
    // Error during value encoding/decoding
    ErrEncoding = errors.New("cache: failed to encode value")
    ErrDecoding = errors.New("cache: failed to decode value")
    
    // Connection or operation failures
    ErrConnectionFailed = errors.New("cache: connection failed")
    ErrOperationFailed = errors.New("cache: operation failed")
    
    // Feature not implemented
    ErrNotImplemented = errors.New("cache: feature not implemented")
)
```

## Implementation Details

- **Redis Adapter**: Thread-safe wrapper for `redis.UniversalClient`
- **LRU Adapter**: In-memory cache with LRU eviction and TTL support
- **Layered Cache**: Combines L1 (fast, limited) and L2 (slower, persistent) caches
