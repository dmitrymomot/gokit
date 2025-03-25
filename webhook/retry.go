package webhook

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"
)

// NetworkError represents a network-related error when sending a webhook
type NetworkError struct {
	Err     error
	Message string
}

func (e *NetworkError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// DefaultMaxRetries is the default number of retry attempts
const DefaultMaxRetries = 3

// DefaultRetryInterval is the default interval between retry attempts
const DefaultRetryInterval = 500 * time.Millisecond

// RetryDecorator wraps a WebhookSender and adds retry functionality
type RetryDecorator struct {
	sender        WebhookSender
	maxRetries    int
	retryInterval time.Duration
	useBackoff    bool
	retryOn       func(resp *Response, err error) bool
	logger        *slog.Logger
}

// RetryOption configures the retry decorator
type RetryOption func(*RetryDecorator)

// WithRetryCount sets the maximum number of retry attempts
func WithRetryCount(max int) RetryOption {
	return func(r *RetryDecorator) {
		if max > 0 {
			r.maxRetries = max
		}
	}
}

// WithRetryDelay sets the interval between retry attempts
func WithRetryDelay(interval time.Duration) RetryOption {
	return func(r *RetryDecorator) {
		if interval > 0 {
			r.retryInterval = interval
		}
	}
}

// WithRetryBackoff enables exponential backoff for retry intervals
// The interval will double after each retry attempt
func WithRetryBackoff() RetryOption {
	return func(r *RetryDecorator) {
		r.useBackoff = true
	}
}

// WithRetryOnStatus adds specific HTTP status codes that should trigger a retry
func WithRetryOnStatus(statusCodes ...int) RetryOption {
	return func(r *RetryDecorator) {
		// Create a map for faster lookups
		statusMap := make(map[int]bool, len(statusCodes))
		for _, code := range statusCodes {
			statusMap[code] = true
		}

		// Keep existing retry condition and add status code check
		existingCheck := r.retryOn
		r.retryOn = func(resp *Response, err error) bool {
			// If we already would retry based on existing conditions
			if existingCheck != nil && existingCheck(resp, err) {
				return true
			}

			// Also retry if the response has one of the specified status codes
			if resp != nil && statusMap[resp.StatusCode] {
				return true
			}

			return false
		}
	}
}

// WithRetryOnServerErrors configures the decorator to retry on all 5xx server errors
func WithRetryOnServerErrors() RetryOption {
	return func(r *RetryDecorator) {
		// Keep existing retry condition and add server error check
		existingCheck := r.retryOn
		r.retryOn = func(resp *Response, err error) bool {
			// If we already would retry based on existing conditions
			if existingCheck != nil && existingCheck(resp, err) {
				return true
			}

			// Retry on any 5xx status code
			if resp != nil && resp.StatusCode >= 500 && resp.StatusCode < 600 {
				return true
			}

			return false
		}
	}
}

// WithRetryOnNetworkErrors configures the decorator to retry on network-related errors
func WithRetryOnNetworkErrors() RetryOption {
	return func(r *RetryDecorator) {
		// Keep existing retry condition and add network error check
		existingCheck := r.retryOn
		r.retryOn = func(resp *Response, err error) bool {
			// If we already would retry based on existing conditions
			if existingCheck != nil && existingCheck(resp, err) {
				return true
			}

			// Check for network-related errors
			var netErr *NetworkError
			if err != nil && (errors.Is(err, context.DeadlineExceeded) ||
				errors.As(err, &netErr) || isTemporaryError(err)) {
				return true
			}

			return false
		}
	}
}

// WithRetryLogger sets a logger for retry operations
func WithRetryLogger(logger *slog.Logger) RetryOption {
	return func(r *RetryDecorator) {
		r.logger = logger
	}
}

// isTemporaryError checks if an error is temporary and worth retrying
func isTemporaryError(err error) bool {
	// Check for common temporary error types
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Temporary()
	}

	// Check for temporary interface (some errors implement this)
	type temporary interface {
		Temporary() bool
	}

	if te, ok := err.(temporary); ok {
		return te.Temporary()
	}

	// Check error strings for common temporary error patterns
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") ||
		strings.Contains(errStr, "reset by peer") ||
		strings.Contains(errStr, "connection closed")
}

// isRetryableError checks if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Most errors are retryable by default, except context cancellation
	if errors.Is(err, context.Canceled) {
		return false
	}

	return true
}

// NewRetryDecorator creates a new webhook sender with retry capabilities
func NewRetryDecorator(sender WebhookSender, opts ...RetryOption) WebhookSender {
	rd := &RetryDecorator{
		sender:        sender,
		maxRetries:    DefaultMaxRetries,
		retryInterval: DefaultRetryInterval,
		retryOn: func(resp *Response, err error) bool {
			return isRetryableError(err)
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(rd)
	}

	return rd
}

// Send sends a webhook with retry logic based on configuration
func (r *RetryDecorator) Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error) {
	var (
		resp         *Response
		err          error
		retryCount   int
		nextInterval = r.retryInterval
	)

	// Initial request (not counted as a retry)
	resp, err = r.sender.Send(ctx, url, params, opts...)

	// Check if context is already canceled
	if ctx.Err() != nil {
		return resp, errors.Join(err, ctx.Err())
	}

	// Try again up to maxRetries times if needed
	for retryCount < r.maxRetries {
		// Check if we should retry
		if !r.retryOn(resp, err) {
			break
		}

		// Log retry attempt if logger is provided
		if r.logger != nil {
			r.logger.InfoContext(ctx, "Retrying webhook request",
				slog.String("url", url),
				slog.Int("attempt", retryCount+1),
				slog.Int("max_retries", r.maxRetries),
				slog.Duration("next_interval", nextInterval),
			)
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			// Context was canceled during wait
			return resp, errors.Join(err, ctx.Err())
		case <-time.After(nextInterval):
			// Continue with retry
		}

		// Perform the retry
		resp, err = r.sender.Send(ctx, url, params, opts...)
		retryCount++

		// If successful, stop retrying
		if err == nil && resp != nil && resp.IsSuccessful() {
			break
		}

		// Update backoff interval if enabled
		if r.useBackoff {
			nextInterval *= 2
		}
	}

	// Log final retry status if logger was provided and we did retry
	if r.logger != nil && retryCount > 0 {
		if err != nil {
			r.logger.ErrorContext(ctx, "Webhook retry failed",
				slog.String("url", url),
				slog.Int("attempts", retryCount+1),
				slog.Any("error", err),
			)
		} else {
			r.logger.InfoContext(ctx, "Webhook retry succeeded",
				slog.String("url", url),
				slog.Int("attempts", retryCount+1),
				slog.Int("status_code", resp.StatusCode),
			)
		}
	}

	return resp, err
}
