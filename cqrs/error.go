package cqrs

import "net/url"

// PublicError represents a public error response
// that can be sent to the client.
type PublicError[Details any] struct {
	Code    int     `json:"code,omitempty"`
	Error   string  `json:"error"`
	Reason  string  `json:"reason,omitempty"`
	Details Details `json:"details,omitempty"`
}

// NewPublicError creates a new PublicError instance. The code is the HTTP status code, message is the
// human-readable error message, reason is the error reason, and details is an optional field that can
// contain additional error details.
func NewPublicError[Details any](code int, err error, reason string, details Details) *PublicError[Details] {
	return &PublicError[Details]{
		Code:    code,
		Error:   err.Error(),
		Reason:  reason,
		Details: details,
	}
}

// NewErrorMessage creates a new PublicError instance with the default code 1000.
// The error is the error message, and the reason is the error reason.
// The details field can contain additional error details. This function is useful
// for creating error messages that are not associated with an HTTP status code.
// For example, you can use this function to create error messages that are not
// related to HTTP status codes, such as validation errors.
func NewErrorMessage(err error, reason string, kv ...string) *PublicError[url.Values] {
	details := make(url.Values)
	for i := 0; i < len(kv)-1; i += 2 {
		details.Set(kv[i], kv[i+1])
	}
	return &PublicError[url.Values]{
		Code:    1000,
		Error:   err.Error(),
		Reason:  reason,
		Details: details,
	}
}
