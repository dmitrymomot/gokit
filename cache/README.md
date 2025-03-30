# Cache Package

This package provides a generic interface and implementations for caching in Go applications.

## Features

- Generic `Cache` interface with standard operations
- Redis adapter implementation using `github.com/redis/go-redis/v9`
- In-memory LRU cache adapter using `github.com/hashicorp/golang-lru/v2`
- Layered cache combining LRU and Redis for multi-level caching
- Consistent error handling with predefined error types
- Thread-safe implementations
- Context support for all operations

## Installation

```bash
go get github.com/dmitrymomot/gokit/cache
go get github.com/redis/go-redis/v9 # For Redis adapter
go get github.com/hashicorp/golang-lru/v2 # For LRU adapter
```

## Usage

### Cache Interface

The package defines a common `Cache` interface that all implementations adhere to:

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

### Redis Cache Adapter

The Redis adapter provides a Redis-backed implementation of the `Cache` interface.

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
    "github.com/redis/go-redis/v9"
)

func Example_redisCache() {
    // Create a Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Create a Redis cache adapter
    redisCache, err := cache.NewRedisAdapter(redisClient)
    if err != nil {
        // Handle connection error
        panic(err)
    }
    defer redisCache.Close()
    
    ctx := context.Background()
    
    // Set a value with a 1 hour TTL
    err = redisCache.Set(ctx, "user:123", []byte(`{"name":"John","age":30}`), time.Hour)
    if err != nil {
        // Handle error
    }
    
    // Get a value
    value, found, err := redisCache.Get(ctx, "user:123")
    if err != nil {
        // Handle error
    }
    if found {
        // Use value ([]byte)
    }
    
    // Check if a key exists
    exists, err := redisCache.Exists(ctx, "user:123")
    if err != nil {
        // Handle error
    }
    
    // Delete a key
    deleted, err := redisCache.Delete(ctx, "user:123")
    if err != nil {
        // Handle error
    }
    
    // Flush all keys (use with caution)
    err = redisCache.Flush(ctx)
    if err != nil {
        // Handle error
    }
}
```

### LRU Cache Adapter

The LRU adapter provides an in-memory implementation of the `Cache` interface using a Least Recently Used (LRU) eviction policy.

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
)

func Example_lruCache() {
    // Create an LRU cache with a capacity of 1000 items
    lruCache, err := cache.NewLRUAdapter(1000)
    if err != nil {
        // Handle error
        panic(err)
    }
    
    ctx := context.Background()
    
    // Set a value with a 5 minute TTL
    err = lruCache.Set(ctx, "session:xyz", []byte("session-data"), 5*time.Minute)
    if err != nil {
        // Handle error
    }
    
    // Get a value
    value, found, err := lruCache.Get(ctx, "session:xyz")
    if err != nil {
        // Handle error
    }
    if found {
        // Use value ([]byte)
    }
    
    // Delete a key
    deleted, err := lruCache.Delete(ctx, "session:xyz")
    if err != nil {
        // Handle error
    }
    
    // LRU cache automatically evicts least recently used items when capacity is reached
}
```

### Layered Cache

The layered cache combines an LRU cache (L1) and a Redis cache (L2) to provide a fast, multi-level caching solution.

```go
import (
    "context"
    "time"
    
    "github.com/dmitrymomot/gokit/cache"
    "github.com/redis/go-redis/v9"
)

func Example_layeredCache() {
    // Create an LRU cache as the first layer (L1)
    lruCache, err := cache.NewLRUAdapter(10000)
    if err != nil {
        panic(err)
    }
    
    // Create a Redis cache as the second layer (L2)
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    redisCache, err := cache.NewRedisAdapter(redisClient)
    if err != nil {
        panic(err)
    }
    
    // Create a layered cache
    layeredCache, err := cache.NewLayeredCache(lruCache, redisCache)
    if err != nil {
        panic(err)
    }
    defer layeredCache.Close()
    
    ctx := context.Background()
    
    // Set a value (writes to both L1 and L2 concurrently)
    err = layeredCache.Set(ctx, "product:456", []byte(`{"id":"456","name":"Widget"}`), time.Hour)
    if err != nil {
        // Handle error
    }
    
    // Get a value (checks L1 first, then L2 if not found in L1)
    // If found in L2 but not L1, it's automatically added to L1
    value, found, err := layeredCache.Get(ctx, "product:456")
    if err != nil {
        // Handle error
    }
    
    // Delete a value (removes from both L1 and L2 concurrently)
    deleted, err := layeredCache.Delete(ctx, "product:456")
    if err != nil {
        // Handle error
    }
}
```

## Error Handling

The package defines standard errors that all implementations use:

```go
var (
    // ErrNotFound indicates that the requested key was not found in the cache.
    ErrNotFound = errors.New("cache: key not found")

    // ErrEncoding indicates an error during encoding the cache value.
    ErrEncoding = errors.New("cache: failed to encode value")

    // ErrDecoding indicates an error during decoding the cached value.
    ErrDecoding = errors.New("cache: failed to decode value")

    // ErrConnectionFailed indicates a failure to connect to the cache backend.
    ErrConnectionFailed = errors.New("cache: connection failed") 

    // ErrOperationFailed indicates a general failure during a cache operation.
    ErrOperationFailed = errors.New("cache: operation failed")

    // ErrNotImplemented indicates a feature is not implemented by the specific cache adapter.
    ErrNotImplemented = errors.New("cache: feature not implemented")
)
```

## Implementation Details

### Redis Adapter

- Wraps any `redis.UniversalClient` (single client, cluster client, etc.)
- Performs connectivity check during initialization
- Handles Redis-specific errors and translates them to package errors
- Safe for concurrent use

### LRU Adapter

- In-memory cache with LRU eviction policy
- Handles item expiration via time-based mechanism
- Thread-safe with read/write locks
- Zero external dependencies at runtime
- Efficient resource usage

### Layered Cache

- Combines two cache implementations: primary (L1) and secondary (L2)
- Read-through and write-through caching strategy
- Automatic LRU cache population from Redis on cache misses
- Parallel operations using Go's concurrency primitives
- Robust error handling with error aggregation
