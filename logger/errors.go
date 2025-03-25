package logger

import "errors"

var (
	// ErrInvalidLogFormat is returned when an invalid log format is provided.
	ErrInvalidLogFormat = errors.New("invalid log format")
)
