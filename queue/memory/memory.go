package memory

import (
	"context"
	"sync"
	"time"

	"github.com/dmitrymomot/gokit/queue"
	"github.com/google/uuid"
)

// Storage implements the queue.Storage interface using an in-memory map.
// This is primarily useful for testing and small applications.
// It is not recommended for production use as jobs will be lost on restart.
type Storage struct {
	jobs      map[string]*queue.Job
	mu        sync.RWMutex
	isRunning bool
}

// New creates a new in-memory storage.
func New() *Storage {
	return &Storage{
		jobs:      make(map[string]*queue.Job),
		isRunning: true,
	}
}

// Ping checks if the storage is available.
func (s *Storage) Ping(ctx context.Context) error {
	if !s.isRunning {
		return queue.ErrStorageUnavailable
	}
	return nil
}

// Put stores a job in the storage.
func (s *Storage) Put(ctx context.Context, job *queue.Job) error {
	if !s.isRunning {
		return queue.ErrStorageUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate a new ID if not set
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	// Store a copy of the job
	s.jobs[job.ID] = job.Clone()
	return nil
}

// Get retrieves a job by ID.
func (s *Storage) Get(ctx context.Context, id string) (*queue.Job, error) {
	if !s.isRunning {
		return nil, queue.ErrStorageUnavailable
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	job, exists := s.jobs[id]
	if !exists {
		return nil, queue.ErrJobNotFound
	}

	// Return a copy of the job
	return job.Clone(), nil
}

// Update updates a job in the storage.
func (s *Storage) Update(ctx context.Context, job *queue.Job) error {
	if !s.isRunning {
		return queue.ErrStorageUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[job.ID]; !exists {
		return queue.ErrJobNotFound
	}

	// Store a copy of the job
	s.jobs[job.ID] = job.Clone()
	return nil
}

// Delete removes a job from the storage.
func (s *Storage) Delete(ctx context.Context, id string) error {
	if !s.isRunning {
		return queue.ErrStorageUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.jobs[id]; !exists {
		return queue.ErrJobNotFound
	}

	delete(s.jobs, id)
	return nil
}

// FetchDue retrieves due jobs ready for processing,
// up to the specified limit, marking them as processing.
func (s *Storage) FetchDue(ctx context.Context, limit int) ([]*queue.Job, error) {
	if !s.isRunning {
		return nil, queue.ErrStorageUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var dueJobs []*queue.Job
	now := time.Now()

	// Find jobs that are ready to be processed
	for _, job := range s.jobs {
		// Check for both pending and retrying jobs that are due for processing
		if (job.Status == queue.JobStatusPending || job.Status == queue.JobStatusRetrying) && now.After(job.RunAt) {
			// Update job status to processing
			job.Status = queue.JobStatusProcessing
			job.UpdatedAt = now
            
			// Add a clone of the job to the result
			dueJobs = append(dueJobs, job.Clone())

			// Stop when we've reached the limit
			if len(dueJobs) >= limit {
				break
			}
		}
	}

	return dueJobs, nil
}

// FetchByStatus retrieves jobs with the specified status,
// up to the specified limit.
func (s *Storage) FetchByStatus(ctx context.Context, status queue.JobStatus, limit int) ([]*queue.Job, error) {
	if !s.isRunning {
		return nil, queue.ErrStorageUnavailable
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var matchingJobs []*queue.Job

	for _, job := range s.jobs {
		if job.Status == status {
			matchingJobs = append(matchingJobs, job.Clone())

			// Stop when we've reached the limit
			if len(matchingJobs) >= limit {
				break
			}
		}
	}

	return matchingJobs, nil
}

// PurgeCompleted removes completed jobs older than the specified duration.
func (s *Storage) PurgeCompleted(ctx context.Context, olderThan time.Duration) error {
	if !s.isRunning {
		return queue.ErrStorageUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	for id, job := range s.jobs {
		if job.Status == queue.JobStatusCompleted && job.UpdatedAt.Before(cutoff) {
			delete(s.jobs, id)
		}
	}

	return nil
}

// PurgeFailed removes failed jobs older than the specified duration.
func (s *Storage) PurgeFailed(ctx context.Context, olderThan time.Duration) error {
	if !s.isRunning {
		return queue.ErrStorageUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)

	for id, job := range s.jobs {
		if job.Status == queue.JobStatusFailed && job.UpdatedAt.Before(cutoff) {
			delete(s.jobs, id)
		}
	}

	return nil
}

// Size returns the total number of jobs in the storage.
func (s *Storage) Size(ctx context.Context) (int, error) {
	if !s.isRunning {
		return 0, queue.ErrStorageUnavailable
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.jobs), nil
}

// Close closes the storage connection.
func (s *Storage) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRunning = false
	s.jobs = nil
	return nil
}
