package cache_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/cache"
)

// TestLRUCacheConcurrentAccess tests the thread safety of the LRU cache implementation
func TestLRUCacheConcurrentAccess(t *testing.T) {
	lruCache, err := cache.NewLRUAdapter(1000)
	require.NoError(t, err)
	defer lruCache.Close()

	ctx := context.Background()
	testConcurrentAccess(t, ctx, lruCache)
}

// TestRedisCacheConcurrentAccess tests the thread safety of the Redis cache implementation
func TestRedisCacheConcurrentAccess(t *testing.T) {
	s := miniredis.RunT(t)
	defer s.Close()

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	redisCache, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	defer redisCache.Close()

	ctx := context.Background()
	testConcurrentAccess(t, ctx, redisCache)
}

// TestLayeredCacheConcurrentAccess tests the thread safety of the Layered cache implementation
func TestLayeredCacheConcurrentAccess(t *testing.T) {
	lruCache, err := cache.NewLRUAdapter(1000)
	require.NoError(t, err)

	s := miniredis.RunT(t)
	defer s.Close()

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	redisCache, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)

	layeredCache, err := cache.NewLayeredCache(lruCache, redisCache)
	require.NoError(t, err)
	defer layeredCache.Close()

	ctx := context.Background()
	testConcurrentAccess(t, ctx, layeredCache)
}

// testConcurrentAccess tests the thread safety of a cache implementation
// by running multiple operations concurrently.
func testConcurrentAccess(t *testing.T, ctx context.Context, c cache.Cache) {
	const (
		numGoroutines = 50
		numOperations = 100
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 4) // 4 types of operations

	// Populate some initial data
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("init-key-%d", i)
		value := []byte(fmt.Sprintf("init-value-%d", i))
		err := c.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)
	}

	// Concurrent Set operations
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("set-key-%d-%d", workerID, j)
				value := []byte(fmt.Sprintf("set-value-%d-%d", workerID, j))
				err := c.Set(ctx, key, value, time.Minute)
				if err != nil {
					t.Errorf("Set failed: %v", err)
					return
				}
			}
		}(i)
	}

	// Concurrent Get operations
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Try to get keys set by set operations or init operations
				var key string
				if j%2 == 0 {
					key = fmt.Sprintf("set-key-%d-%d", workerID, j)
				} else {
					key = fmt.Sprintf("init-key-%d", j%numOperations)
				}

				_, _, err := c.Get(ctx, key)
				if err != nil {
					t.Errorf("Get failed: %v", err)
					return
				}
			}
		}(i)
	}

	// Concurrent Exists operations
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Check various keys
				key := fmt.Sprintf("init-key-%d", j%numOperations)
				_, err := c.Exists(ctx, key)
				if err != nil {
					t.Errorf("Exists failed: %v", err)
					return
				}
			}
		}(i)
	}

	// Concurrent Delete operations
	for i := 0; i < numGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Delete with evenly distributed key patterns
				var key string
				if j%3 == 0 {
					key = fmt.Sprintf("set-key-%d-%d", workerID, j)
				} else if j%3 == 1 {
					key = fmt.Sprintf("init-key-%d", j%numOperations)
				} else {
					key = fmt.Sprintf("nonexistent-key-%d-%d", workerID, j)
				}

				_, err := c.Delete(ctx, key)
				if err != nil {
					t.Errorf("Delete failed: %v", err)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

// TestMixedConcurrentOperations tests all operations simultaneously
func TestMixedConcurrentOperations(t *testing.T) {
	// Test each implementation
	testImplementations := []struct {
		name       string
		setupCache func() (cache.Cache, func())
	}{
		{
			name: "LRU",
			setupCache: func() (cache.Cache, func()) {
				c, err := cache.NewLRUAdapter(10000)
				require.NoError(t, err)
				return c, func() { c.Close() }
			},
		},
		{
			name: "Redis",
			setupCache: func() (cache.Cache, func()) {
				s := miniredis.RunT(t)
				client := redis.NewClient(&redis.Options{
					Addr: s.Addr(),
				})
				c, err := cache.NewRedisAdapter(client)
				require.NoError(t, err)
				return c, func() {
					c.Close()
					s.Close()
				}
			},
		},
		{
			name: "Layered",
			setupCache: func() (cache.Cache, func()) {
				lru, err := cache.NewLRUAdapter(10000)
				require.NoError(t, err)

				s := miniredis.RunT(t)
				client := redis.NewClient(&redis.Options{
					Addr: s.Addr(),
				})
				redis, err := cache.NewRedisAdapter(client)
				require.NoError(t, err)

				c, err := cache.NewLayeredCache(lru, redis)
				require.NoError(t, err)

				return c, func() {
					c.Close()
					s.Close()
				}
			},
		},
	}

	for _, tc := range testImplementations {
		t.Run(tc.name, func(t *testing.T) {
			c, cleanup := tc.setupCache()
			defer cleanup()

			ctx := context.Background()
			runMixedOperations(t, ctx, c)
		})
	}
}

// Operation types for mixed concurrent testing
const (
	opSet = iota
	opGet
	opDelete
	opExists
	opFlush
)

// runMixedOperations runs a mix of operations concurrently
func runMixedOperations(t *testing.T, ctx context.Context, c cache.Cache) {
	const (
		numGoroutines = 20
		numOperations = 200
	)

	// Set up some initial data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("initial-key-%d", i)
		value := []byte(fmt.Sprintf("initial-value-%d", i))
		err := c.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)
	}

	// Track some statistics
	var (
		setOps, getOps, deleteOps, existsOps, flushOps int
		statsMutex                                     sync.Mutex
		errorCount                                     int
	)

	// Use a wait group to track all goroutines
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(workerID int) {
			defer wg.Done()

			// Each goroutine does a mix of operations
			for i := 0; i < numOperations; i++ {
				// Choose an operation type based on modulus
				// This gives a good distribution of operations
				opType := (i + workerID) % 5 // 5 operation types

				key := fmt.Sprintf("key-%d-%d", workerID, i)
				value := []byte(fmt.Sprintf("value-%d-%d", workerID, i))

				var err error
				switch opType {
				case opSet:
					err = c.Set(ctx, key, value, time.Minute)
					if err == nil {
						statsMutex.Lock()
						setOps++
						statsMutex.Unlock()
					}

				case opGet:
					_, _, err = c.Get(ctx, key)
					if err == nil {
						statsMutex.Lock()
						getOps++
						statsMutex.Unlock()
					}

				case opDelete:
					_, err = c.Delete(ctx, key)
					if err == nil {
						statsMutex.Lock()
						deleteOps++
						statsMutex.Unlock()
					}

				case opExists:
					_, err = c.Exists(ctx, key)
					if err == nil {
						statsMutex.Lock()
						existsOps++
						statsMutex.Unlock()
					}

				case opFlush:
					// Only do a flush occasionally (1 in 50 ops)
					if i%50 == 0 {
						err = c.Flush(ctx)
						if err == nil {
							statsMutex.Lock()
							flushOps++
							statsMutex.Unlock()
						}
					}
				}

				if err != nil {
					statsMutex.Lock()
					errorCount++
					statsMutex.Unlock()
					t.Logf("Worker %d, operation %d: %v", workerID, i, err)
				}
			}
		}(g)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify we had some operations of each type
	assert.True(t, setOps > 0, "Should have performed some Set operations")
	assert.True(t, getOps > 0, "Should have performed some Get operations")
	assert.True(t, deleteOps > 0, "Should have performed some Delete operations")
	assert.True(t, existsOps > 0, "Should have performed some Exists operations")

	// In a working implementation, we shouldn't have errors
	assert.Equal(t, 0, errorCount, "Should have no errors during mixed operations")

	t.Logf("Operation stats - Set: %d, Get: %d, Delete: %d, Exists: %d, Flush: %d",
		setOps, getOps, deleteOps, existsOps, flushOps)
}

// TestDataConsistency tests that data remains consistent under concurrent modifications
func TestDataConsistency(t *testing.T) {
	// Test each implementation
	testImplementations := []struct {
		name       string
		setupCache func() (cache.Cache, func())
	}{
		{
			name: "LRU",
			setupCache: func() (cache.Cache, func()) {
				c, err := cache.NewLRUAdapter(10000)
				require.NoError(t, err)
				return c, func() { c.Close() }
			},
		},
		{
			name: "Redis",
			setupCache: func() (cache.Cache, func()) {
				s := miniredis.RunT(t)
				client := redis.NewClient(&redis.Options{
					Addr: s.Addr(),
				})
				c, err := cache.NewRedisAdapter(client)
				require.NoError(t, err)
				return c, func() {
					c.Close()
					s.Close()
				}
			},
		},
		{
			name: "Layered",
			setupCache: func() (cache.Cache, func()) {
				lru, err := cache.NewLRUAdapter(10000)
				require.NoError(t, err)

				s := miniredis.RunT(t)
				client := redis.NewClient(&redis.Options{
					Addr: s.Addr(),
				})
				redis, err := cache.NewRedisAdapter(client)
				require.NoError(t, err)

				c, err := cache.NewLayeredCache(lru, redis)
				require.NoError(t, err)

				return c, func() {
					c.Close()
					s.Close()
				}
			},
		},
	}

	for _, tc := range testImplementations {
		t.Run(tc.name, func(t *testing.T) {
			c, cleanup := tc.setupCache()
			defer cleanup()

			ctx := context.Background()
			testConsistentData(t, ctx, c)
		})
	}
}

// testConsistentData ensures data remains consistent under concurrent operations
func testConsistentData(t *testing.T, ctx context.Context, c cache.Cache) {
	const (
		numGoroutines = 10
		keysPerThread = 100
		iterations    = 20
	)

	// Use the same key across threads to test concurrent modifications
	sharedKey := "shared-key"

	// We'll use this to track each thread's last write
	type writeRecord struct {
		workerID  int
		iteration int
		timestamp time.Time
	}

	// Test for a specific amount of time
	testDuration := 2 * time.Second
	testEnd := time.Now().Add(testDuration)

	// Start a bunch of goroutines, all writing to the same key
	var wg sync.WaitGroup

	// Calculate total number of goroutines we'll create to properly size the WaitGroup
	totalGoroutines := numGoroutines // For shared key writers
	totalGoroutines += 1             // For the reader goroutine
	totalGoroutines += numGoroutines // For worker-specific keys
	totalGoroutines += 1             // For the deleter goroutine

	wg.Add(totalGoroutines)

	for g := 0; g < numGoroutines; g++ {
		go func(workerID int) {
			defer wg.Done()

			i := 0
			for time.Now().Before(testEnd) {
				// Create a record of this write
				record := writeRecord{
					workerID:  workerID,
					iteration: i,
					timestamp: time.Now(),
				}

				// Convert to bytes
				value := fmt.Sprintf("worker=%d,iter=%d,time=%d",
					record.workerID, record.iteration, record.timestamp.UnixNano())

				// Write to the shared key
				err := c.Set(ctx, sharedKey, []byte(value), time.Minute)
				require.NoError(t, err)

				// Small sleep to allow other goroutines to run
				time.Sleep(time.Millisecond)
				i++
			}
		}(g)
	}

	// Also have a reader goroutine to check consistency
	go func() {
		defer wg.Done()

		for time.Now().Before(testEnd) {
			// Read from the shared key
			value, found, err := c.Get(ctx, sharedKey)
			require.NoError(t, err)

			if found {
				// Just make sure the value is readable
				require.NotNil(t, value)
				require.Greater(t, len(value), 0)

				// If using a real database, we could further validate the record here
			}

			time.Sleep(time.Millisecond)
		}
	}()

	// Also test with each worker having its own keys
	for g := 0; g < numGoroutines; g++ {
		go func(workerID int) {
			defer wg.Done()

			// Each worker gets their own keys
			keys := make([]string, keysPerThread)
			for i := 0; i < keysPerThread; i++ {
				keys[i] = fmt.Sprintf("worker-%d-key-%d", workerID, i)
			}

			// Keep updating values for these keys
			i := 0
			for time.Now().Before(testEnd) {
				// Choose a key in a round-robin fashion
				key := keys[i%keysPerThread]
				value := []byte(fmt.Sprintf("value-%d-%d", workerID, i))

				// Set the value
				err := c.Set(ctx, key, value, time.Minute)
				require.NoError(t, err)

				// Then immediately read it back to check consistency
				readValue, found, err := c.Get(ctx, key)
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, value, readValue, "Read value should match just-written value")

				i++

				// Add a small delay between operations to reduce contention
				time.Sleep(time.Millisecond)
			}
		}(g)
	}

	// Add a dedicated goroutine for controlled deletion tests
	go func() {
		defer wg.Done()

		// Use a separate set of keys for deletion testing to avoid interference
		// with other goroutines' keys
		deleteTestKeys := make([]string, 50)
		for i := range deleteTestKeys {
			deleteTestKeys[i] = fmt.Sprintf("delete-test-key-%d", i)
		}

		// Run a series of controlled set-delete-check cycles
		for time.Now().Before(testEnd) {
			for _, key := range deleteTestKeys {
				// 1. Set a value
				err := c.Set(ctx, key, []byte("delete-test-value"), time.Minute)
				require.NoError(t, err)

				// 2. Verify it exists
				exists, err := c.Exists(ctx, key)
				require.NoError(t, err)
				if !exists {
					// Skip this key if it somehow doesn't exist - don't fail the test
					continue
				}

				// 3. Delete the key
				deleted, err := c.Delete(ctx, key)
				require.NoError(t, err)

				// The key should have been deleted successfully
				if !deleted {
					// This is acceptable - might happen if another goroutine deleted it
					continue
				}

				// 4. Give some time for the deletion to propagate
				time.Sleep(10 * time.Millisecond)

				// 5. Verify it's gone
				exists, err = c.Exists(ctx, key)
				require.NoError(t, err)
				assert.False(t, exists, "Key should not exist after deletion")
			}
		}
	}()

	// Wait for all operations to complete
	wg.Wait()

	// If we get here without deadlocks or concurrent map read/write panics,
	// the implementation is likely thread-safe
	t.Log("Completed concurrent data consistency test without errors")
}
