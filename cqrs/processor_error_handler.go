package cqrs

import (
	"context"
	"log/slog"
)

// UNKNOWN is the default error reason string.
// It is used when the error is not found in the errorReasons map.
const UNKNOWN = "UNKNOWN"

// ErrorIsNil is the error reason string for nil errors.
const ErrorIsNil = "ERROR_IS_NIL"

// ProcessorErrorsHandler is a function that handles errors in the processor.
// It publishes an error event to the event bus if the error is not nil.
// The error event contains the error message and the error reason.
// The error reason is used to categorize the error. It is used to determine
// the type of error that occurred.
func ProcessorErrorsHandler(
	eb interface {
		Publish(ctx context.Context, event any) error
	},
	log *slog.Logger,
	errorReason func(error) string,
) func(ctx context.Context, err error) error {
	return func(ctx context.Context, err error) error {
		if err == nil {
			return nil
		}

		if errorReason == nil {
			errorReason = DefaultErrorReason
		}

		reason := errorReason(err)
		if reason == UNKNOWN {
			return err
		}

		// Publish the error event
		if err := eb.Publish(ctx, NewErrorMessage(err, reason)); err != nil {
			log.ErrorContext(ctx, "Failed to publish error event", "error", err, "reason", reason)
		}
		return nil
	}
}

// DefaultErrorReason is the default implementation of errorReason.
// It returns a string that identifies the type/reason of the error.
// If no reason is found, it returns UNKNOWN.
func DefaultErrorReason(err error) string {
	if err == nil {
		return ErrorIsNil
	}
	return UNKNOWN
}
