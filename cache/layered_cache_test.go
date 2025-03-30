package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/cache"
)

// setupTestCaches creates a test environment with both LRU and Redis caches.
func setupTestCaches(t *testing.T) (*cache.LRUAdapter, *cache.RedisAdapter, *miniredis.Miniredis) {
	// Set up LRU cache
	lruCache, err := cache.NewLRUAdapter(100)
	require.NoError(t, err)

	// Set up Redis cache with miniredis
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	redisCache, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)

	return lruCache, redisCache, s
}

func TestNewLayeredCache(t *testing.T) {
	lruCache, redisCache, _ := setupTestCaches(t)

	// Test successful creation
	lc, err := cache.NewLayeredCache(lruCache, redisCache)
	require.NoError(t, err)
	require.NotNil(t, lc)

	// Test with nil L1
	lc, err = cache.NewLayeredCache(nil, redisCache)
	require.Error(t, err)
	require.Nil(t, lc)

	// Test with nil L2
	lc, err = cache.NewLayeredCache(lruCache, nil)
	require.Error(t, err)
	require.Nil(t, lc)

	// Test with both nil
	lc, err = cache.NewLayeredCache(nil, nil)
	require.Error(t, err)
	require.Nil(t, lc)
}

func TestLayeredCache_Get(t *testing.T) {
	l1, l2, s := setupTestCaches(t)
	layeredCache, err := cache.NewLayeredCache(l1, l2)
	require.NoError(t, err)
	ctx := context.Background()

	// 1. Test Get on an empty cache (should be cache miss)
	value, found, err := layeredCache.Get(ctx, "nonexistent_key")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, value)

	// 2. Set in L2 directly (simulating another service writing to Redis)
	testKey := "test_l2_key"
	testValue := []byte("test_value")
	s.Set(testKey, string(testValue))

	// 3. Get should retrieve from L2 and populate L1
	value, found, err = layeredCache.Get(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testValue, value)

	// 4. Verify the value is now in L1 by disconnecting Redis
	s.Close()
	
	// Value should still be retrievable from L1
	value, found, err = layeredCache.Get(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testValue, value)
}

func TestLayeredCache_Set(t *testing.T) {
	l1, l2, s := setupTestCaches(t)
	layeredCache, err := cache.NewLayeredCache(l1, l2)
	require.NoError(t, err)
	ctx := context.Background()

	testKey := "test_set_key"
	testValue := []byte("test_set_value")
	testTTL := 5 * time.Minute

	// 1. Set through layered cache
	err = layeredCache.Set(ctx, testKey, testValue, testTTL)
	require.NoError(t, err)

	// 2. Verify in L1
	value, found, err := l1.Get(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testValue, value)

	// 3. Verify in L2
	storedValue, err := s.Get(testKey)
	require.NoError(t, err)
	assert.Equal(t, string(testValue), storedValue)
	assert.Equal(t, testTTL, s.TTL(testKey))

	// 4. Test Set with L2 error
	s.Close()
	err = layeredCache.Set(ctx, "another_key", testValue, testTTL)
	require.Error(t, err)
	// The L1 cache should still have been updated despite L2 error
	value, found, err = l1.Get(ctx, "another_key")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testValue, value)
}

func TestLayeredCache_Delete(t *testing.T) {
	l1, l2, s := setupTestCaches(t)
	layeredCache, err := cache.NewLayeredCache(l1, l2)
	require.NoError(t, err)
	ctx := context.Background()

	testKey := "test_delete_key"
	testValue := []byte("test_delete_value")

	// 1. Set a value in both caches
	err = layeredCache.Set(ctx, testKey, testValue, time.Minute)
	require.NoError(t, err)

	// 2. Verify it's in both caches
	value, found, err := l1.Get(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testValue, value)
	
	storedValue, err := s.Get(testKey)
	require.NoError(t, err)
	assert.Equal(t, string(testValue), storedValue)

	// 3. Delete the key
	deleted, err := layeredCache.Delete(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, deleted)

	// 4. Verify it's deleted from both caches
	_, found, err = l1.Get(ctx, testKey)
	require.NoError(t, err)
	assert.False(t, found)
	
	assert.False(t, s.Exists(testKey))

	// 5. Test Delete with non-existent key
	deleted, err = layeredCache.Delete(ctx, "nonexistent_key")
	require.NoError(t, err)
	assert.False(t, deleted)

	// 6. Test Delete with L2 error but L1 success
	testKey = "l2_error_key"
	err = l1.Set(ctx, testKey, testValue, time.Minute)
	require.NoError(t, err)
	s.Set(testKey, string(testValue))
	
	s.Close() // Close Redis connection
	
	// Delete should fail because L2 is down
	_, err = layeredCache.Delete(ctx, testKey)
	require.Error(t, err)
}

func TestLayeredCache_Exists(t *testing.T) {
	l1, l2, s := setupTestCaches(t)
	layeredCache, err := cache.NewLayeredCache(l1, l2)
	require.NoError(t, err)
	ctx := context.Background()

	testKey := "test_exists_key"
	testValue := []byte("test_exists_value")

	// 1. Test Exists on an empty cache
	exists, err := layeredCache.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.False(t, exists)

	// 2. Set a value in L1 only
	err = l1.Set(ctx, testKey, testValue, time.Minute)
	require.NoError(t, err)

	// 3. Check Exists (should be true)
	exists, err = layeredCache.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, exists)

	// 4. Delete from L1 but set in L2
	_, err = l1.Delete(ctx, testKey)
	require.NoError(t, err)
	s.Set(testKey, string(testValue))

	// 5. Check Exists again (should be true from L2)
	exists, err = layeredCache.Exists(ctx, testKey)
	require.NoError(t, err)
	assert.True(t, exists)

	// 6. Test with L2 error
	s.Close()
	_, err = layeredCache.Exists(ctx, testKey)
	require.Error(t, err)
}

func TestLayeredCache_Flush(t *testing.T) {
	l1, l2, s := setupTestCaches(t)
	layeredCache, err := cache.NewLayeredCache(l1, l2)
	require.NoError(t, err)
	ctx := context.Background()

	// 1. Set values in both caches
	err = layeredCache.Set(ctx, "key1", []byte("val1"), time.Minute)
	require.NoError(t, err)
	err = layeredCache.Set(ctx, "key2", []byte("val2"), time.Minute)
	require.NoError(t, err)

	// 2. Verify keys exist
	exists, err := layeredCache.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.True(t, exists)
	exists, err = layeredCache.Exists(ctx, "key2")
	require.NoError(t, err)
	assert.True(t, exists)

	// 3. Flush the cache
	err = layeredCache.Flush(ctx)
	require.NoError(t, err)

	// 4. Verify keys are gone
	exists, err = layeredCache.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = layeredCache.Exists(ctx, "key2")
	require.NoError(t, err)
	assert.False(t, exists)

	// 5. Test Flush with L2 error
	err = layeredCache.Set(ctx, "key3", []byte("val3"), time.Minute)
	require.NoError(t, err)
	
	s.Close() // Close Redis connection
	
	// Flush should fail because L2 is down
	err = layeredCache.Flush(ctx)
	require.Error(t, err)
}

func TestLayeredCache_Close(t *testing.T) {
	l1, l2, _ := setupTestCaches(t)
	layeredCache, err := cache.NewLayeredCache(l1, l2)
	require.NoError(t, err)

	// Close should succeed
	err = layeredCache.Close()
	require.NoError(t, err)
}
