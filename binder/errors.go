package binder

import "errors"

var (
	// ErrInvalidContentType is returned when the content type is not supported
	ErrInvalidContentType = errors.New("invalid content type")
	// ErrEmptyBody is returned when the request body is empty
	ErrEmptyBody = errors.New("empty body")
	// ErrInvalidJSON is returned when the request body is not valid JSON
	ErrInvalidJSON = errors.New("invalid JSON format")
	// ErrInvalidQueryParams is returned when the query parameters cannot be bound
	ErrInvalidQueryParams = errors.New("invalid query parameters")
	// ErrInvalidFormData is returned when the form data cannot be bound
	ErrInvalidFormData = errors.New("invalid form data")
	// ErrUnsupportedType is returned when the target type is not supported
	ErrUnsupportedType = errors.New("unsupported target type")
	// ErrInvalidRequest is returned when the request is nil or invalid
	ErrInvalidRequest = errors.New("invalid request")
)
