package queue

import "time"

// Config holds configuration options for the queue.
type Config struct {
	// Concurrency is the number of worker goroutines to start.
	// Default is 10.
	Concurrency int

	// RetryDelayFunc is a function that calculates the delay before retrying a failed job.
	// If not provided, DefaultRetryDelayFunc will be used.
	RetryDelayFunc RetryDelayFunc

	// MaxRetries is the default maximum number of retries for jobs.
	// Default is 3.
	MaxRetries int

	// PollInterval is the time between polling the storage for new jobs.
	// Default is 100ms.
	PollInterval time.Duration

	// JobTimeout is the maximum time a job can run before it is considered failed.
	// Default is 30 minutes.
	JobTimeout time.Duration

	// ShutdownTimeout is the maximum time to wait for jobs to complete during shutdown.
	// Default is 30 seconds.
	ShutdownTimeout time.Duration

	// Logger is an optional logger for queue operations.
	// If not provided, logging will be disabled.
	// Logger Logger
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() Config {
	return Config{
		Concurrency:     10,
		RetryDelayFunc:  DefaultRetryDelayFunc,
		MaxRetries:      3,
		PollInterval:    100 * time.Millisecond,
		JobTimeout:      30 * time.Minute,
		ShutdownTimeout: 30 * time.Second,
	}
}
