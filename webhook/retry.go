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
	
	// Check for context cancellation (don't retry)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	
	// Check for network-related errors
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return true
	}
	
	// Check for temporary errors
	return isTemporaryError(err)
}

// NewRetryDecorator creates a new webhook sender with retry capabilities
func NewRetryDecorator(sender WebhookSender, opts ...RetryOption) WebhookSender {
	rd := &RetryDecorator{
		sender:        sender,
		maxRetries:    DefaultMaxRetries,
		retryInterval: DefaultRetryInterval,
		retryOn: func(resp *Response, err error) bool {
			// By default, retry on any error except context cancellation
			return err != nil && 
				!errors.Is(err, context.Canceled) && 
				!errors.Is(err, context.DeadlineExceeded)
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
		resp        *Response
		err         error
		lastErr     error
		attempt     int
		shouldRetry bool
		nextInterval = r.retryInterval
		didRetry    bool
	)
	
	// Make initial attempt (attempt 0)
	resp, err = r.sender.Send(ctx, url, params, opts...)
	if err != nil {
		lastErr = err
	}
	
	// Check if context is already canceled - if so, return immediately
	if errors.Is(ctx.Err(), context.Canceled) {
		return resp, errors.Join(lastErr, ctx.Err())
	}
	
	// Retry logic (attempts 1 to maxRetries)
	for attempt = 1; attempt <= r.maxRetries; attempt++ {
		// Check if we should retry
		shouldRetry = r.retryOn(resp, err)
		if !shouldRetry {
			break
		}
		
		// Mark that we did at least one retry
		didRetry = true
		
		// Log retry attempt if logger is provided
		if r.logger != nil {
			r.logger.InfoContext(ctx, "Retrying webhook request",
				slog.String("url", url),
				slog.Int("attempt", attempt),
				slog.Int("max_retries", r.maxRetries),
				slog.Duration("next_interval", nextInterval),
			)
		}
		
		// Wait before next retry
		select {
		case <-ctx.Done():
			// Context canceled during wait
			return resp, errors.Join(lastErr, ctx.Err())
		case <-time.After(nextInterval):
			// Wait completed, proceed with retry
		}
		
		// Make retry attempt
		resp, err = r.sender.Send(ctx, url, params, opts...)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode < 400 {
			// If we got a successful response, break the retry loop
			lastErr = nil
			break
		}
		
		// If using backoff, double the interval for next retry
		if r.useBackoff {
			nextInterval *= 2
		}
	}
	
	// Log final retry result if logger is provided
	if r.logger != nil && didRetry {
		if lastErr != nil {
			r.logger.ErrorContext(ctx, "Webhook retry failed",
				slog.String("url", url),
				slog.Int("attempts", attempt),
				slog.Any("error", lastErr),
			)
		} else {
			r.logger.InfoContext(ctx, "Webhook retry succeeded",
				slog.String("url", url),
				slog.Int("attempts", attempt),
				slog.Int("status_code", resp.StatusCode),
			)
		}
	}
	
	return resp, lastErr
}
