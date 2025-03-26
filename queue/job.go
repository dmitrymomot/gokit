package queue

import (
	"context"
	"encoding/json"
	"time"
)

// JobStatus represents the status of a job in the queue.
type JobStatus string

const (
	// JobStatusPending indicates the job is waiting to be processed.
	JobStatusPending JobStatus = "pending"
	// JobStatusProcessing indicates the job is currently being processed.
	JobStatusProcessing JobStatus = "processing"
	// JobStatusCompleted indicates the job has been successfully processed.
	JobStatusCompleted JobStatus = "completed"
	// JobStatusFailed indicates the job has failed and won't be retried.
	JobStatusFailed JobStatus = "failed"
	// JobStatusRetrying indicates the job has failed and is scheduled for retry.
	JobStatusRetrying JobStatus = "retrying"
)

// Job represents a task to be processed by the queue.
type Job struct {
	// ID is the unique identifier for the job.
	ID string `json:"id"`
	// TaskName is the name of the task, used to route the job to the correct handler.
	TaskName string `json:"task_name"`
	// Payload is the data to be processed by the job handler.
	Payload []byte `json:"payload"`
	// Status is the current status of the job.
	Status JobStatus `json:"status"`
	// CreatedAt is the time the job was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the time the job was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// RunAt is the time the job should be run.
	RunAt time.Time `json:"run_at"`
	// RetryCount is the number of times the job has been retried.
	RetryCount int `json:"retry_count"`
	// MaxRetries is the maximum number of times the job should be retried before failing.
	MaxRetries int `json:"max_retries"`
	// LastError is the last error encountered while processing the job.
	LastError string `json:"last_error,omitempty"`
}

// NewJob creates a new job with the given task name and payload.
func NewJob(taskName string, payload any) (*Job, error) {
	serializedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Job{
		TaskName:   taskName,
		Payload:    serializedPayload,
		Status:     JobStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		RunAt:      now,
		RetryCount: 0,
		MaxRetries: 3, // Default max retries
	}, nil
}

// GetPayload deserializes the job payload into the provided value.
func (j *Job) GetPayload(value any) error {
	if err := json.Unmarshal(j.Payload, value); err != nil {
		return ErrInvalidJobPayload
	}
	return nil
}

// Clone returns a deep copy of the job.
func (j *Job) Clone() *Job {
	payload := make([]byte, len(j.Payload))
	copy(payload, j.Payload)

	return &Job{
		ID:         j.ID,
		TaskName:   j.TaskName,
		Payload:    payload,
		Status:     j.Status,
		CreatedAt:  j.CreatedAt,
		UpdatedAt:  j.UpdatedAt,
		RunAt:      j.RunAt,
		RetryCount: j.RetryCount,
		MaxRetries: j.MaxRetries,
		LastError:  j.LastError,
	}
}

// ShouldRetry returns true if the job should be retried.
func (j *Job) ShouldRetry() bool {
	return j.RetryCount < j.MaxRetries
}

// IsReady returns true if the job is ready to be processed.
func (j *Job) IsReady() bool {
	return j.Status == JobStatusPending && time.Now().After(j.RunAt)
}

// Handler defines the function signature for job handlers.
// It is used internally to represent registered handler functions.
type Handler func(ctx context.Context, job *Job) error

// HandlerFunc is a generic type that represents a type-safe handler function.
// The function takes a context and a strongly-typed payload parameter.
type HandlerFunc[T any] func(ctx context.Context, payload T) error
