package queue_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/dmitrymomot/gokit/queue/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func TestSimpleQueue_AddHandler(t *testing.T) {
	q := queue.New(memory.New())

	// Valid handler
	err := q.AddHandler("send_email", func(ctx context.Context, payload testEmailPayload) error {
		return nil
	})
	require.NoError(t, err)

	// Invalid handler - not a function
	err = q.AddHandler("invalid", "not a function")
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrInvalidHandler)

	// Invalid handler - wrong signature
	err = q.AddHandler("invalid", func() {})
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrInvalidHandler)

	// Invalid handler - wrong first parameter
	err = q.AddHandler("invalid", func(s string, payload testEmailPayload) error {
		return nil
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrInvalidHandler)

	// Invalid handler - wrong return type
	err = q.AddHandler("invalid", func(ctx context.Context, payload testEmailPayload) string {
		return ""
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrInvalidHandler)
}

func TestSimpleQueue_Enqueue(t *testing.T) {
	storage := memory.New()
	q := queue.New(storage)

	// Register handler
	handlerCalled := false
	handlerPayload := testEmailPayload{}
	err := q.AddHandler("send_email", func(ctx context.Context, payload testEmailPayload) error {
		handlerCalled = true
		handlerPayload = payload
		return nil
	})
	require.NoError(t, err)

	// Create a WaitGroup for the queue goroutine
	var queueWg sync.WaitGroup
	queueWg.Add(1)

	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer queueWg.Done()
		err := q.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("Error running queue: %v", err)
		}
	}()

	// Allow time for queue to start
	time.Sleep(100 * time.Millisecond)

	// Enqueue a job
	payload := testEmailPayload{
		To:      "test@example.com",
		Subject: "Test Subject",
		Body:    "Test Body",
	}
	jobID, err := q.Enqueue(ctx, "send_email", payload)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)

	// Allow time for job to process
	time.Sleep(200 * time.Millisecond)

	// Verify handler was called with correct payload
	assert.True(t, handlerCalled)
	assert.Equal(t, payload.To, handlerPayload.To)
	assert.Equal(t, payload.Subject, handlerPayload.Subject)
	assert.Equal(t, payload.Body, handlerPayload.Body)

	// Try to enqueue a job with unknown task
	_, err = q.Enqueue(ctx, "unknown_task", payload)
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrHandlerNotFound)

	// Stop the queue and wait for goroutines to finish
	cancel()
	queueWg.Wait()
}

func TestSimpleQueue_EnqueueIn(t *testing.T) {
	storage := memory.New()
	q := queue.New(storage)

	// Register handler
	var handleCalled bool
	var mu sync.Mutex // Protect handleCalled variable
	var wg sync.WaitGroup
	wg.Add(1)

	err := q.AddHandler("delayed_task", func(ctx context.Context, payload testEmailPayload) error {
		mu.Lock()
		handleCalled = true
		mu.Unlock()
		wg.Done()
		return nil
	})
	require.NoError(t, err)

	// Create a WaitGroup for the queue goroutine
	var queueWg sync.WaitGroup
	queueWg.Add(1)

	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer queueWg.Done()
		err := q.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("Error running queue: %v", err)
		}
	}()

	// Allow time for queue to start
	time.Sleep(100 * time.Millisecond)

	// Test with a short delay
	delay := 300 * time.Millisecond
	startTime := time.Now()

	jobID, err := q.EnqueueIn(ctx, "delayed_task", testEmailPayload{}, delay)
	require.NoError(t, err)
	require.NotEmpty(t, jobID)

	// Task should not have been processed yet
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	assert.False(t, handleCalled, "Task should not be processed before delay")
	mu.Unlock()

	// Wait for the job to be processed
	wg.Wait()
	elapsedTime := time.Since(startTime)

	// Verify the delay was respected
	assert.GreaterOrEqual(t, elapsedTime, delay, "Job was processed before delay elapsed")

	// Stop the queue and wait for goroutines to finish
	cancel()
	queueWg.Wait()
}

func TestSimpleQueue_RunAndStop(t *testing.T) {
	storage := memory.New()
	q := queue.New(storage)

	// Add a test handler to make the queue process something
	err := q.AddHandler("test_task", func(ctx context.Context, payload testEmailPayload) error {
		return nil
	})
	require.NoError(t, err)

	// Create a WaitGroup for the queue goroutine
	var queueWg sync.WaitGroup
	queueWg.Add(1)

	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer queueWg.Done()
		err := q.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("Error running queue: %v", err)
		}
	}()

	// Give the queue time to actually start
	time.Sleep(200 * time.Millisecond)

	// Verify it's running by trying to run it again
	err = q.Run(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrQueueAlreadyRunning)

	// Now try to stop it
	err = q.Stop(context.Background())
	require.NoError(t, err)

	// Cancel the context and wait for the goroutine to finish
	cancel()
	queueWg.Wait()

	// Verify it's stopped
	err = q.Stop(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, queue.ErrQueueNotRunning)
}

func TestSimpleQueue_Retries(t *testing.T) {
	// Create a test-specific storage implementation that we can directly inspect
	storage := memory.New()
	
	// Use a shorter retry delay for faster testing
	retryFunc := func(retryCount int) time.Duration {
		return 50 * time.Millisecond
	}
	
	// Create the queue with our test configuration
	q := queue.New(storage, queue.WithRetryDelayFunc(retryFunc))
	
	// Create an atomic counter to track calls to our handler
	var attemptCount int32
	
	// Register a simple handler that will fail on the first two attempts
	err := q.AddHandler("retry_test", func(ctx context.Context, payload testEmailPayload) error {
		// Increment the attempt count
		attemptCount++
		
		// Fail on first and second attempts
		if attemptCount <= 2 {
			t.Logf("Handler attempt %d: returning error", attemptCount)
			return errors.New("test error")
		}
		
		t.Logf("Handler attempt %d: successful", attemptCount)
		return nil
	})
	require.NoError(t, err)
	
	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go func() {
		if err := q.Run(ctx); err != nil && err != context.Canceled {
			t.Logf("Queue error: %v", err)
		}
	}()
	
	// Wait for queue to start
	time.Sleep(200 * time.Millisecond)
	
	// Submit a job
	jobID, err := q.Enqueue(ctx, "retry_test", testEmailPayload{
		To:      "user@example.com",
		Subject: "Retry Test",
		Body:    "This is a test for the retry functionality",
	})
	require.NoError(t, err)
	
	// Create a timeout context for our test
	testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer testCancel()
	
	// Wait for 3 handler calls or timeout
	for atomic.LoadInt32(&attemptCount) < 3 {
		select {
		case <-testCtx.Done():
			// Get the job status to help debug
			job, err := storage.Get(context.Background(), jobID)
			if err == nil {
				t.Logf("Current job status: %s, retry count: %d", job.Status, job.RetryCount)
			}
			t.Fatalf("Test timed out waiting for all retries. Only saw %d calls", atomic.LoadInt32(&attemptCount))
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	
	// Verify we got all 3 expected calls
	assert.Equal(t, int32(3), atomic.LoadInt32(&attemptCount), "Handler should have been called exactly 3 times")
	
	// Optional: verify the job ended up in the completed state
	job, err := storage.Get(context.Background(), jobID)
	require.NoError(t, err)
	assert.Equal(t, queue.JobStatusCompleted, job.Status)
}

func TestSimpleQueue_WithMiddleware(t *testing.T) {
	storage := memory.New()

	// Create a counter for completed jobs with a mutex
	var completedJobs int
	var mu sync.Mutex
	
	// Create a channel to signal when all jobs are done
	done := make(chan struct{})
	
	// Create a test middleware that counts calls
	var middlewareCalls int
	testMiddleware := func(next queue.Handler) queue.Handler {
		return func(ctx context.Context, job *queue.Job) error {
			mu.Lock()
			middlewareCalls++
			mu.Unlock()
			return next(ctx, job)
		}
	}

	q := queue.New(storage, queue.WithMiddleware(testMiddleware))

	// Register handler
	var handlerCalls int
	err := q.AddHandler("middleware_test", func(ctx context.Context, payload testEmailPayload) error {
		mu.Lock()
		handlerCalls++
		completedJobs++
		
		// Check if all jobs are complete
		if completedJobs >= 3 {
			defer close(done)
		}
		mu.Unlock()
		
		return nil
	})
	require.NoError(t, err)

	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Create a separate context with timeout for the test
	testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer testCancel()

	go func() {
		err := q.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("Error running queue: %v", err)
		}
	}()

	// Allow time for queue to start
	time.Sleep(100 * time.Millisecond)

	// Enqueue multiple jobs
	for i := 0; i < 3; i++ {
		jobID, err := q.Enqueue(ctx, "middleware_test", testEmailPayload{})
		require.NoError(t, err)
		require.NotEmpty(t, jobID)
	}

	// Wait for all jobs to be processed or timeout
	select {
	case <-done:
		// All jobs completed successfully
	case <-testCtx.Done():
		t.Fatal("Test timed out waiting for jobs to complete")
	}

	// Stop the queue
	cancel()

	// Verify call counts
	mu.Lock()
	finalHandlerCalls := handlerCalls
	finalMiddlewareCalls := middlewareCalls
	mu.Unlock()

	// Both handler and middleware should have been called 3 times
	assert.Equal(t, 3, finalHandlerCalls)
	assert.Equal(t, 3, finalMiddlewareCalls)
}

func TestSimpleQueue_WithConcurrency(t *testing.T) {
	storage := memory.New()

	// Create a queue with high concurrency
	q := queue.New(storage, queue.WithConcurrency(10))

	// Create channels to track job processing with proper buffer size
	processingChan := make(chan struct{}, 10)
	completedChan := make(chan struct{}, 10)
	
	// Create a separate context with timeout for the test
	testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer testCancel()

	err := q.AddHandler("slow_task", func(ctx context.Context, payload testEmailPayload) error {
		processingChan <- struct{}{}
		// Simulate some work
		time.Sleep(200 * time.Millisecond)
		completedChan <- struct{}{}
		return nil
	})
	require.NoError(t, err)

	// Start the queue
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := q.Run(ctx)
		if err != nil && err != context.Canceled {
			t.Logf("Error running queue: %v", err)
		}
	}()

	// Allow time for queue to start
	time.Sleep(100 * time.Millisecond)

	// Enqueue multiple jobs all at once
	startTime := time.Now()
	jobCount := 5

	for i := 0; i < jobCount; i++ {
		_, err := q.Enqueue(ctx, "slow_task", testEmailPayload{})
		require.NoError(t, err)
	}

	// Wait for all jobs to start processing with timeout
	for i := 0; i < jobCount; i++ {
		select {
		case <-processingChan:
			// Job is processing
		case <-testCtx.Done():
			t.Fatal("Test timed out waiting for jobs to start processing")
		}
	}

	// Wait for all jobs to complete with timeout
	for i := 0; i < jobCount; i++ {
		select {
		case <-completedChan:
			// Job is complete
		case <-testCtx.Done():
			t.Fatal("Test timed out waiting for jobs to complete")
		}
	}

	// Stop the queue
	cancel()

	duration := time.Since(startTime)

	// If processing was truly parallel, total time should be close to the duration
	// of a single job (200ms) plus some overhead, not 5 * 200ms = 1000ms
	assert.Less(t, duration, 700*time.Millisecond, "Jobs did not appear to execute in parallel, took %v", duration)
}
