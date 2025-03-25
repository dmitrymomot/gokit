package webhook

import "errors"

// Error definitions for the webhook package
var (
	// ErrInvalidURL is returned when the webhook URL is invalid or empty
	ErrInvalidURL = errors.New("invalid webhook URL")

	// ErrInvalidMethod is returned when the HTTP method is invalid
	ErrInvalidMethod = errors.New("invalid HTTP method")

	// ErrMarshalParams is returned when there's an error marshaling request parameters
	ErrMarshalParams = errors.New("failed to marshal request parameters")

	// ErrCreateRequest is returned when there's an error creating the HTTP request
	ErrCreateRequest = errors.New("failed to create HTTP request")

	// ErrSendRequest is returned when there's an error sending the HTTP request
	ErrSendRequest = errors.New("failed to send HTTP request")

	// ErrReadResponse is returned when there's an error reading the HTTP response
	ErrReadResponse = errors.New("failed to read HTTP response")

	// ErrResponseTimeout is returned when the webhook request times out
	ErrResponseTimeout = errors.New("webhook request timed out")
)
