package sanitizer

import "errors"

var (
	// ErrInvalidSanitizerConfiguration is returned when a sanitizer is configured incorrectly.
	ErrInvalidSanitizerConfiguration = errors.New("invalid sanitizer configuration")
	// ErrUnknownSanitizer is returned when trying to use an unregistered sanitizer.
	ErrUnknownSanitizer = errors.New("unknown sanitizer")
)
