package queue

import "errors"

// Queue errors
var (
	// ErrQueueClosed is returned when trying to enqueue a job to a closed queue.
	ErrQueueClosed = errors.New("queue is closed")
	// ErrQueueAlreadyRunning is returned when trying to start an already running queue.
	ErrQueueAlreadyRunning = errors.New("queue is already running")
	// ErrQueueNotRunning is returned when trying to stop a queue that is not running.
	ErrQueueNotRunning = errors.New("queue is not running")
	// ErrHandlerNotFound is returned when a handler for a task is not registered.
	ErrHandlerNotFound = errors.New("handler not found")
	// ErrInvalidJobPayload is returned when a job payload cannot be deserialized.
	ErrInvalidJobPayload = errors.New("invalid job payload")
	// ErrStorageUnavailable is returned when the storage backend is unavailable.
	ErrStorageUnavailable = errors.New("storage is unavailable")
	// ErrJobNotFound is returned when a job is not found in the storage.
	ErrJobNotFound = errors.New("job not found")
	// ErrInvalidRetryCount is returned when a job has an invalid retry count.
	ErrInvalidRetryCount = errors.New("invalid retry count")
	// ErrUnknownJobStatus is returned when a job has an unknown status.
	ErrUnknownJobStatus = errors.New("unknown job status")
	// ErrInvalidHandler is returned when trying to register an invalid handler.
	ErrInvalidHandler = errors.New("invalid handler")
)
