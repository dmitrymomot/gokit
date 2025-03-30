package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/cache"
)

func TestNewLRUAdapter(t *testing.T) {
	adapter, err := cache.NewLRUAdapter(10)
	require.NoError(t, err)
	require.NotNil(t, adapter)
}

func TestLRUAdapter_GetSet(t *testing.T) {
	adapter, err := cache.NewLRUAdapter(2) // Small size for testing eviction
	require.NoError(t, err)
	ctx := context.Background()

	key1, val1 := "key1", []byte("value1")
	key2, val2 := "key2", []byte("value2")
	key3, val3 := "key3", []byte("value3")

	// 1. Test Get non-existent
	_, found, err := adapter.Get(ctx, key1)
	require.NoError(t, err)
	assert.False(t, found)

	// 2. Set and Get key1
	err = adapter.Set(ctx, key1, val1, time.Minute)
	require.NoError(t, err)
	retVal, found, err := adapter.Get(ctx, key1)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val1, retVal)

	// 3. Set key2
	err = adapter.Set(ctx, key2, val2, 0) // No expiry
	require.NoError(t, err)
	retVal, found, err = adapter.Get(ctx, key2)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val2, retVal)

	// 4. Set key3 (should evict key1 as it was least recently used)
	err = adapter.Set(ctx, key3, val3, time.Minute)
	require.NoError(t, err)

	// Verify key1 is evicted
	_, found, err = adapter.Get(ctx, key1)
	require.NoError(t, err)
	assert.False(t, found, "key1 should be evicted")

	// Verify key2 is still present
	retVal, found, err = adapter.Get(ctx, key2)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val2, retVal)

	// Verify key3 is present
	retVal, found, err = adapter.Get(ctx, key3)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, val3, retVal)

	// 5. Test Get on an expired key
	keyExp, valExp := "keyExp", []byte("expiredVal")
	err = adapter.Set(ctx, keyExp, valExp, 1*time.Millisecond) // Very short TTL
	require.NoError(t, err)

	time.Sleep(5 * time.Millisecond) // Wait for expiration

	_, found, err = adapter.Get(ctx, keyExp)
	require.NoError(t, err)
	assert.False(t, found, "Expired key should not be found")
}

func TestLRUAdapter_Delete(t *testing.T) {
	adapter, err := cache.NewLRUAdapter(10)
	require.NoError(t, err)
	ctx := context.Background()

	key := "deleteKey"
	val := []byte("deleteVal")

	// 1. Delete non-existent key
	deleted, err := adapter.Delete(ctx, key)
	require.NoError(t, err)
	assert.False(t, deleted)

	// 2. Set and Delete the key
	err = adapter.Set(ctx, key, val, time.Minute)
	require.NoError(t, err)

	deleted, err = adapter.Delete(ctx, key)
	require.NoError(t, err)
	assert.True(t, deleted)

	// 3. Verify deletion
	_, found, err := adapter.Get(ctx, key)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestLRUAdapter_Exists(t *testing.T) {
	adapter, err := cache.NewLRUAdapter(10)
	require.NoError(t, err)
	ctx := context.Background()

	key := "existsKey"
	val := []byte("existsVal")
	expiredKey := "expiredExistsKey"

	// 1. Exists non-existent key
	exists, err := adapter.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// 2. Set and check Exists
	err = adapter.Set(ctx, key, val, time.Minute)
	require.NoError(t, err)
	exists, err = adapter.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// 3. Set an expiring key
	err = adapter.Set(ctx, expiredKey, val, 1*time.Millisecond)
	require.NoError(t, err)

	// Check exists before expiry
	exists, err = adapter.Exists(ctx, expiredKey)
	require.NoError(t, err)
	assert.True(t, exists)

	time.Sleep(5 * time.Millisecond) // Wait for expiration

	// Check exists after expiry
	exists, err = adapter.Exists(ctx, expiredKey)
	require.NoError(t, err)
	assert.False(t, exists, "Expired key should not exist")
}

func TestLRUAdapter_Flush(t *testing.T) {
	adapter, err := cache.NewLRUAdapter(10)
	require.NoError(t, err)
	ctx := context.Background()

	// Set some keys
	require.NoError(t, adapter.Set(ctx, "flush1", []byte("v1"), time.Minute))
	require.NoError(t, adapter.Set(ctx, "flush2", []byte("v2"), 0))

	// Check they exist
	exists, err := adapter.Exists(ctx, "flush1")
	require.NoError(t, err)
	assert.True(t, exists)
	exists, err = adapter.Exists(ctx, "flush2")
	require.NoError(t, err)
	assert.True(t, exists)

	// Flush the cache
	err = adapter.Flush(ctx)
	require.NoError(t, err)

	// Verify keys are gone
	exists, err = adapter.Exists(ctx, "flush1")
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = adapter.Exists(ctx, "flush2")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestLRUAdapter_Close(t *testing.T) {
	adapter, err := cache.NewLRUAdapter(10)
	require.NoError(t, err)

	// Close should be a no-op and return nil
	err = adapter.Close()
	assert.NoError(t, err)
}
