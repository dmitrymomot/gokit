package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookSender is the interface for sending webhooks
type WebhookSender interface {
	// Send webhooks with minimal required parameters and optional request options
	Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error)
}

// webhookSender implements the WebhookSender interface
type webhookSender struct {
	client         *http.Client
	defaultMethod  string
	defaultHeaders map[string]string
	defaultTimeout time.Duration
	maxRetries     int
	retryInterval  time.Duration
}

// NewWebhookSender creates a new webhook sender
func NewWebhookSender(opts ...SenderOption) WebhookSender {
	s := &webhookSender{
		client:        http.DefaultClient,
		defaultMethod: http.MethodPost,
		defaultHeaders: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		defaultTimeout: 30 * time.Second,
		maxRetries:     0,
		retryInterval:  time.Second,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Send implements the WebhookSender interface
func (s *webhookSender) Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error) {
	if url == "" {
		return nil, ErrInvalidURL
	}

	// Set up default request options
	options := &requestOptions{
		Method:  s.defaultMethod,
		Headers: make(map[string]string),
		Timeout: s.defaultTimeout,
	}

	// Apply global default headers
	for k, v := range s.defaultHeaders {
		options.Headers[k] = v
	}

	// Apply request-specific options
	for _, opt := range opts {
		opt(options)
	}

	// Create request
	req := &Request{
		URL:     url,
		Method:  options.Method,
		Headers: options.Headers,
		Params:  params,
		Timeout: options.Timeout,
	}

	// Execute request with retry logic
	var resp *Response
	var err error

	attempts := 0
	maxAttempts := s.maxRetries + 1 // +1 for the initial attempt

	for attempts < maxAttempts {
		resp, err = s.doSend(ctx, req)
		attempts++

		// If no error and successful response, or we've reached max attempts, break the loop
		if (err == nil && resp.IsSuccessful()) || attempts >= maxAttempts {
			break
		}

		// If there was an error or unsuccessful response and we have retries left, wait and retry
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrResponseTimeout, ctx.Err())
		case <-time.After(s.retryInterval):
			// continue with retry
		}
	}

	return resp, err
}

// doSend performs the actual HTTP request
func (s *webhookSender) doSend(ctx context.Context, req *Request) (*Response, error) {
	httpReq, err := createHTTPRequest(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreateRequest, err)
	}

	// Add the request to the context
	httpReq = httpReq.WithContext(ctx)

	// Create a context with timeout if specified
	var cancel context.CancelFunc
	if req.Timeout > 0 {
		var timeoutCtx context.Context
		timeoutCtx, cancel = context.WithTimeout(ctx, req.Timeout)
		httpReq = httpReq.WithContext(timeoutCtx)
	}

	// Ensure cancel is called
	if cancel != nil {
		defer cancel()
	}

	// Execute the request
	startTime := time.Now()
	httpResp, err := s.client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendRequest, err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReadResponse, err)
	}

	// Create response
	response := &Response{
		StatusCode: httpResp.StatusCode,
		Body:       body,
		Headers:    httpResp.Header,
		Duration:   duration,
		Request:    req,
	}

	return response, nil
}

// marshalParams marshals the parameters to JSON and returns a buffer
func marshalParams(params any) (*bytes.Buffer, error) {
	if params == nil {
		return bytes.NewBuffer(nil), nil
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMarshalParams, err)
	}

	return bytes.NewBuffer(data), nil
}
