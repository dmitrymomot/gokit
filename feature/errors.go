package feature

import "errors"

// Predefined errors for the feature package.
var (
	// ErrFlagNotFound indicates that the requested feature flag was not found.
	ErrFlagNotFound = errors.New("feature: flag not found")

	// ErrInvalidFlag indicates that the provided flag parameters are invalid.
	ErrInvalidFlag = errors.New("feature: invalid flag parameters")

	// ErrProviderNotInitialized indicates the feature provider is not properly initialized.
	ErrProviderNotInitialized = errors.New("feature: provider not initialized")

	// ErrInvalidContext indicates the context passed does not contain required values.
	ErrInvalidContext = errors.New("feature: invalid context")

	// ErrInvalidStrategy indicates an issue with the rollout strategy configuration.
	ErrInvalidStrategy = errors.New("feature: invalid rollout strategy")

	// ErrOperationFailed indicates a general failure during an operation.
	ErrOperationFailed = errors.New("feature: operation failed")
)
