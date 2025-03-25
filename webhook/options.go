package webhook

import (
	"net/http"
	"time"
)

// RequestOption configures a single webhook request
type RequestOption func(*requestOptions)

// requestOptions holds all configurable request settings
type requestOptions struct {
	Method  string
	Headers map[string]string
	Timeout time.Duration
}

// WithMethod sets the HTTP method for a request
func WithMethod(method string) RequestOption {
	return func(o *requestOptions) {
		o.Method = method
	}
}

// WithHeader adds a header to the request
func WithHeader(key, value string) RequestOption {
	return func(o *requestOptions) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers[key] = value
	}
}

// WithHeaders sets multiple headers for the request
func WithHeaders(headers map[string]string) RequestOption {
	return func(o *requestOptions) {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		for k, v := range headers {
			o.Headers[k] = v
		}
	}
}

// WithRequestTimeout sets a timeout for this specific request
func WithRequestTimeout(timeout time.Duration) RequestOption {
	return func(o *requestOptions) {
		o.Timeout = timeout
	}
}

// SenderOption configures the webhook sender
type SenderOption func(*webhookSender)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) SenderOption {
	return func(s *webhookSender) {
		s.client = client
	}
}

// WithDefaultTimeout sets the default timeout for all requests
func WithDefaultTimeout(timeout time.Duration) SenderOption {
	return func(s *webhookSender) {
		s.defaultTimeout = timeout
	}
}

// WithDefaultHeaders sets default headers for all requests
func WithDefaultHeaders(headers map[string]string) SenderOption {
	return func(s *webhookSender) {
		if s.defaultHeaders == nil {
			s.defaultHeaders = make(map[string]string)
		}
		for k, v := range headers {
			s.defaultHeaders[k] = v
		}
	}
}

// WithDefaultMethod sets the default HTTP method for all requests
func WithDefaultMethod(method string) SenderOption {
	return func(s *webhookSender) {
		s.defaultMethod = method
	}
}

// WithMaxRetries sets the maximum number of retries for failed requests
func WithMaxRetries(retries int) SenderOption {
	return func(s *webhookSender) {
		s.maxRetries = retries
	}
}

// WithRetryInterval sets the interval between retries
func WithRetryInterval(interval time.Duration) SenderOption {
	return func(s *webhookSender) {
		s.retryInterval = interval
	}
}
