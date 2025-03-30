package cache

import (
	"context"
	"time"
)

// Cache defines the interface for a generic cache store.
// Implementations can include in-memory caches, Redis, Memcached, etc.
type Cache interface {
	// Get retrieves an item from the cache.
	// It returns the item (as []byte) and true if found, otherwise nil and false.
	// Errors during retrieval will result in nil, false, and the error.
	Get(ctx context.Context, key string) ([]byte, bool, error)

	// Set adds an item to the cache with an optional expiration duration.
	// If ttl is 0, the item will not expire (if supported by the backend).
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes an item from the cache.
	// It returns true if the item was deleted, false if it didn't exist.
	// Errors during deletion will result in false and the error.
	Delete(ctx context.Context, key string) (bool, error)

	// Exists checks if an item exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)

	// Flush removes all items from the cache (if supported).
	Flush(ctx context.Context) error

	// Close closes the cache connection or releases resources (if applicable).
	Close() error
}
