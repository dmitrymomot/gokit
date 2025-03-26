package queue

import (
	"context"
	"time"
)

// Storage defines the interface for queue storage backends.
// This interface can be implemented for different storage solutions
// like in-memory, Redis, MongoDB, etc.
type Storage interface {
	// Ping checks if the storage is available.
	Ping(ctx context.Context) error

	// Put stores a job in the storage.
	Put(ctx context.Context, job *Job) error

	// Get retrieves a job by ID.
	Get(ctx context.Context, id string) (*Job, error)

	// Update updates a job in the storage.
	Update(ctx context.Context, job *Job) error

	// Delete removes a job from the storage.
	Delete(ctx context.Context, id string) error

	// FetchDue retrieves due jobs ready for processing,
	// up to the specified limit, marking them as processing.
	FetchDue(ctx context.Context, limit int) ([]*Job, error)

	// FetchByStatus retrieves jobs with the specified status,
	// up to the specified limit.
	FetchByStatus(ctx context.Context, status JobStatus, limit int) ([]*Job, error)

	// PurgeCompleted removes completed jobs older than the specified duration.
	PurgeCompleted(ctx context.Context, olderThan time.Duration) error

	// PurgeFailed removes failed jobs older than the specified duration.
	PurgeFailed(ctx context.Context, olderThan time.Duration) error

	// Size returns the total number of jobs in the storage.
	Size(ctx context.Context) (int, error)

	// Close closes the storage connection.
	Close(ctx context.Context) error
}
