package webhook_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/gokit/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerDecorator(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Create a custom logger that writes to our buffer
	handler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Create a webhook sender
	baseSender := webhook.NewWebhookSender()
	
	// Create the logger decorator
	loggerSender := webhook.NewLoggerDecorator(baseSender, logger)

	// Send a request
	params := map[string]string{"key": "value"}
	ctx := context.Background()
	resp, err := loggerSender.Send(ctx, server.URL, params)

	// Verify the response
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, resp.IsSuccessful())

	// Verify that logs were created
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Sending webhook request")
	assert.Contains(t, logOutput, "Webhook request completed")
	assert.Contains(t, logOutput, server.URL)
	assert.Contains(t, logOutput, "status_code=200")
	assert.Contains(t, logOutput, "params=map[key:value]")
}

func TestLoggerDecoratorWithHideParams(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Create a custom logger that writes to our buffer
	handler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Create a webhook sender
	baseSender := webhook.NewWebhookSender()
	
	// Create the logger decorator with hideParams option
	loggerSender := webhook.NewLoggerDecorator(baseSender, logger, webhook.WithHideParams())

	// Send a request with sensitive data that should be hidden
	params := map[string]string{
		"api_key": "secret_api_key_123",
		"password": "very_secret_password",
	}
	ctx := context.Background()
	resp, err := loggerSender.Send(ctx, server.URL, params)

	// Verify the response
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, resp.IsSuccessful())

	// Verify that logs were created but sensitive params are hidden
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Sending webhook request")
	assert.Contains(t, logOutput, "Webhook request completed")
	assert.Contains(t, logOutput, server.URL)
	assert.Contains(t, logOutput, "status_code=200")
	
	// Ensure sensitive params are not in the logs
	assert.NotContains(t, logOutput, "secret_api_key_123")
	assert.NotContains(t, logOutput, "very_secret_password")
	assert.NotContains(t, logOutput, "params=")
}

func TestLoggerDecoratorWithMaskedFields(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Create a custom logger that writes to our buffer
	handler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Create a webhook sender
	baseSender := webhook.NewWebhookSender()
	
	// Create the logger decorator with masked fields
	loggerSender := webhook.NewLoggerDecorator(
		baseSender, 
		logger, 
		webhook.WithMaskedFields("api_key", "password"),
	)

	// Send a request with some fields that should be masked
	params := map[string]string{
		"api_key": "secret_api_key_123",
		"password": "very_secret_password",
		"username": "john_doe",
		"event": "user.created",
	}
	ctx := context.Background()
	resp, err := loggerSender.Send(ctx, server.URL, params)

	// Verify the response
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, resp.IsSuccessful())

	// Verify that logs were created with only sensitive fields masked
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Sending webhook request")
	assert.Contains(t, logOutput, "Webhook request completed")
	assert.Contains(t, logOutput, server.URL)
	assert.Contains(t, logOutput, "status_code=200")
	
	// Ensure sensitive data is masked but non-sensitive is not
	assert.NotContains(t, logOutput, "secret_api_key_123")
	assert.NotContains(t, logOutput, "very_secret_password")
	assert.Contains(t, logOutput, "username")
	assert.Contains(t, logOutput, "john_doe")
	assert.Contains(t, logOutput, "event")
	assert.Contains(t, logOutput, "user.created")
}

func TestLoggerDecoratorWithMaskedFieldsStruct(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Create a custom logger that writes to our buffer
	handler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Create a webhook sender
	baseSender := webhook.NewWebhookSender()
	
	// Create the logger decorator with masked fields
	loggerSender := webhook.NewLoggerDecorator(
		baseSender, 
		logger, 
		webhook.WithMaskedFields("api_key", "password"),
	)

	// Define a struct with JSON tags
	type UserParams struct {
		APIKey   string `json:"api_key"`
		Password string `json:"password"`
		Username string `json:"username"`
		UserID   int    `json:"user_id"`
	}

	// Send a request with struct that has some fields that should be masked
	params := UserParams{
		APIKey:   "secret_api_key_123",
		Password: "very_secret_password",
		Username: "john_doe",
		UserID:   12345,
	}
	ctx := context.Background()
	resp, err := loggerSender.Send(ctx, server.URL, params)

	// Verify the response
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, resp.IsSuccessful())

	// Verify that logs were created with only sensitive fields masked
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Sending webhook request")
	assert.Contains(t, logOutput, "Webhook request completed")
	assert.Contains(t, logOutput, server.URL)
	assert.Contains(t, logOutput, "status_code=200")
	
	// Ensure sensitive data is masked but non-sensitive is not
	assert.NotContains(t, logOutput, "secret_api_key_123")
	assert.NotContains(t, logOutput, "very_secret_password")
	assert.Contains(t, logOutput, "Username")
	assert.Contains(t, logOutput, "john_doe")
	assert.Contains(t, logOutput, "UserID")
	assert.Contains(t, logOutput, "12345")
}

func TestLoggerDecoratorWithError(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer
	
	// Create a custom logger that writes to our buffer
	handler := slog.NewTextHandler(&logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)

	// Create a webhook sender
	baseSender := webhook.NewWebhookSender()
	
	// Create the logger decorator
	loggerSender := webhook.NewLoggerDecorator(baseSender, logger)

	// Send a request to an invalid URL
	params := map[string]string{"key": "value"}
	ctx := context.Background()
	_, err := loggerSender.Send(ctx, "invalid-url", params)

	// Verify the error
	require.Error(t, err)

	// Verify that error logs were created
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "Sending webhook request")
	assert.Contains(t, logOutput, "Webhook request failed")
	assert.Contains(t, logOutput, "invalid-url")
}

func TestLoggerDecoratorWithNilLogger(t *testing.T) {
	// Create a webhook sender
	baseSender := webhook.NewWebhookSender()
	
	// Create the logger decorator with nil logger (should use default)
	loggerSender := webhook.NewLoggerDecorator(baseSender, nil)

	// This just checks that we don't panic with a nil logger
	assert.NotNil(t, loggerSender)
}
