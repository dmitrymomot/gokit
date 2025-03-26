package queue_test

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/dmitrymomot/gokit/queue/memory"
	"github.com/dmitrymomot/gokit/queue/redis"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// BenchPayload is a small payload for benchmarking
type BenchPayload struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// noop handler does nothing for benchmark purposes
func noopHandler(ctx context.Context, payload BenchPayload) error {
	return nil
}

// slowHandler sleeps for a short time to simulate work
func slowHandler(ctx context.Context, payload BenchPayload) error {
	time.Sleep(5 * time.Millisecond)
	return nil
}

// setupMemoryQueue prepares a queue with in-memory storage for benchmarking
func setupMemoryQueue(b *testing.B, concurrency int) (*queue.SimpleQueue, context.CancelFunc) {
	b.Helper()
	storage := memory.New()
	q := queue.New(storage, queue.WithConcurrency(concurrency))

	err := q.AddHandler("noop", noopHandler)
	require.NoError(b, err)

	err = q.AddHandler("slow", slowHandler)
	require.NoError(b, err)

	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := q.Run(ctx); err != nil {
			b.Logf("Queue stopped with error: %v", err)
		}
	}()

	// Wait for queue to start
	time.Sleep(100 * time.Millisecond)

	return q, cancel
}

// trySetupRedisQueue attempts to set up a Redis queue for benchmarking
// This will be skipped if Redis is not available
func trySetupRedisQueue(b *testing.B, concurrency int) (*queue.SimpleQueue, context.CancelFunc, bool) {
	b.Helper()
	
	// Try to connect to Redis, skip if unavailable
	client := redisClient.NewClient(&redisClient.Options{
		Addr: "localhost:6379",
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	
	err := client.Ping(ctx).Err()
	if err != nil {
		return nil, nil, false
	}
	
	// Redis is available, set up the queue
	storage := redis.New(client)
	q := queue.New(storage, queue.WithConcurrency(concurrency))

	err = q.AddHandler("noop", noopHandler)
	require.NoError(b, err)

	err = q.AddHandler("slow", slowHandler)
	require.NoError(b, err)

	// Start the queue
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		if err := q.Run(ctx); err != nil {
			b.Logf("Queue stopped with error: %v", err)
		}
	}()

	// Wait for queue to start
	time.Sleep(100 * time.Millisecond)

	return q, cancel, true
}

// BenchmarkEnqueue measures the performance of enqueueing jobs
func BenchmarkEnqueue(b *testing.B) {
	benchCases := []struct {
		name        string
		concurrency int
		payloads    int
	}{
		{"SingleWorker", 1, 1},
		{"10Workers", 10, 1},
		{"50Workers", 50, 1},
		{"Batch100", 10, 100},
	}

	for _, bc := range benchCases {
		// Run with in-memory storage
		b.Run(fmt.Sprintf("Memory_%s", bc.name), func(b *testing.B) {
			q, cancel := setupMemoryQueue(b, bc.concurrency)
			defer cancel()

			ctx := context.Background()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for p := 0; p < bc.payloads; p++ {
					payload := BenchPayload{
						ID:      i*bc.payloads + p,
						Message: fmt.Sprintf("benchmark message %d", i*bc.payloads+p),
					}
					_, err := q.Enqueue(ctx, "noop", payload)
					if err != nil {
						b.Fatalf("Failed to enqueue job: %v", err)
					}
				}
			}
		})

		// Run with Redis storage if available
		b.Run(fmt.Sprintf("Redis_%s", bc.name), func(b *testing.B) {
			q, cancel, ok := trySetupRedisQueue(b, bc.concurrency)
			if !ok {
				b.Skip("Redis not available")
			}
			defer cancel()

			ctx := context.Background()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for p := 0; p < bc.payloads; p++ {
					payload := BenchPayload{
						ID:      i*bc.payloads + p,
						Message: fmt.Sprintf("benchmark message %d", i*bc.payloads+p),
					}
					_, err := q.Enqueue(ctx, "noop", payload)
					if err != nil {
						b.Fatalf("Failed to enqueue job: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkProcessing measures the performance of job processing
func BenchmarkProcessing(b *testing.B) {
	benchCases := []struct {
		name        string
		concurrency int
		handler     string
	}{
		{"NoopSingleWorker", 1, "noop"},
		{"Noop10Workers", 10, "noop"},
		{"Noop50Workers", 50, "noop"},
		{"SlowSingleWorker", 1, "slow"},
		{"Slow10Workers", 10, "slow"},
		{"Slow50Workers", 50, "slow"},
	}

	for _, bc := range benchCases {
		b.Run(fmt.Sprintf("Memory_%s", bc.name), func(b *testing.B) {
			q, cancel := setupMemoryQueue(b, bc.concurrency)
			defer cancel()

			ctx := context.Background()
			
			// Pre-create payloads to eliminate that overhead from the benchmark
			payloads := make([]BenchPayload, b.N)
			for i := 0; i < b.N; i++ {
				payloads[i] = BenchPayload{
					ID:      i,
					Message: fmt.Sprintf("benchmark message %d", i),
				}
			}
			
			b.ResetTimer()
			
			// Enqueue all jobs
			for i := 0; i < b.N; i++ {
				_, err := q.Enqueue(ctx, bc.handler, payloads[i])
				if err != nil {
					b.Fatalf("Failed to enqueue job: %v", err)
				}
			}
			
			// Wait for all jobs to be processed
			// This is a simple approach - in real benchmarks you might want 
			// to track when all jobs are completed
			time.Sleep(time.Duration(b.N*10) * time.Millisecond)
		})
	}
}

// BenchmarkConcurrentClients measures performance with multiple concurrent clients
func BenchmarkConcurrentClients(b *testing.B) {
	benchCases := []struct {
		name        string
		concurrency int
		clients     int
	}{
		{"1Worker_2Clients", 1, 2},
		{"1Worker_10Clients", 1, 10},
		{"10Workers_10Clients", 10, 10},
		{"50Workers_50Clients", 50, 50},
	}

	for _, bc := range benchCases {
		b.Run(fmt.Sprintf("Memory_%s", bc.name), func(b *testing.B) {
			q, cancel := setupMemoryQueue(b, bc.concurrency)
			defer cancel()

			ctx := context.Background()
			b.ResetTimer()
			
			// Use WaitGroup to synchronize clients
			var wg sync.WaitGroup
			jobsPerClient := b.N / bc.clients
			if jobsPerClient < 1 {
				jobsPerClient = 1
			}
			
			// Start simulated clients
			for c := 0; c < bc.clients; c++ {
				wg.Add(1)
				go func(clientID int) {
					defer wg.Done()
					for i := 0; i < jobsPerClient; i++ {
						payload := BenchPayload{
							ID:      clientID*jobsPerClient + i,
							Message: fmt.Sprintf("client %d message %d", clientID, i),
						}
						_, err := q.Enqueue(ctx, "noop", payload)
						if err != nil {
							b.Logf("Client %d failed to enqueue job: %v", clientID, err)
						}
					}
				}(c)
			}
			
			wg.Wait()
		})
	}
}

// BenchmarkWithMiddleware measures the overhead of middleware
func BenchmarkWithMiddleware(b *testing.B) {
	benchCases := []struct {
		name           string
		withMiddleware bool
	}{
		{"NoMiddleware", false},
		{"WithLogging", true},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			storage := memory.New()
			
			var q *queue.SimpleQueue
			if bc.withMiddleware {
				// Create a null logger that doesn't actually write anywhere
				nullLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
				q = queue.New(storage, 
					queue.WithConcurrency(10),
					queue.WithMiddleware(queue.WithLogging(nullLogger)))
			} else {
				q = queue.New(storage, queue.WithConcurrency(10))
			}

			err := q.AddHandler("noop", noopHandler)
			require.NoError(b, err)

			// Start the queue
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			
			go func() {
				if err := q.Run(ctx); err != nil {
					b.Logf("Queue stopped with error: %v", err)
				}
			}()

			// Wait for queue to start
			time.Sleep(100 * time.Millisecond)
			
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				payload := BenchPayload{
					ID:      i,
					Message: fmt.Sprintf("benchmark message %d", i),
				}
				_, err := q.Enqueue(ctx, "noop", payload)
				if err != nil {
					b.Fatalf("Failed to enqueue job: %v", err)
				}
			}
			
			// Wait for all jobs to be processed
			time.Sleep(time.Duration(b.N*5) * time.Millisecond)
		})
	}
}
