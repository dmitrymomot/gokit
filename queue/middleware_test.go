package queue_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChain(t *testing.T) {
	// Create multiple middleware
	var order []string
	
	middleware1 := func(next queue.Handler) queue.Handler {
		return func(ctx context.Context, job *queue.Job) error {
			order = append(order, "middleware1_before")
			err := next(ctx, job)
			order = append(order, "middleware1_after")
			return err
		}
	}
	
	middleware2 := func(next queue.Handler) queue.Handler {
		return func(ctx context.Context, job *queue.Job) error {
			order = append(order, "middleware2_before")
			err := next(ctx, job)
			order = append(order, "middleware2_after")
			return err
		}
	}
	
	// Create a simple handler
	handler := func(ctx context.Context, job *queue.Job) error {
		order = append(order, "handler")
		return nil
	}
	
	// Chain the middleware
	chainedMiddleware := queue.Chain(middleware1, middleware2)
	finalHandler := chainedMiddleware(handler)
	
	// Create a dummy job
	job, err := queue.NewJob("test", struct{}{})
	require.NoError(t, err)
	
	// Execute the handler with the chained middleware
	err = finalHandler(context.Background(), job)
	require.NoError(t, err)
	
	// Check that middleware was executed in the correct order
	expected := []string{
		"middleware1_before",  // Outer middleware runs first
		"middleware2_before",  // Inner middleware runs next
		"handler",             // Handler runs in the middle
		"middleware2_after",   // Inner middleware completes
		"middleware1_after",   // Outer middleware completes
	}
	
	assert.Equal(t, expected, order)
}

func TestWithLogging(t *testing.T) {
	// Create a logger that writes to a buffer
	buffer := &testLogBuffer{}
	logger := slog.New(slog.NewTextHandler(buffer, &slog.HandlerOptions{Level: slog.LevelInfo}))
	
	// Create the logging middleware
	loggingMiddleware := queue.WithLogging(logger)
	
	// Test with a successful handler
	t.Run("Success", func(t *testing.T) {
		buffer.Reset()
		
		handler := func(ctx context.Context, job *queue.Job) error {
			return nil
		}
		
		// Apply the middleware
		wrappedHandler := loggingMiddleware(handler)
		
		// Create a dummy job
		job, err := queue.NewJob("test_success", struct{}{})
		require.NoError(t, err)
		job.ID = "success-job-id"
		
		// Execute the handler
		err = wrappedHandler(context.Background(), job)
		require.NoError(t, err)
		
		// Check that logs were written
		assert.Contains(t, buffer.String(), "Processing job")
		assert.Contains(t, buffer.String(), "success-job-id")
		assert.Contains(t, buffer.String(), "test_success")
		assert.Contains(t, buffer.String(), "Job completed")
		assert.NotContains(t, buffer.String(), "Job failed")
	})
	
	// Test with a failing handler
	t.Run("Failure", func(t *testing.T) {
		buffer.Reset()
		
		expectedErr := errors.New("test error")
		handler := func(ctx context.Context, job *queue.Job) error {
			return expectedErr
		}
		
		// Apply the middleware
		wrappedHandler := loggingMiddleware(handler)
		
		// Create a dummy job
		job, err := queue.NewJob("test_failure", struct{}{})
		require.NoError(t, err)
		job.ID = "failure-job-id"
		
		// Execute the handler
		err = wrappedHandler(context.Background(), job)
		assert.ErrorIs(t, err, expectedErr)
		
		// Check that logs were written
		assert.Contains(t, buffer.String(), "Processing job")
		assert.Contains(t, buffer.String(), "failure-job-id")
		assert.Contains(t, buffer.String(), "test_failure")
		assert.Contains(t, buffer.String(), "Job failed")
		assert.Contains(t, buffer.String(), "test error")
		assert.NotContains(t, buffer.String(), "Job completed")
	})
}

func TestWithRecovery(t *testing.T) {
	recoveryMiddleware := queue.WithRecovery()
	
	// Test with a panicking handler
	handler := func(ctx context.Context, job *queue.Job) error {
		panic("test panic")
	}
	
	// Apply the middleware
	wrappedHandler := recoveryMiddleware(handler)
	
	// Create a dummy job
	job, err := queue.NewJob("test_panic", struct{}{})
	require.NoError(t, err)
	
	// Execute the handler - should not panic
	err = wrappedHandler(context.Background(), job)
	assert.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrJobPanicked)
}

func TestWithTimeout(t *testing.T) {
	timeoutMiddleware := queue.WithTimeout(100 * time.Millisecond)
	
	// Test with a handler that exceeds the timeout
	t.Run("Timeout", func(t *testing.T) {
		handler := func(ctx context.Context, job *queue.Job) error {
			select {
			case <-time.After(300 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		// Apply the middleware
		wrappedHandler := timeoutMiddleware(handler)
		
		// Create a dummy job
		job, err := queue.NewJob("test_timeout", struct{}{})
		require.NoError(t, err)
		
		// Execute the handler - should return context deadline exceeded
		start := time.Now()
		err = wrappedHandler(context.Background(), job)
		duration := time.Since(start)
		
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		
		// Should take approximately the timeout duration
		assert.Less(t, duration, 200*time.Millisecond, "Handler should have timed out in ~100ms")
	})
	
	// Test with a handler that completes before the timeout
	t.Run("NoTimeout", func(t *testing.T) {
		handler := func(ctx context.Context, job *queue.Job) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		}
		
		// Apply the middleware
		wrappedHandler := timeoutMiddleware(handler)
		
		// Create a dummy job
		job, err := queue.NewJob("test_no_timeout", struct{}{})
		require.NoError(t, err)
		
		// Execute the handler - should complete normally
		err = wrappedHandler(context.Background(), job)
		assert.NoError(t, err)
	})
}

// testLogBuffer is a simple io.Writer that captures log output
type testLogBuffer struct {
	data []byte
}

func (b *testLogBuffer) Write(p []byte) (n int, err error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *testLogBuffer) String() string {
	return string(b.data)
}

func (b *testLogBuffer) Reset() {
	b.data = nil
}

// Create a mock MetricsRecorder for testing the metrics middleware
type mockMetricsRecorder struct {
	startCount      int
	completionCount int
	failureCount    int
	lastTaskName    string
	lastError       error
}

func (m *mockMetricsRecorder) RecordJobStart(ctx context.Context, taskName string) {
	m.startCount++
	m.lastTaskName = taskName
}

func (m *mockMetricsRecorder) RecordJobCompletion(ctx context.Context, taskName string, duration time.Duration) {
	m.completionCount++
	m.lastTaskName = taskName
}

func (m *mockMetricsRecorder) RecordJobFailure(ctx context.Context, taskName string, err error, duration time.Duration) {
	m.failureCount++
	m.lastTaskName = taskName
	m.lastError = err
}

func TestWithMetrics(t *testing.T) {
	recorder := &mockMetricsRecorder{}
	metricsMiddleware := queue.WithMetrics(recorder)
	
	// Test with a successful handler
	t.Run("Success", func(t *testing.T) {
		handler := func(ctx context.Context, job *queue.Job) error {
			return nil
		}
		
		// Apply the middleware
		wrappedHandler := metricsMiddleware(handler)
		
		// Create a dummy job
		job, err := queue.NewJob("test_metrics_success", struct{}{})
		require.NoError(t, err)
		
		// Execute the handler
		err = wrappedHandler(context.Background(), job)
		require.NoError(t, err)
		
		// Check metrics were recorded
		assert.Equal(t, 1, recorder.startCount)
		assert.Equal(t, 1, recorder.completionCount)
		assert.Equal(t, 0, recorder.failureCount)
		assert.Equal(t, "test_metrics_success", recorder.lastTaskName)
	})
	
	// Test with a failing handler
	t.Run("Failure", func(t *testing.T) {
		// Reset the recorder
		recorder = &mockMetricsRecorder{}
		metricsMiddleware = queue.WithMetrics(recorder)
		
		expectedErr := errors.New("metrics test error")
		handler := func(ctx context.Context, job *queue.Job) error {
			return expectedErr
		}
		
		// Apply the middleware
		wrappedHandler := metricsMiddleware(handler)
		
		// Create a dummy job
		job, err := queue.NewJob("test_metrics_failure", struct{}{})
		require.NoError(t, err)
		
		// Execute the handler
		err = wrappedHandler(context.Background(), job)
		assert.ErrorIs(t, err, expectedErr)
		
		// Check metrics were recorded
		assert.Equal(t, 1, recorder.startCount)
		assert.Equal(t, 0, recorder.completionCount)
		assert.Equal(t, 1, recorder.failureCount)
		assert.Equal(t, "test_metrics_failure", recorder.lastTaskName)
		assert.ErrorIs(t, recorder.lastError, expectedErr)
	})
}
