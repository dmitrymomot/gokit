package webhook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/webhook"
	"github.com/stretchr/testify/assert"
)

// setupTestServer creates a test HTTP server for benchmark testing
func setupTestServer(b *testing.B, responseDelay time.Duration, responseStatus int) *httptest.Server {
	b.Helper()

	var mu sync.Mutex
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track request count in a thread-safe way
		mu.Lock()
		requestCount++
		count := requestCount
		mu.Unlock()

		// Simulate processing time
		if responseDelay > 0 {
			time.Sleep(responseDelay)
		}

		// Set content type
		w.Header().Set("Content-Type", "application/json")

		// Set status code
		w.WriteHeader(responseStatus)

		// Return response with request count to verify all requests are processed
		if err := json.NewEncoder(w).Encode(map[string]any{
			"success": responseStatus >= 200 && responseStatus < 300,
			"count":   count,
		}); err != nil {
			b.Error(err)
		}
	}))

	// Add cleanup to close the server when test completes
	b.Cleanup(func() {
		server.Close()
	})

	return server
}

// BenchmarkSequentialWebhookSending benchmarks sequential webhook sending performance
func BenchmarkSequentialWebhookSending(b *testing.B) {
	server := setupTestServer(b, 5*time.Millisecond, http.StatusOK)

	// Create webhook sender
	sender := webhook.NewWebhookSender()

	// Parameters to send
	params := map[string]string{
		"event": "test_event",
		"data":  "benchmark_payload",
	}

	// Reset timer before benchmark loop
	b.ResetTimer()

	for b.Loop() {
		ctx := context.Background()
		resp, err := sender.Send(ctx, server.URL, params)

		// Verify successful response but don't fail the benchmark
		if err != nil || resp == nil || !resp.IsSuccessful() {
			b.Logf("Webhook send failed: %v", err)
		}
	}
}

// BenchmarkParallelWebhookSending benchmarks parallel webhook sending performance
func BenchmarkParallelWebhookSending(b *testing.B) {
	server := setupTestServer(b, 5*time.Millisecond, http.StatusOK)

	// Create webhook sender
	sender := webhook.NewWebhookSender()

	// Parameters to send
	params := map[string]string{
		"event": "test_event",
		"data":  "benchmark_payload",
	}

	// Reset timer before benchmark loop
	b.ResetTimer()

	// Run b.N iterations in parallel
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			resp, err := sender.Send(ctx, server.URL, params)

			// Verify successful response but don't fail the benchmark
			if err != nil || resp == nil || !resp.IsSuccessful() {
				b.Logf("Webhook send failed: %v", err)
			}
		}
	})
}

// BenchmarkParallelWebhookSendingWithRetry benchmarks parallel webhook sending with retry
func BenchmarkParallelWebhookSendingWithRetry(b *testing.B) {
	// Create a server that will sometimes fail (50% of the time)
	alternatingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a deterministic status code based on request count
		requestID := time.Now().UnixNano() % 2
		statusCode := http.StatusOK

		if requestID == 0 {
			statusCode = http.StatusInternalServerError
			time.Sleep(10 * time.Millisecond) // Delay errors a bit more
		} else {
			time.Sleep(5 * time.Millisecond)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		json.NewEncoder(w).Encode(map[string]any{
			"success": statusCode == http.StatusOK,
			"id":      requestID,
		})
	}))

	// Add cleanup to close the server when test completes
	b.Cleanup(func() {
		alternatingServer.Close()
	})

	// Create base webhook sender
	baseSender := webhook.NewWebhookSender()

	// Add retry decorator with server error retry
	sender := webhook.NewRetryDecorator(
		baseSender,
		webhook.WithRetryCount(2),
		webhook.WithRetryDelay(20*time.Millisecond),
		webhook.WithRetryOnServerErrors(),
	)

	// Parameters to send
	params := map[string]string{
		"event": "test_event",
		"data":  "benchmark_payload",
	}

	// Reset timer before benchmark loop
	b.ResetTimer()

	// Run b.N iterations in parallel
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			resp, err := sender.Send(ctx, alternatingServer.URL, params)

			// Verify successful response but don't fail the benchmark
			if err != nil || resp == nil || !resp.IsSuccessful() {
				b.Logf("Webhook send with retry failed: %v", err)
			}
		}
	})
}

// BenchmarkParallelWebhookSendingWithLogger benchmarks parallel webhook sending with logger
func BenchmarkParallelWebhookSendingWithLogger(b *testing.B) {
	server := setupTestServer(b, 5*time.Millisecond, http.StatusOK)

	// Create base webhook sender
	baseSender := webhook.NewWebhookSender()

	// Add logger decorator (which should handle concurrent logging properly)
	sender := webhook.NewLoggerDecorator(baseSender, nil)

	// Parameters to send
	params := map[string]string{
		"event": "test_event",
		"data":  "benchmark_payload",
	}

	// Reset timer before benchmark loop
	b.ResetTimer()

	// Run b.N iterations in parallel
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			resp, err := sender.Send(ctx, server.URL, params)

			// Verify successful response but don't fail the benchmark
			if err != nil || resp == nil || !resp.IsSuccessful() {
				b.Logf("Webhook send with logger failed: %v", err)
			}
		}
	})
}

// TestConcurrentSafety tests the thread safety of webhook sending
func TestConcurrentSafety(t *testing.T) {
	// Create a test server that tracks requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}))
	defer server.Close()

	// Create webhook sender
	sender := webhook.NewWebhookSender()

	// Create retry decorator to test its thread safety
	retrySender := webhook.NewRetryDecorator(
		sender,
		webhook.WithRetryCount(1),
		webhook.WithRetryDelay(10*time.Millisecond),
	)

	// Number of concurrent goroutines
	concurrency := 50

	// Create wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(concurrency)

	// Track errors that occur during concurrent sending
	var errMu sync.Mutex
	var errors []error

	// Track success count
	var successCount int32
	var successMu sync.Mutex

	// Launch concurrent goroutines
	for i := range concurrency {
		go func(id int) {
			defer wg.Done()

			params := map[string]any{
				"event":        "test_event",
				"data":         "concurrent_payload",
				"goroutine_id": id,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			// Use the retry sender to test its thread safety
			resp, err := retrySender.Send(ctx, server.URL, params)

			if err != nil {
				errMu.Lock()
				errors = append(errors, err)
				errMu.Unlock()
				return
			}

			if resp.IsSuccessful() {
				successMu.Lock()
				successCount++
				successMu.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify no errors occurred
	assert.Empty(t, errors, "Expected no errors during concurrent webhook sending")

	// Verify all requests were successful
	assert.Equal(t, int32(concurrency), successCount, "Expected all concurrent requests to succeed")
}
