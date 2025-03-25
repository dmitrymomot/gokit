package webhook_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSender implements webhook.WebhookSender for testing
type mockSender struct {
	// responses defines what to return on each call
	responses []*webhook.Response
	// errors defines what errors to return on each call
	errors []error
	// callCount tracks how many times Send was called
	callCount atomic.Int32
	// lastParams tracks the last params sent
	lastParams any
	// lastURL tracks the last URL used
	lastURL string
	// lastOpts tracks the last options used
	lastOpts []webhook.RequestOption
}

func (m *mockSender) Send(ctx context.Context, url string, params any, opts ...webhook.RequestOption) (*webhook.Response, error) {
	count := m.callCount.Add(1) - 1 // get current call index (0-based)
	
	// Save parameters for inspection
	m.lastParams = params
	m.lastURL = url
	m.lastOpts = opts
	
	// Return pre-configured response/error for this call
	if int(count) < len(m.errors) && m.errors[count] != nil {
		return nil, m.errors[count]
	}
	
	if int(count) < len(m.responses) {
		return m.responses[count], nil
	}
	
	// Default successful response
	return &webhook.Response{StatusCode: http.StatusOK}, nil
}

func TestRetryDecorator(t *testing.T) {
	// Create a context for all tests
	ctx := context.Background()

	t.Run("successful_first_attempt", func(t *testing.T) {
		// Create mock sender that succeeds on first try
		mock := &mockSender{
			responses: []*webhook.Response{
				{StatusCode: http.StatusOK, Body: []byte(`{"success":true}`)},
			},
		}
		
		// Create retry decorator
		sender := webhook.NewRetryDecorator(mock)
		
		// Send request
		resp, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		
		// Verify results
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(1), mock.callCount.Load())
	})
	
	t.Run("retry_on_network_error", func(t *testing.T) {
		// Create mock sender that fails with a network error, then succeeds
		mock := &mockSender{
			errors: []error{
				errors.New("connection refused"),
				nil,
			},
			responses: []*webhook.Response{
				nil,
				{StatusCode: http.StatusOK, Body: []byte(`{"success":true}`)},
			},
		}
		
		// Create retry decorator
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(3),
			webhook.WithRetryDelay(10*time.Millisecond),
			webhook.WithRetryOnNetworkErrors(),
		)
		
		// Send request
		resp, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		
		// Verify results
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(2), mock.callCount.Load())
	})
	
	t.Run("retry_on_status_code", func(t *testing.T) {
		// Create mock sender that returns 503, then succeeds
		mock := &mockSender{
			responses: []*webhook.Response{
				{StatusCode: http.StatusServiceUnavailable, Body: []byte(`{"error":"service unavailable"}`)},
				{StatusCode: http.StatusOK, Body: []byte(`{"success":true}`)},
			},
		}
		
		// Create retry decorator that retries on 503
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(3),
			webhook.WithRetryDelay(10*time.Millisecond),
			webhook.WithRetryOnStatus(http.StatusServiceUnavailable),
		)
		
		// Send request
		resp, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		
		// Verify results
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(2), mock.callCount.Load())
	})
	
	t.Run("retry_on_server_errors", func(t *testing.T) {
		// Create mock sender that returns multiple 5xx errors, then succeeds
		mock := &mockSender{
			responses: []*webhook.Response{
				{StatusCode: http.StatusInternalServerError, Body: []byte(`{"error":"server error"}`)},
				{StatusCode: http.StatusBadGateway, Body: []byte(`{"error":"bad gateway"}`)},
				{StatusCode: http.StatusOK, Body: []byte(`{"success":true}`)},
			},
		}
		
		// Create retry decorator that retries on all 5xx errors
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(3),
			webhook.WithRetryDelay(10*time.Millisecond),
			webhook.WithRetryOnServerErrors(),
		)
		
		// Send request
		resp, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		
		// Verify results
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(3), mock.callCount.Load())
	})
	
	t.Run("max_retries_exceeded", func(t *testing.T) {
		// Create mock sender that always fails
		errExpected := errors.New("persistent error")
		mock := &mockSender{
			errors: []error{
				errExpected,
				errExpected,
				errExpected,
				errExpected,
			},
		}
		
		// Create retry decorator with max 3 retries
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(3),
			webhook.WithRetryDelay(10*time.Millisecond),
			webhook.WithRetryOnNetworkErrors(),
		)
		
		// Send request
		_, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		
		// Verify results
		require.Error(t, err)
		assert.Equal(t, errExpected, err)
		assert.Equal(t, int32(4), mock.callCount.Load()) // original + 3 retries
	})
	
	t.Run("retry_with_backoff", func(t *testing.T) {
		// Create mock sender that fails a few times
		mock := &mockSender{
			errors: []error{
				errors.New("error 1"),
				errors.New("error 2"),
				nil,
			},
			responses: []*webhook.Response{
				nil,
				nil,
				{StatusCode: http.StatusOK, Body: []byte(`{"success":true}`)},
			},
		}
		
		// Create retry decorator with backoff
		startTime := time.Now()
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(3),
			webhook.WithRetryDelay(50*time.Millisecond),
			webhook.WithRetryBackoff(),
			webhook.WithRetryOnNetworkErrors(),
		)
		
		// Send request
		resp, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		elapsed := time.Since(startTime)
		
		// Verify results
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(3), mock.callCount.Load())
		
		// With backoff, we should have waited at least:
		// 1st retry: 50ms
		// 2nd retry: 100ms (doubled from first)
		// Total: ~150ms minimum
		assert.GreaterOrEqual(t, elapsed.Milliseconds(), int64(140))
	})
	
	t.Run("with_logger", func(t *testing.T) {
		// Setup a buffer to capture log output
		var logBuffer bytes.Buffer
		handler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(handler)
		
		// Create mock sender that fails once then succeeds
		mock := &mockSender{
			errors: []error{
				errors.New("temporary error"),
				nil,
			},
			responses: []*webhook.Response{
				nil,
				{StatusCode: http.StatusOK, Body: []byte(`{"success":true}`)},
			},
		}
		
		// Create retry decorator with logger
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(3),
			webhook.WithRetryDelay(10*time.Millisecond),
			webhook.WithRetryLogger(logger),
			webhook.WithRetryOnNetworkErrors(),
		)
		
		// Send request
		resp, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
		
		// Verify results
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, int32(2), mock.callCount.Load())
		
		// Verify logs
		logs := logBuffer.String()
		assert.Contains(t, logs, "Retrying webhook request")
		assert.Contains(t, logs, "Webhook retry succeeded")
		assert.Contains(t, logs, "attempt=1")
		assert.Contains(t, logs, "max_retries=3")
	})
	
	t.Run("context_cancellation", func(t *testing.T) {
		// Create a context that will be canceled
		ctx, cancel := context.WithCancel(context.Background())
		
		// Create mock sender with long delay and network error
		mock := &mockSender{
			errors: []error{
				errors.New("will retry"),
			},
		}
		
		// Create retry decorator with a delay
		sender := webhook.NewRetryDecorator(
			mock,
			webhook.WithRetryCount(5),
			webhook.WithRetryDelay(200*time.Millisecond),
			webhook.WithRetryOnNetworkErrors(),
		)
		
		// Start send operation in a goroutine
		errCh := make(chan error, 1)
		go func() {
			_, err := sender.Send(ctx, "https://api.example.com", map[string]string{"test": "value"})
			errCh <- err
		}()
		
		// Cancel the context after a short time
		time.Sleep(20 * time.Millisecond)
		cancel()
		
		// Get the result
		var err error
		select {
		case err = <-errCh:
			// Got response
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Timed out waiting for response")
		}
		
		// Verify context cancellation was properly handled
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
