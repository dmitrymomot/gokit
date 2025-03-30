package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter implements the Cache interface using a Redis client.
type RedisAdapter struct {
	client redis.UniversalClient
}

// NewRedisAdapter creates a new Redis cache adapter.
// It accepts any redis.UniversalClient (e.g., *redis.Client, *redis.ClusterClient).
func NewRedisAdapter(client redis.UniversalClient) (*RedisAdapter, error) {
	// Ping the server to ensure connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Join(ErrConnectionFailed, err)
	}
	return &RedisAdapter{client: client}, nil
}

// Get retrieves an item from the Redis cache.
func (r *RedisAdapter) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil // Use nil error for cache miss, consistent with interface
		}
		return nil, false, errors.Join(ErrOperationFailed, err)
	}
	return val, true, nil
}

// Set adds an item to the Redis cache.
func (r *RedisAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return errors.Join(ErrOperationFailed, err)
	}
	return nil
}

// Delete removes an item from the Redis cache.
func (r *RedisAdapter) Delete(ctx context.Context, key string) (bool, error) {
	deletedCount, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return false, errors.Join(ErrOperationFailed, err)
	}
	return deletedCount > 0, nil
}

// Exists checks if an item exists in the Redis cache.
func (r *RedisAdapter) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, errors.Join(ErrOperationFailed, err)
	}
	return exists > 0, nil
}

// Flush removes all items from the current Redis database.
// Warning: Use with caution, especially in clustered environments or shared databases.
func (r *RedisAdapter) Flush(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return errors.Join(ErrOperationFailed, err)
	}
	return nil
}

// Close closes the Redis client connection.
func (r *RedisAdapter) Close() error {
	if err := r.client.Close(); err != nil {
		return errors.Join(ErrConnectionFailed, err) // Reusing ConnectionFailed for close errors
	}
	return nil
}
