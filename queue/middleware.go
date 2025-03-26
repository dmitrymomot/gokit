// Package queue provides a concurrent job processing system with support for
// type-safe handlers, job retries, and delayed job execution.
package queue

import (
	"context"
	"log/slog"
	"time"
)

// Middleware defines a function that wraps a job handler to add functionality
// before or after the handler execution. Middleware can be used for logging,
// metrics, tracing, error handling, etc.
type Middleware func(handler Handler) Handler

// Chain combines multiple middleware into a single middleware.
// The first middleware in the chain will be executed first (outermost),
// followed by the second, and so on.
func Chain(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// WithLogging returns a middleware that logs job execution details.
// It logs when a job starts processing, when it completes, and when it fails.
func WithLogging(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next Handler) Handler {
		return func(ctx context.Context, job *Job) (err error) {
			startTime := time.Now()

			logger.InfoContext(ctx, "Processing job",
				"job_id", job.ID,
				"task", job.TaskName,
				"retry_count", job.RetryCount,
			)

			defer func() {
				duration := time.Since(startTime)
				if err != nil {
					logger.ErrorContext(ctx, "Job failed",
						"job_id", job.ID,
						"task", job.TaskName,
						"error", err.Error(),
						"duration_ms", duration.Milliseconds(),
						"retry_count", job.RetryCount,
					)
				} else {
					logger.InfoContext(ctx, "Job completed",
						"job_id", job.ID,
						"task", job.TaskName,
						"duration_ms", duration.Milliseconds(),
						"retry_count", job.RetryCount,
					)
				}
			}()

			return next(ctx, job)
		}
	}
}

// WithMetrics returns a middleware that records job execution metrics.
// A MetricsRecorder interface must be provided to record the metrics.
// This is just a skeleton - users would need to implement the MetricsRecorder interface.
type MetricsRecorder interface {
	RecordJobStart(ctx context.Context, taskName string)
	RecordJobCompletion(ctx context.Context, taskName string, duration time.Duration)
	RecordJobFailure(ctx context.Context, taskName string, err error, duration time.Duration)
}

// WithMetrics returns a middleware that records job execution metrics.
func WithMetrics(recorder MetricsRecorder) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job *Job) (err error) {
			startTime := time.Now()
			recorder.RecordJobStart(ctx, job.TaskName)

			defer func() {
				duration := time.Since(startTime)
				if err != nil {
					recorder.RecordJobFailure(ctx, job.TaskName, err, duration)
				} else {
					recorder.RecordJobCompletion(ctx, job.TaskName, duration)
				}
			}()

			return next(ctx, job)
		}
	}
}

// WithRecovery returns a middleware that recovers from panics in job handlers.
// When a panic occurs, it is converted to an error and returned.
func WithRecovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job *Job) (err error) {
			defer func() {
				if r := recover(); r != nil {
					// Convert panic to error
					switch x := r.(type) {
					case string:
						err = ErrJobPanicked
					case error:
						err = x
					default:
						err = ErrJobPanicked
					}
				}
			}()

			return next(ctx, job)
		}
	}
}

// WithTimeout returns a middleware that sets a timeout for job execution.
// If the job takes longer than the specified timeout, it is cancelled.
func WithTimeout(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, job *Job) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			return next(timeoutCtx, job)
		}
	}
}
