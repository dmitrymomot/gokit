# Cache Package

A generic, multi-level caching system with Redis and LRU implementations.

## Installation

```bash
go get github.com/dmitrymomot/gokit/cache
```

## Overview

The `cache` package provides a unified caching interface with multiple implementations, including Redis, in-memory LRU, and a layered combination of both. It offers a consistent API for basic cache operations like get, set, delete, and flush across all implementations. All operations are context-aware and thread-safe, making it suitable for concurrent applications.

## Features

- Generic `Cache` interface for consistent API across implementations
- Redis adapter using `github.com/redis/go-redis/v9`
- In-memory LRU adapter using `github.com/hashicorp/golang-lru/v2`
- Layered cache for multi-level caching (LRU + Redis)
- Context support for all operations
- Thread-safe implementations
- Comprehensive error handling with specific error types
- Concurrent operations in layered cache for better performance

## Usage

### Basic Interface

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
)

// The Cache interface is implemented by all adapters
var cacher cache.Cache // This will be one of the implementations
ctx := context.Background()

// Basic operations
err := cacher.Set(ctx, "mykey", []byte("value"), time.Hour)
value, found, err := cacher.Get(ctx, "mykey")
// found is true if key exists, false otherwise
// err will be non-nil if an error occurred

// Check if key exists
exists, err := cacher.Exists(ctx, "mykey")

// Delete a key
deleted, err := cacher.Delete(ctx, "mykey")
// deleted is true if the key was deleted, false if it didn't exist

// Clear all keys
err = cacher.Flush(ctx)
```

### Redis Cache

```go
import (
    "context"
    "time"
    "fmt"
    
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
    switch {
    case errors.Is(err, cache.ErrConnectionFailed):
        // Handle connection failure
    default:
        // Handle other errors
    }
}
defer redisCache.Close()

ctx := context.Background()

// Set with expiration
err = redisCache.Set(ctx, "user:123", []byte(`{"name":"John"}`), time.Hour)
if err != nil {
    // Handle error
}

// Get value
value, found, err := redisCache.Get(ctx, "user:123")
if err != nil {
    // Handle error
}
if found {
    fmt.Printf("Found user: %s\n", value)
    // value = []byte(`{"name":"John"}`)
}
```

### LRU Cache

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
)

// Create an in-memory LRU cache with capacity of 1000 items
lruCache, err := cache.NewLRUAdapter(1000)
if err != nil {
    // Handle error
}

ctx := context.Background()

// Cache session data with 5 minute expiration
err = lruCache.Set(ctx, "session:xyz", []byte("session-data"), 5*time.Minute)
if err != nil {
    // Handle error
}

// Retrieve cached data
value, found, err := lruCache.Get(ctx, "session:xyz")
if err != nil {
    // Handle error
}
if found {
    // value = []byte("session-data")
}
```

### Layered Cache

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
    "github.com/redis/go-redis/v9"
)

// Create LRU cache (L1 - fast local memory)
lruCache, err := cache.NewLRUAdapter(10000)
if err != nil {
    // Handle error
}

// Create Redis cache (L2 - persistent distributed cache)
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

ctx := context.Background()

// Store product data for one hour
err = layeredCache.Set(ctx, "product:123", []byte(`{"name":"Widget"}`), time.Hour)
if err != nil {
    // Handle error
}

// Get checks L1 first, then L2 if not found in L1
// If found in L2 but not L1, it's backfilled to L1 automatically
value, found, err := layeredCache.Get(ctx, "product:123")
if err != nil {
    // Handle error
}
if found {
    // value = []byte(`{"name":"Widget"}`)
}
```

## Error Handling

```go
import (
    "context"
    "errors"
    "fmt"
    
    "github.com/dmitrymomot/gokit/cache"
)

ctx := context.Background()

// Use your cache implementation
var c cache.Cache

_, _, err := c.Get(ctx, "some-key")
if err != nil {
    switch {
    case errors.Is(err, cache.ErrNotFound):
        // Handle key not found
    case errors.Is(err, cache.ErrConnectionFailed):
        // Handle connection failure
    case errors.Is(err, cache.ErrDecoding):
        // Handle data corruption
    default:
        // Handle other errors
    }
}
```

## Best Practices

1. **TTL Management**:
   - Use appropriate TTL values based on data volatility
   - Consider shorter TTLs for L1 cache in LayeredCache to save memory

2. **Error Handling**:
   - Always check for errors, especially for distributed caches like Redis
   - Implement graceful degradation when cache services are unavailable

3. **Cache Invalidation**:
   - Plan for invalidation strategies (time-based, event-based)
   - Consider using cache namespaces for easier bulk invalidation

4. **Concurrency**:
   - All implementations are thread-safe but consider your access patterns
   - Monitor connection pools for Redis adapter under high load

5. **Data Size**:
   - Be mindful of memory usage when caching large objects
   - Consider compression for large values if needed

## API Reference

### Core Interface

```go
type Cache interface {
    // Get retrieves an item from the cache.
    // It returns the item (as []byte) and true if found, otherwise nil and false.
    Get(ctx context.Context, key string) ([]byte, bool, error)
    
    // Set adds an item to the cache with an optional expiration duration.
    // If ttl is 0, the item will not expire (if supported by the backend).
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    
    // Delete removes an item from the cache.
    // It returns true if the item was deleted, false if it didn't exist.
    Delete(ctx context.Context, key string) (bool, error)
    
    // Exists checks if an item exists in the cache.
    Exists(ctx context.Context, key string) (bool, error)
    
    // Flush removes all items from the cache (if supported).
    Flush(ctx context.Context) error
    
    // Close closes the cache connection or releases resources (if applicable).
    Close() error
}
```

### Functions

```go
func NewRedisAdapter(client redis.UniversalClient) (*RedisAdapter, error)
```
Creates a new Redis cache adapter from a redis.UniversalClient.

```go
func NewLRUAdapter(size int) (*LRUAdapter, error)
```
Creates a new in-memory LRU cache with the specified maximum number of items.

```go
func NewLayeredCache(l1, l2 Cache) (*LayeredCache, error)
```
Creates a new layered cache with L1 (usually LRU) and L2 (usually Redis) cache layers.

### Error Types

```go
var ErrNotFound = errors.New("cache: key not found")
var ErrEncoding = errors.New("cache: failed to encode value")
var ErrDecoding = errors.New("cache: failed to decode value")
var ErrConnectionFailed = errors.New("cache: connection failed")
var ErrOperationFailed = errors.New("cache: operation failed")
var ErrNotImplemented = errors.New("cache: feature not implemented")
