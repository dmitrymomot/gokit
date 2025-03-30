package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// LayeredCache combines two cache layers (e.g., local LRU and distributed Redis).
// Reads first check the primary (L1) cache, then the secondary (L2) cache.
// Writes go to both caches.
type LayeredCache struct {
	l1 Cache // Primary cache (e.g., LRU)
	l2 Cache // Secondary cache (e.g., Redis)
}

// NewLayeredCache creates a new layered cache.
// l1: The primary, faster cache (e.g., LRUAdapter).
// l2: The secondary, persistent cache (e.g., RedisAdapter).
func NewLayeredCache(l1, l2 Cache) (*LayeredCache, error) {
	if l1 == nil || l2 == nil {
		return nil, errors.New("cache: both L1 and L2 caches must be provided")
	}
	return &LayeredCache{l1: l1, l2: l2}, nil
}

// Get retrieves an item, checking L1 then L2.
// If found in L2 but not L1, it's added to L1 before returning.
func (lc *LayeredCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	// 1. Try L1 cache
	val, found, err := lc.l1.Get(ctx, key)
	if err != nil {
		// Log L1 error? For now, treat as miss and try L2.
		fmt.Printf("LayeredCache: Error getting key '%s' from L1: %v\n", key, err)
	}
	if found {
		return val, true, nil // Found in L1
	}

	// 2. Try L2 cache
	val, found, err = lc.l2.Get(ctx, key)
	if err != nil {
		// L2 error is more critical
		return nil, false, fmt.Errorf("LayeredCache: Error getting key '%s' from L2: %w", key, err)
	}
	if !found {
		return nil, false, nil // Not found in L1 or L2
	}

	// 3. Found in L2, not in L1. Add to L1 (best effort).
	// Determine TTL for L1 - should we use a default or try to infer from L2?
	// For simplicity, use a default or maybe a configurable L1 TTL.
	// Let's use a short default TTL for L1 for now.
	l1TTL := 5 * time.Minute // TODO: Make this configurable
	if setErr := lc.l1.Set(ctx, key, val, l1TTL); setErr != nil {
		// Log L1 set error? For now, ignore and return the value from L2.
		fmt.Printf("LayeredCache: Error setting key '%s' in L1 after L2 hit: %v\n", key, setErr)
	}

	return val, true, nil // Return value found in L2
}

// Set adds an item to both L1 and L2 caches concurrently.
func (lc *LayeredCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	eg, childCtx := errgroup.WithContext(ctx)

	// Set in L1
	eg.Go(func() error {
		// Use a potentially shorter TTL for L1 if needed, or the same TTL.
		// Using the same TTL for now.
		err := lc.l1.Set(childCtx, key, value, ttl)
		if err != nil {
			// Log L1 set error? Return it to potentially signal partial failure.
			fmt.Printf("LayeredCache: Error setting key '%s' in L1: %v\n", key, err)
			return fmt.Errorf("L1 set failed: %w", err)
		}
		return nil
	})

	// Set in L2
	eg.Go(func() error {
		err := lc.l2.Set(childCtx, key, value, ttl)
		if err != nil {
			fmt.Printf("LayeredCache: Error setting key '%s' in L2: %v\n", key, err)
			return fmt.Errorf("L2 set failed: %w", err)
		}
		return nil
	})

	// Wait for both operations
	if err := eg.Wait(); err != nil {
		// Return the combined/first error
		return errors.Join(ErrOperationFailed, err)
	}

	return nil
}

// Delete removes an item from both L1 and L2 caches concurrently.
// Returns true if the item was deleted from L2 (considered the source of truth).
func (lc *LayeredCache) Delete(ctx context.Context, key string) (bool, error) {
	eg, childCtx := errgroup.WithContext(ctx)
	var deletedL2 bool
	var deleteL2Err error

	// Delete from L1 (best effort)
	eg.Go(func() error {
		_, err := lc.l1.Delete(childCtx, key)
		if err != nil {
			fmt.Printf("LayeredCache: Error deleting key '%s' from L1: %v\n", key, err)
			// Don't return error here, L2 is the source of truth for deletion status
		}
		return nil // L1 deletion is best effort
	})

	// Delete from L2
	eg.Go(func() error {
		var err error
		deletedL2, err = lc.l2.Delete(childCtx, key)
		if err != nil {
			fmt.Printf("LayeredCache: Error deleting key '%s' from L2: %v\n", key, err)
			deleteL2Err = fmt.Errorf("L2 delete failed: %w", err)
			return deleteL2Err // Propagate L2 error
		}
		return nil
	})

	// Wait for both operations
	waitErr := eg.Wait() // This will capture deleteL2Err if it occurred

	if waitErr != nil {
		return false, errors.Join(ErrOperationFailed, waitErr)
	}

	return deletedL2, nil // Return L2 deletion status
}

// Exists checks if an item exists, checking L1 then L2.
func (lc *LayeredCache) Exists(ctx context.Context, key string) (bool, error) {
	// 1. Check L1
	exists, err := lc.l1.Exists(ctx, key)
	if err != nil {
		fmt.Printf("LayeredCache: Error checking existence for key '%s' in L1: %v\n", key, err)
		// Proceed to check L2 even if L1 fails
	}
	if exists {
		return true, nil
	}

	// 2. Check L2
	exists, err = lc.l2.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("LayeredCache: Error checking existence for key '%s' in L2: %w", key, err)
	}

	// If it exists in L2 but not L1, we don't backfill on Exists call.
	return exists, nil
}

// Flush removes all items from both L1 and L2 caches concurrently.
func (lc *LayeredCache) Flush(ctx context.Context) error {
	eg, childCtx := errgroup.WithContext(ctx)

	// Flush L1
	eg.Go(func() error {
		err := lc.l1.Flush(childCtx)
		if err != nil {
			fmt.Printf("LayeredCache: Error flushing L1: %v\n", err)
			return fmt.Errorf("L1 flush failed: %w", err)
		}
		return nil
	})

	// Flush L2
	eg.Go(func() error {
		err := lc.l2.Flush(childCtx)
		if err != nil {
			fmt.Printf("LayeredCache: Error flushing L2: %v\n", err)
			return fmt.Errorf("L2 flush failed: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return errors.Join(ErrOperationFailed, err)
	}

	return nil
}

// Close closes both L1 and L2 cache connections concurrently.
func (lc *LayeredCache) Close() error {
	// Use a simple WaitGroup as context isn't directly applicable here
	var wg sync.WaitGroup
	var closeErrs []error
	var mu sync.Mutex

	wg.Add(2)

	// Close L1
	go func() {
		defer wg.Done()
		if err := lc.l1.Close(); err != nil {
			fmt.Printf("LayeredCache: Error closing L1: %v\n", err)
			mu.Lock()
			closeErrs = append(closeErrs, fmt.Errorf("L1 close failed: %w", err))
			mu.Unlock()
		}
	}()

	// Close L2
	go func() {
		defer wg.Done()
		if err := lc.l2.Close(); err != nil {
			fmt.Printf("LayeredCache: Error closing L2: %v\n", err)
			mu.Lock()
			closeErrs = append(closeErrs, fmt.Errorf("L2 close failed: %w", err))
			mu.Unlock()
		}
	}()

	wg.Wait()

	if len(closeErrs) > 0 {
		return errors.Join(closeErrs...)
	}

	return nil
}
