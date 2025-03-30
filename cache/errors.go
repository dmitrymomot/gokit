package cache

import "errors"

// Predefined errors for the cache package.
var (
	// ErrNotFound indicates that the requested key was not found in the cache.
	ErrNotFound = errors.New("cache: key not found")

	// ErrEncoding indicates an error during encoding the cache value.
	ErrEncoding = errors.New("cache: failed to encode value")

	// ErrDecoding indicates an error during decoding the cached value.
	ErrDecoding = errors.New("cache: failed to decode value")

	// ErrConnectionFailed indicates a failure to connect to the cache backend.
	ErrConnectionFailed = errors.New("cache: connection failed")

	// ErrOperationFailed indicates a general failure during a cache operation (Set, Delete, etc.).
	ErrOperationFailed = errors.New("cache: operation failed")

	// ErrNotImplemented indicates a feature is not implemented by the specific cache adapter.
	ErrNotImplemented = errors.New("cache: feature not implemented")
)
