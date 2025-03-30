package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/cache"
)

// setupMiniredis starts a miniredis instance and returns a redis client connected to it.
func setupMiniredis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	return s, client
}

func TestNewRedisAdapter(t *testing.T) {
	_, client := setupMiniredis(t)
	defer client.Close()

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Test connection failure
	fakeClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Non-existent server
	})
	defer fakeClient.Close()
	_, err = cache.NewRedisAdapter(fakeClient)
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrConnectionFailed)
}

func TestRedisAdapter_Get(t *testing.T) {
	s, client := setupMiniredis(t)
	defer client.Close()

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	ctx := context.Background()

	key := "test_get_key"
	value := []byte("test_value")

	// Test Get non-existent key
	val, found, err := adapter.Get(ctx, key)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Nil(t, val)

	// Set a key and then Get it
	require.NoError(t, s.Set(key, string(value)))

	val, found, err = adapter.Get(ctx, key)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, val)

	// Test Get with error (simulate connection drop)
	s.Close() // Close the miniredis server
	_, _, err = adapter.Get(ctx, "another_key")
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrOperationFailed)
}

func TestRedisAdapter_Set(t *testing.T) {
	s, client := setupMiniredis(t)
	defer client.Close()

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	ctx := context.Background()

	key := "test_set_key"
	value := []byte("set_value")
	ttl := 5 * time.Minute

	// Set a key with TTL
	err = adapter.Set(ctx, key, value, ttl)
	require.NoError(t, err)

	// Verify in miniredis
	storedValue, err := s.Get(key)
	require.NoError(t, err)
	assert.Equal(t, string(value), storedValue)
	assert.Equal(t, ttl, s.TTL(key))

	// Set a key without TTL (ttl=0)
	keyNoTTL := "key_no_ttl"
	err = adapter.Set(ctx, keyNoTTL, value, 0)
	require.NoError(t, err)
	storedValue, err = s.Get(keyNoTTL)
	require.NoError(t, err)
	assert.Equal(t, string(value), storedValue)
	assert.Equal(t, time.Duration(0), s.TTL(keyNoTTL)) // Miniredis TTL is 0 if not set

	// Test Set with error (simulate connection drop)
	s.Close()
	err = adapter.Set(ctx, "another_key", value, ttl)
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrOperationFailed)
}

func TestRedisAdapter_Delete(t *testing.T) {
	s, client := setupMiniredis(t)
	defer client.Close()

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	ctx := context.Background()

	key := "test_delete_key"
	value := "delete_value"

	// Test Delete non-existent key
	deleted, err := adapter.Delete(ctx, key)
	require.NoError(t, err)
	assert.False(t, deleted)

	// Set a key and then Delete it
	require.NoError(t, s.Set(key, value))

	deleted, err = adapter.Delete(ctx, key)
	require.NoError(t, err)
	assert.True(t, deleted)

	// Verify it's gone
	_, found, err := adapter.Get(ctx, key)
	require.NoError(t, err)
	assert.False(t, found)

	// Test Delete with error
	s.Close()
	_, err = adapter.Delete(ctx, "another_key")
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrOperationFailed)
}

func TestRedisAdapter_Exists(t *testing.T) {
	s, client := setupMiniredis(t)
	defer client.Close()

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	ctx := context.Background()

	key := "test_exists_key"
	value := "exists_value"

	// Test Exists non-existent key
	exists, err := adapter.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Set a key and then check Exists
	require.NoError(t, s.Set(key, value))

	exists, err = adapter.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test Exists with error
	s.Close()
	_, err = adapter.Exists(ctx, "another_key")
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrOperationFailed)
}

func TestRedisAdapter_Flush(t *testing.T) {
	s, client := setupMiniredis(t)
	defer client.Close()

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)
	ctx := context.Background()

	// Set some keys
	require.NoError(t, s.Set("key1", "val1"))
	require.NoError(t, s.Set("key2", "val2"))

	// Flush the database
	err = adapter.Flush(ctx)
	require.NoError(t, err)

	// Verify keys are gone
	exists, err := adapter.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, exists)
	exists, err = adapter.Exists(ctx, "key2")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test Flush with error
	s.Close()
	err = adapter.Flush(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, cache.ErrOperationFailed)
}

func TestRedisAdapter_Close(t *testing.T) {
	_, client := setupMiniredis(t)

	adapter, err := cache.NewRedisAdapter(client)
	require.NoError(t, err)

	// Close the adapter (which closes the client)
	err = adapter.Close()
	require.NoError(t, err)

	// Try an operation on the closed client (should fail)
	ctx := context.Background()
	_, _, err = adapter.Get(ctx, "some_key")
	require.Error(t, err) // go-redis client returns errors after Close()
	// Note: We check for ErrOperationFailed because the underlying client error
	// might not always map directly to ErrConnectionFailed after a Close()
	assert.True(t, errors.Is(err, cache.ErrOperationFailed) || errors.Is(err, redis.ErrClosed), "Expected ErrOperationFailed or redis.ErrClosed")
}
