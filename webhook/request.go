package webhook

import (
	"net/http"
	"time"
)

// Request represents a webhook request
type Request struct {
	URL     string
	Method  string
	Headers map[string]string
	Params  interface{}
	Timeout time.Duration
}

// Response represents a webhook response
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	Duration   time.Duration
	Request    *Request // Reference to the original request
}

// IsSuccessful returns true if the response status code is in the 2xx range
func (r *Response) IsSuccessful() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// createHTTPRequest creates a new HTTP request with the given parameters
func createHTTPRequest(req *Request) (*http.Request, error) {
	if req.URL == "" {
		return nil, ErrInvalidURL
	}

	var httpReq *http.Request
	var err error

	// Create request with or without body depending on the HTTP method
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		// These methods typically don't have a body
		httpReq, err = http.NewRequest(req.Method, req.URL, nil)
	default:
		// For other methods like POST, PUT, PATCH, include the body
		data, err := marshalParams(req.Params)
		if err != nil {
			return nil, err
		}
		httpReq, err = http.NewRequest(req.Method, req.URL, data)
	}

	if err != nil {
		return nil, err
	}

	// Add headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}
