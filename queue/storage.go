package queue

import (
	"context"
	"time"
)

// Storage defines the interface for queue storage backends.
// This interface can be implemented for different storage solutions
// like in-memory, Redis, MongoDB, etc. Storage implementations must
// be thread-safe as they will be accessed concurrently from multiple goroutines.
type Storage interface {
	// Ping checks if the storage is available.
	// It should return nil if the storage is operational or an error otherwise.
	Ping(ctx context.Context) error

	// Put stores a job in the storage.
	// It should generate an ID for the job if one is not already set.
	// Returns an error if the storage operation fails.
	Put(ctx context.Context, job *Job) error

	// Get retrieves a job by ID.
	// Returns the job if found, or ErrJobNotFound if not found.
	// May return other errors if the storage operation fails.
	Get(ctx context.Context, id string) (*Job, error)

	// Update updates a job in the storage.
	// Returns ErrJobNotFound if the job doesn't exist or other errors
	// if the storage operation fails.
	Update(ctx context.Context, job *Job) error

	// Delete removes a job from the storage.
	// Returns ErrJobNotFound if the job doesn't exist or other errors
	// if the storage operation fails.
	Delete(ctx context.Context, id string) error

	// FetchDue retrieves due jobs ready for processing,
	// up to the specified limit, marking them as processing.
	// Jobs are considered due when their status is pending and their
	// RunAt time is in the past. The implementation should atomically
	// update the status to prevent the same job from being processed
	// by multiple workers.
	FetchDue(ctx context.Context, limit int) ([]*Job, error)

	// FetchByStatus retrieves jobs with the specified status,
	// up to the specified limit.
	// This is useful for monitoring and maintenance operations.
	FetchByStatus(ctx context.Context, status JobStatus, limit int) ([]*Job, error)

	// PurgeCompleted removes completed jobs older than the specified duration.
	// This helps manage storage size by removing jobs that have been
	// successfully processed and are no longer needed.
	PurgeCompleted(ctx context.Context, olderThan time.Duration) error

	// PurgeFailed removes failed jobs older than the specified duration.
	// This helps manage storage size by removing jobs that have failed
	// and are no longer needed.
	PurgeFailed(ctx context.Context, olderThan time.Duration) error

	// Size returns the total number of jobs in the storage.
	// This is useful for monitoring and diagnostic purposes.
	Size(ctx context.Context) (int, error)

	// Close closes the storage connection.
	// Any resources used by the storage should be released.
	// After Close is called, no other methods should be called on the storage.
	Close(ctx context.Context) error
}
