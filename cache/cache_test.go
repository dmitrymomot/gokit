package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/cache"
)

// TestCacheInterface verifies that all implementations properly satisfy the Cache interface
func TestCacheInterface(t *testing.T) {
	// Test each implementation separately
	testImplementations := []struct {
		name        string
		setupCache  func() (cache.Cache, func())
	}{
		{
			name: "LRU",
			setupCache: func() (cache.Cache, func()) {
				adapter, err := cache.NewLRUAdapter(100)
				require.NoError(t, err)
				return adapter, func() { _ = adapter.Close() }
			},
		},
		{
			name: "Redis",
			setupCache: func() (cache.Cache, func()) {
				s := miniredis.RunT(t)
				redisClient := redis.NewClient(&redis.Options{
					Addr: s.Addr(),
				})
				adapter, err := cache.NewRedisAdapter(redisClient)
				require.NoError(t, err)
				return adapter, func() { 
					_ = adapter.Close()
					s.Close()
				}
			},
		},
		{
			name: "Layered",
			setupCache: func() (cache.Cache, func()) {
				// Create fresh instances for each test
				lruCache, err := cache.NewLRUAdapter(100)
				require.NoError(t, err)
				
				s := miniredis.RunT(t)
				redisClient := redis.NewClient(&redis.Options{
					Addr: s.Addr(),
				})
				redisCache, err := cache.NewRedisAdapter(redisClient)
				require.NoError(t, err)
				
				layeredCache, err := cache.NewLayeredCache(lruCache, redisCache)
				require.NoError(t, err)
				
				return layeredCache, func() { 
					_ = layeredCache.Close()
					s.Close() 
				}
			},
		},
	}

	for _, tc := range testImplementations {
		t.Run(tc.name, func(t *testing.T) {
			cache, cleanup := tc.setupCache()
			defer cleanup()
			
			testCacheImplementation(t, cache)
		})
	}
}

// testCacheImplementation runs a basic functionality test on any Cache implementation
func testCacheImplementation(t *testing.T, c cache.Cache) {
	ctx := context.Background()
	testKey := "test_key"
	testValue := []byte("test_value")

	// Test Set and Get
	err := c.Set(ctx, testKey, testValue, time.Minute)
	require.NoError(t, err)

	value, found, err := c.Get(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testValue, value)

	// Test Exists
	exists, err := c.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Delete
	deleted, err := c.Delete(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, deleted)

	// Verify deletion
	exists, err = c.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.False(t, exists)

	// We don't test Close() here anymore since it's handled by the cleanup function
}

// TestCommonCacheOperations tests operations common to all cache implementations
// with different data types, key patterns, and edge cases
func TestCommonCacheOperations(t *testing.T) {
	// Create a small LRU cache for quick testing
	c, err := cache.NewLRUAdapter(10)
	require.NoError(t, err)
	
	ctx := context.Background()
	
	t.Run("EmptyValues", func(t *testing.T) {
		err := c.Set(ctx, "empty_key", []byte{}, time.Minute)
		require.NoError(t, err)
		
		val, found, err := c.Get(ctx, "empty_key")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, []byte{}, val)
	})
	
	t.Run("BinaryData", func(t *testing.T) {
		binaryData := []byte{0x00, 0xFF, 0x0F, 0xF0, 0x01, 0x02}
		err := c.Set(ctx, "binary_key", binaryData, time.Minute)
		require.NoError(t, err)
		
		val, found, err := c.Get(ctx, "binary_key")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, binaryData, val)
	})
	
	t.Run("LongKeys", func(t *testing.T) {
		longKey := "very_long_key_" + string(make([]byte, 100)) // 100-byte suffix
		err := c.Set(ctx, longKey, []byte("long_key_value"), time.Minute)
		require.NoError(t, err)
		
		val, found, err := c.Get(ctx, longKey)
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, []byte("long_key_value"), val)
	})
	
	t.Run("ZeroTTL", func(t *testing.T) {
		err := c.Set(ctx, "zero_ttl_key", []byte("zero_ttl_value"), 0)
		require.NoError(t, err)
		
		// Zero TTL should not expire
		time.Sleep(10 * time.Millisecond)
		
		val, found, err := c.Get(ctx, "zero_ttl_key")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, []byte("zero_ttl_value"), val)
	})
	
	t.Run("FlushOperation", func(t *testing.T) {
		// Set multiple keys
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("flush_key_%d", i)
			err := c.Set(ctx, key, []byte("flush_value"), time.Minute)
			require.NoError(t, err)
		}
		
		// Verify at least one exists
		exists, err := c.Exists(ctx, "flush_key_0")
		require.NoError(t, err)
		assert.True(t, exists)
		
		// Flush all
		err = c.Flush(ctx)
		require.NoError(t, err)
		
		// Verify keys are gone
		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("flush_key_%d", i)
			exists, err := c.Exists(ctx, key)
			require.NoError(t, err)
			assert.False(t, exists)
		}
	})
}
