// Package queue provides a concurrent job processing system with support for
// type-safe handlers, job retries, and delayed job execution.
package queue

import (
	"context"
	"reflect"
	"sync"
	"time"
)

// Queue defines the interface for interacting with the job queue.
// It provides methods for registering handlers, enqueueing jobs,
// and controlling the queue's lifecycle.
type Queue interface {
	// AddHandler registers a type-safe handler function for the given task name.
	// Handler function must have the signature func(context.Context, T) error
	// where T is any type that can be marshaled to and from JSON.
	AddHandler(taskName string, handlerFunc any) error

	// Enqueue adds a job to the queue for immediate processing.
	// Payload must be a type that can be marshaled to and from JSON.
	// Returns the job ID and any error that occurred.
	Enqueue(ctx context.Context, taskName string, payload any) (string, error)

	// EnqueueIn adds a job to the queue to be processed after the specified delay.
	// Payload must be a type that can be marshaled to and from JSON.
	// Returns the job ID and any error that occurred.
	EnqueueIn(ctx context.Context, taskName string, payload any, delay time.Duration) (string, error)

	// Run starts processing jobs from the queue. This method is blocking
	// and should be started in a goroutine.
	// It will continue processing jobs until the context is cancelled.
	Run(ctx context.Context) error

	// Stop gracefully stops the queue, allowing in-progress jobs to complete
	// but not processing any new jobs.
	// Returns an error if the queue could not be stopped.
	Stop(ctx context.Context) error
}

// SimpleQueue is the default implementation of the Queue interface.
// It provides a concurrent job processing system with support for
// type-safe handlers, job retries, and delayed job execution.
type SimpleQueue struct {
	storage        Storage           // Storage backend for jobs
	handlers       map[string]Handler // Registered handlers
	concurrency    int                // Number of worker goroutines
	running        bool               // Whether the queue is currently running
	mu             sync.RWMutex       // Lock for thread safety
	stopChan       chan struct{}      // Channel to signal workers to stop
	wg             sync.WaitGroup     // WaitGroup for worker goroutines
	retryDelayFunc RetryDelayFunc     // Function to calculate retry delay
	middleware     Middleware         // Middleware chain to apply to all handlers
	cancel         context.CancelFunc // Function to cancel worker contexts
}

// RetryDelayFunc is a function that calculates the delay before retrying a failed job.
// It takes the number of previous retries and returns a duration to wait before the next attempt.
type RetryDelayFunc func(retryCount int) time.Duration

// DefaultRetryDelayFunc returns a default exponential backoff function
// with an initial delay of 1 second and a multiplier of 2.
// This creates delays of 1s, 2s, 4s, 8s, 16s, etc. for sequential retries.
func DefaultRetryDelayFunc(retryCount int) time.Duration {
	// Exponential backoff: 1s, 2s, 4s, 8s, 16s, etc.
	return time.Duration(1<<uint(retryCount)) * time.Second
}

// Option is a function that configures a SimpleQueue.
// This follows the functional options pattern for configuring structs.
type Option func(*SimpleQueue)

// WithConcurrency sets the number of worker goroutines.
// If the value is less than or equal to zero, the default concurrency will be used.
func WithConcurrency(concurrency int) Option {
	return func(q *SimpleQueue) {
		if concurrency > 0 {
			q.concurrency = concurrency
		}
	}
}

// WithRetryDelayFunc sets the function used to calculate retry delays.
// If nil is provided, the default retry delay function will be used.
func WithRetryDelayFunc(fn RetryDelayFunc) Option {
	return func(q *SimpleQueue) {
		if fn != nil {
			q.retryDelayFunc = fn
		}
	}
}

// WithMiddleware sets the middleware to be applied to all handlers.
// If multiple middleware are provided, they will be chained together.
func WithMiddleware(middleware ...Middleware) Option {
	return func(q *SimpleQueue) {
		if len(middleware) == 1 {
			q.middleware = middleware[0]
		} else if len(middleware) > 1 {
			q.middleware = Chain(middleware...)
		}
	}
}

// New creates a new SimpleQueue with the given storage and options.
// The storage parameter is required and must not be nil.
// Optional configurations can be provided through Option functions.
func New(storage Storage, opts ...Option) *SimpleQueue {
	queue := &SimpleQueue{
		storage:        storage,
		handlers:       make(map[string]Handler),
		concurrency:    10, // Default concurrency
		stopChan:       make(chan struct{}),
		retryDelayFunc: DefaultRetryDelayFunc,
	}

	// Apply options
	for _, opt := range opts {
		opt(queue)
	}

	return queue
}

// AddHandler registers a type-safe handler function for the given task name.
// The handler function must have the signature func(context.Context, T) error
// where T is any type that can be marshaled to and from JSON.
// Returns an error if the handler function has an invalid signature.
func (q *SimpleQueue) AddHandler(taskName string, handlerFunc any) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Validate handler function
	handlerType := reflect.TypeOf(handlerFunc)
	if handlerType.Kind() != reflect.Func {
		return ErrInvalidHandler
	}

	if handlerType.NumIn() != 2 || handlerType.NumOut() != 1 {
		return ErrInvalidHandler
	}

	// First parameter must be context.Context
	if !handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return ErrInvalidHandler
	}

	// Return type must be error
	if !handlerType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return ErrInvalidHandler
	}

	// Create adapter function that converts from Job to the specific payload type
	handlerValue := reflect.ValueOf(handlerFunc)
	payloadType := handlerType.In(1)

	wrapper := func(ctx context.Context, job *Job) error {
		// Create a new instance of the payload type
		payload := reflect.New(payloadType).Interface()

		// Deserialize the job payload into the payload instance
		if err := job.GetPayload(payload); err != nil {
			return err
		}

		// Call the handler with the deserialized payload (get the Elem since we created with New)
		args := []reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(payload).Elem(),
		}

		results := handlerValue.Call(args)
		if results[0].IsNil() {
			return nil
		}

		return results[0].Interface().(error)
	}

	// Apply middleware if configured
	handler := wrapper
	if q.middleware != nil {
		handler = q.middleware(wrapper)
	}

	// Store the wrapper function in the handlers map
	q.handlers[taskName] = handler
	return nil
}

// Enqueue adds a job to the queue for immediate processing.
// The payload must be a type that can be marshaled to JSON.
// Returns the job ID and any error that occurred.
// If the queue is not running, ErrQueueClosed will be returned.
// If no handler is registered for the task name, ErrHandlerNotFound will be returned.
func (q *SimpleQueue) Enqueue(ctx context.Context, taskName string, payload any) (string, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if !q.running {
		return "", ErrQueueClosed
	}

	// Check if handler exists for this task
	if _, exists := q.handlers[taskName]; !exists {
		return "", ErrHandlerNotFound
	}

	// Create a new job
	job, err := NewJob(taskName, payload)
	if err != nil {
		return "", err
	}

	// Save the job in storage
	if err := q.storage.Put(ctx, job); err != nil {
		return "", err
	}

	return job.ID, nil
}

// EnqueueIn adds a job to the queue to be processed after the specified delay.
// The payload must be a type that can be marshaled to JSON.
// Returns the job ID and any error that occurred.
// If the queue is not running, ErrQueueClosed will be returned.
// If no handler is registered for the task name, ErrHandlerNotFound will be returned.
func (q *SimpleQueue) EnqueueIn(ctx context.Context, taskName string, payload any, delay time.Duration) (string, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if !q.running {
		return "", ErrQueueClosed
	}

	// Check if handler exists for this task
	if _, exists := q.handlers[taskName]; !exists {
		return "", ErrHandlerNotFound
	}

	// Create a new job
	job, err := NewJob(taskName, payload)
	if err != nil {
		return "", err
	}

	// Set the job to run in the future
	job.RunAt = time.Now().Add(delay)

	// Save the job in storage
	if err := q.storage.Put(ctx, job); err != nil {
		return "", err
	}

	return job.ID, nil
}

// Run starts processing jobs from the queue.
// This method is blocking and should be started in a goroutine.
// It will continue processing jobs until the context is cancelled.
// Returns an error if the queue is already running.
func (q *SimpleQueue) Run(ctx context.Context) error {
	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return ErrQueueAlreadyRunning
	}
	
	q.running = true
	q.stopChan = make(chan struct{})
	
	// Create a cancellable context for the workers
	ctx, cancel := context.WithCancel(ctx)
	q.cancel = cancel
	
	q.mu.Unlock()

	// Start the worker goroutines
	q.startWorkers(ctx)

	// Wait for the context to be cancelled
	<-ctx.Done()
	return ctx.Err()
}

// Stop gracefully stops the queue.
// It allows in-progress jobs to complete but does not process any new jobs.
// Returns an error if the queue is not running or if the stop operation times out.
func (q *SimpleQueue) Stop(ctx context.Context) error {
	q.mu.Lock()
	if !q.running {
		q.mu.Unlock()
		return ErrQueueNotRunning
	}
	
	q.running = false
	close(q.stopChan)
	
	// Cancel the worker contexts if cancel function exists
	if q.cancel != nil {
		q.cancel()
	}
	
	q.mu.Unlock()

	// Wait for all workers to finish
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	// Wait for workers to finish or context deadline
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// startWorkers starts the worker goroutines.
func (q *SimpleQueue) startWorkers(ctx context.Context) {
	for i := 0; i < q.concurrency; i++ {
		q.wg.Add(1)
		go q.worker(ctx)
	}
}

// worker is a goroutine that processes jobs.
func (q *SimpleQueue) worker(ctx context.Context) {
	defer q.wg.Done()

	for {
		select {
		case <-q.stopChan:
			// Queue has been explicitly stopped
			return
		case <-ctx.Done():
			// Context has been cancelled (either by Stop or parent context)
			return
		default:
			// Check if queue is still running
			q.mu.RLock()
			isRunning := q.running
			q.mu.RUnlock()
			
			if !isRunning {
				return
			}

			// Fetch jobs ready for processing
			jobs, err := q.storage.FetchDue(ctx, 1)
			if err != nil {
				// Check if context was cancelled during fetch
				select {
				case <-ctx.Done():
					return
				default:
					// Log error and continue
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}

			if len(jobs) == 0 {
				// No jobs to process, sleep briefly to avoid CPU spinning
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}

			// Process the job
			job := jobs[0]
			q.processJob(ctx, job)
		}
	}
}

// processJob processes a single job and handles retries.
func (q *SimpleQueue) processJob(ctx context.Context, job *Job) {
	// Get handler for this task
	q.mu.RLock()
	handler, exists := q.handlers[job.TaskName]
	q.mu.RUnlock()

	if !exists {
		// Update job status to failed
		job.Status = JobStatusFailed
		job.LastError = ErrHandlerNotFound.Error()
		job.UpdatedAt = time.Now()
		_ = q.storage.Update(ctx, job)
		return
	}

	// Execute the handler
	err := handler(ctx, job)

	if err != nil {
		// Handle failure
		job.RetryCount++
		job.LastError = err.Error()
		job.UpdatedAt = time.Now()

		if job.ShouldRetry() {
			// Schedule for retry
			job.Status = JobStatusRetrying
			job.RunAt = time.Now().Add(q.retryDelayFunc(job.RetryCount))
		} else {
			// Mark as failed
			job.Status = JobStatusFailed
		}
	} else {
		// Mark as completed
		job.Status = JobStatusCompleted
		job.UpdatedAt = time.Now()
	}

	// Update job in storage
	_ = q.storage.Update(ctx, job)
}
