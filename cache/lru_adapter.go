package cache

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

// lruItem stores the value and its expiration time.
type lruItem struct {
	value    []byte
	expiryAt time.Time // Zero time means no expiration
}

// LRUAdapter implements the Cache interface using an in-memory LRU cache.
type LRUAdapter struct {
	cache *lru.Cache[string, lruItem]
	mu    sync.RWMutex // Protects access to the cache
}

// NewLRUAdapter creates a new in-memory LRU cache adapter.
// size: the maximum number of items the cache can hold.
func NewLRUAdapter(size int) (*LRUAdapter, error) {
	// Initialize the LRU cache
	// We don't need EvictedFunc, OnEvict, or DiscardOldest for basic LRU
	c, err := lru.New[string, lruItem](size)
	if err != nil {
		return nil, err // Should not happen with valid size > 0
	}

	return &LRUAdapter{
		cache: c,
	}, nil
}

// Get retrieves an item from the LRU cache, checking for expiration.
func (l *LRUAdapter) Get(ctx context.Context, key string) ([]byte, bool, error) {
	_ = ctx // Context is not used for in-memory operations but part of the interface

	// First try with read lock
	l.mu.RLock()
	item, found := l.cache.Get(key)
	if !found {
		l.mu.RUnlock()
		return nil, false, nil // Cache miss
	}

	// Check for expiration while holding read lock
	if !item.expiryAt.IsZero() && time.Now().After(item.expiryAt) {
		// We found an expired item - switch to write lock to remove it
		l.mu.RUnlock() // Release read lock before acquiring write lock

		// Acquire write lock to remove expired item
		l.mu.Lock()
		// Need to check again after acquiring write lock to avoid race conditions
		if tempItem, tempFound := l.cache.Peek(key); tempFound {
			if !tempItem.expiryAt.IsZero() && time.Now().After(tempItem.expiryAt) {
				// Still expired, safe to remove
				l.cache.Remove(key)
			}
		}
		l.mu.Unlock()
		return nil, false, nil // Treat expired as a cache miss
	}

	// Item exists and is not expired
	valueBytes := make([]byte, len(item.value))
	copy(valueBytes, item.value) // Make a copy to avoid potential concurrent modification
	l.mu.RUnlock()

	return valueBytes, true, nil
}

// Set adds an item to the LRU cache with an optional TTL.
func (l *LRUAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	_ = ctx // Context is not used for in-memory operations

	var expiry time.Time
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}

	item := lruItem{
		value:    value,
		expiryAt: expiry,
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Add(key, item)

	return nil
}

// Delete removes an item from the LRU cache.
func (l *LRUAdapter) Delete(ctx context.Context, key string) (bool, error) {
	_ = ctx // Context is not used
	l.mu.Lock()
	defer l.mu.Unlock()

	// Use Peek and Remove to check existence before removal
	_, exists := l.cache.Peek(key)
	if exists {
		l.cache.Remove(key)
	}
	return exists, nil
}

// Exists checks if a non-expired item exists in the LRU cache.
func (l *LRUAdapter) Exists(ctx context.Context, key string) (bool, error) {
	_ = ctx // Context is not used
	l.mu.RLock()
	defer l.mu.RUnlock()

	item, found := l.cache.Peek(key) // Use Peek to avoid affecting LRU order
	if !found {
		return false, nil
	}

	// Check for expiration without removing
	if !item.expiryAt.IsZero() && time.Now().After(item.expiryAt) {
		// Consider expired items as non-existent for this check
		return false, nil
	}

	return true, nil
}

// Flush removes all items from the LRU cache.
func (l *LRUAdapter) Flush(ctx context.Context) error {
	_ = ctx // Context is not used
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache.Purge() // Purge removes all entries
	return nil
}

// Close is a no-op for the in-memory LRU cache adapter.
func (l *LRUAdapter) Close() error {
	// No resources to release for the basic LRU cache
	return nil
}
