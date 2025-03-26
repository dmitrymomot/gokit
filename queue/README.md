# Queue

A simple, developer-friendly queue system for Go applications with support for multiple storage backends and worker pools.

## Features

- Type-safe job handlers with Go generics
- Concurrent job processing with configurable worker pools
- Automatic job retries with exponential backoff
- Delayed job execution
- Pluggable storage backends (in-memory provided, Redis and MongoDB planned)
- Graceful shutdown and error handling

## Installation

```bash
go get github.com/dmitrymomot/gokit/queue
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/dmitrymomot/gokit/queue/memory"
)

// Define your job payload
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func main() {
	// Create a new in-memory storage
	storage := memory.New()

	// Create a new queue with 5 workers
	q := queue.New(storage, queue.WithConcurrency(5))

	// Register a handler for "send_email" task
	err := q.AddHandler("send_email", func(ctx context.Context, payload EmailPayload) error {
		fmt.Printf("Sending email to %s with subject: %s\n", payload.To, payload.Subject)
		// Actual email sending logic would go here
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to register handler: %v", err)
	}

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start processing jobs in a goroutine
	go func() {
		if err := q.Run(ctx); err != nil {
			log.Printf("Queue stopped with error: %v", err)
		}
	}()

	// Enqueue a job for immediate processing
	jobID, err := q.Enqueue(ctx, "send_email", EmailPayload{
		To:      "user@example.com",
		Subject: "Hello from Queue",
		Body:    "This is a test email",
	})
	if err != nil {
		log.Fatalf("Failed to enqueue job: %v", err)
	}
	fmt.Printf("Enqueued job with ID: %s\n", jobID)

	// Enqueue a job to be processed after 5 seconds
	jobID, err = q.EnqueueIn(ctx, "send_email", EmailPayload{
		To:      "delayed@example.com",
		Subject: "Delayed Email",
		Body:    "This email was delayed by 5 seconds",
	}, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to enqueue delayed job: %v", err)
	}
	fmt.Printf("Enqueued delayed job with ID: %s\n", jobID)

	// Wait for jobs to complete
	time.Sleep(10 * time.Second)

	// Gracefully shutdown the queue
	if err := q.Stop(context.Background()); err != nil {
		log.Printf("Error shutting down queue: %v", err)
	}
}
```

## Advanced Usage

### Custom Retry Strategy

You can customize the retry behavior by providing a custom retry delay function:

```go
// Custom exponential backoff with initial delay of 500ms and max delay of 1 hour
retryFn := func(retryCount int) time.Duration {
    delay := time.Duration(500*math.Pow(2, float64(retryCount))) * time.Millisecond
    if delay > time.Hour {
        return time.Hour
    }
    return delay
}

// Create queue with custom retry function
q := queue.New(storage, queue.WithRetryDelayFunc(retryFn))
```

### Multiple Handlers

You can register multiple handlers for different types of tasks:

```go
// Email notification handler
q.AddHandler("send_email", func(ctx context.Context, payload EmailPayload) error {
    // Send email
    return nil
})

// SMS notification handler
type SMSPayload struct {
    To      string `json:"to"`
    Message string `json:"message"`
}

q.AddHandler("send_sms", func(ctx context.Context, payload SMSPayload) error {
    // Send SMS
    return nil
})
```

## Storage Backends

### In-Memory Storage

The in-memory storage is suitable for development, testing, or small applications. Jobs are stored in memory and will be lost if the application restarts.

```go
storage := memory.New()
q := queue.New(storage)
```

### Redis Storage (Planned)

Redis storage will be implemented in a future release.

### MongoDB Storage (Planned)

MongoDB storage will be implemented in a future release.

## Error Handling

The queue package provides several error types to help with error handling:

- `ErrQueueClosed`: Returned when trying to enqueue a job to a closed queue
- `ErrQueueAlreadyRunning`: Returned when trying to start an already running queue
- `ErrQueueNotRunning`: Returned when trying to stop a queue that is not running
- `ErrHandlerNotFound`: Returned when a handler for a task is not registered
- `ErrInvalidJobPayload`: Returned when a job payload cannot be deserialized
- `ErrStorageUnavailable`: Returned when the storage backend is unavailable
- `ErrJobNotFound`: Returned when a job is not found in the storage

You can check for specific errors using the `errors.Is` function:

```go
_, err := q.Enqueue(ctx, "unknown_task", payload)
if errors.Is(err, queue.ErrHandlerNotFound) {
    // Handle the specific error
}
```
