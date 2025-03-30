package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/dmitrymomot/gokit/cache"
)

// setupBenchCache creates a cache instance for benchmarking.
func setupBenchCache(b *testing.B, cacheType string) (cache.Cache, func()) {
	switch cacheType {
	case "lru":
		lru, err := cache.NewLRUAdapter(100000) // Large size for benchmarking
		if err != nil {
			b.Fatalf("Failed to create LRU cache: %v", err)
		}
		return lru, func() { _ = lru.Close() }

	case "redis":
		s := miniredis.RunT(b)
		client := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		redisCache, err := cache.NewRedisAdapter(client)
		if err != nil {
			b.Fatalf("Failed to create Redis cache: %v", err)
		}
		return redisCache, func() {
			_ = redisCache.Close()
			s.Close()
		}

	case "layered":
		lru, err := cache.NewLRUAdapter(100000)
		if err != nil {
			b.Fatalf("Failed to create LRU cache: %v", err)
		}

		s := miniredis.RunT(b)
		client := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		redisCache, err := cache.NewRedisAdapter(client)
		if err != nil {
			b.Fatalf("Failed to create Redis cache: %v", err)
		}

		layeredCache, err := cache.NewLayeredCache(lru, redisCache)
		if err != nil {
			b.Fatalf("Failed to create Layered cache: %v", err)
		}

		return layeredCache, func() {
			_ = layeredCache.Close()
			s.Close()
		}

	default:
		b.Fatalf("Unknown cache type: %s", cacheType)
		return nil, nil
	}
}

// BenchmarkLRUCache_Set benchmarks Set operations on the LRU cache.
func BenchmarkLRUCache_Set(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "lru")
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, key, value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}
}

// BenchmarkRedisCache_Set benchmarks Set operations on the Redis cache.
func BenchmarkRedisCache_Set(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "redis")
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, key, value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}
}

// BenchmarkLayeredCache_Set benchmarks Set operations on the Layered cache.
func BenchmarkLayeredCache_Set(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "layered")
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, key, value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}
}

// BenchmarkLRUCache_Get benchmarks Get operations on the LRU cache.
func BenchmarkLRUCache_Get(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "lru")
	defer cleanup()

	ctx := context.Background()

	// Prepare data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, key, value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through prepared keys to avoid cache misses
		key := fmt.Sprintf("key-%d", i%1000)
		_, found, err := cache.Get(ctx, key)
		if err != nil {
			b.Fatalf("Failed to get: %v", err)
		}
		if !found {
			b.Fatalf("Key not found: %s", key)
		}
	}
}

// BenchmarkRedisCache_Get benchmarks Get operations on the Redis cache.
func BenchmarkRedisCache_Get(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "redis")
	defer cleanup()

	ctx := context.Background()

	// Prepare data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, key, value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through prepared keys to avoid cache misses
		key := fmt.Sprintf("key-%d", i%1000)
		_, found, err := cache.Get(ctx, key)
		if err != nil {
			b.Fatalf("Failed to get: %v", err)
		}
		if !found {
			b.Fatalf("Key not found: %s", key)
		}
	}
}

// BenchmarkLayeredCache_Get benchmarks Get operations on the Layered cache.
func BenchmarkLayeredCache_Get(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "layered")
	defer cleanup()

	ctx := context.Background()

	// Prepare data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, key, value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through prepared keys to avoid cache misses
		key := fmt.Sprintf("key-%d", i%1000)
		_, found, err := cache.Get(ctx, key)
		if err != nil {
			b.Fatalf("Failed to get: %v", err)
		}
		if !found {
			b.Fatalf("Key not found: %s", key)
		}
	}
}

// BenchmarkLRUCache_Delete benchmarks Delete operations on the LRU cache.
func BenchmarkLRUCache_Delete(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "lru")
	defer cleanup()

	ctx := context.Background()

	// Prepare data
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, keys[i], value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cache.Delete(ctx, keys[i])
		if err != nil {
			b.Fatalf("Failed to delete: %v", err)
		}
	}
}

// BenchmarkRedisCache_Delete benchmarks Delete operations on the Redis cache.
func BenchmarkRedisCache_Delete(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "redis")
	defer cleanup()

	ctx := context.Background()

	// Prepare data
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, keys[i], value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cache.Delete(ctx, keys[i])
		if err != nil {
			b.Fatalf("Failed to delete: %v", err)
		}
	}
}

// BenchmarkLayeredCache_Delete benchmarks Delete operations on the Layered cache.
func BenchmarkLayeredCache_Delete(b *testing.B) {
	cache, cleanup := setupBenchCache(b, "layered")
	defer cleanup()

	ctx := context.Background()

	// Prepare data
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = fmt.Sprintf("key-%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		if err := cache.Set(ctx, keys[i], value, time.Minute); err != nil {
			b.Fatalf("Failed to set: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cache.Delete(ctx, keys[i])
		if err != nil {
			b.Fatalf("Failed to delete: %v", err)
		}
	}
}

// BenchmarkLayeredCache_GetWithL2Fill simulates retrieving data that's only in L2 (Redis),
// requiring population of L1 (LRU) cache.
func BenchmarkLayeredCache_GetWithL2Fill(b *testing.B) {
	// Create individual caches
	lru, err := cache.NewLRUAdapter(100000)
	if err != nil {
		b.Fatalf("Failed to create LRU cache: %v", err)
	}

	s := miniredis.RunT(b)
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	redisCache, err := cache.NewRedisAdapter(client)
	if err != nil {
		b.Fatalf("Failed to create Redis cache: %v", err)
	}

	// Create layered cache
	layeredCache, err := cache.NewLayeredCache(lru, redisCache)
	if err != nil {
		b.Fatalf("Failed to create Layered cache: %v", err)
	}
	
	defer func() {
		_ = layeredCache.Close()
		s.Close()
	}()

	ctx := context.Background()

	// Prepare data only in Redis (L2)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		value := fmt.Sprintf("value-%d", i)
		
		// Set directly in Redis
		err := s.Set(key, value)
		if err != nil {
			b.Fatalf("Failed to set in Redis: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through prepared keys
		key := fmt.Sprintf("key-%d", i%1000)
		_, found, err := layeredCache.Get(ctx, key)
		if err != nil {
			b.Fatalf("Failed to get: %v", err)
		}
		if !found {
			b.Fatalf("Key not found: %s", key)
		}
	}
}
