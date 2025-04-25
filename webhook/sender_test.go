package webhook_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookSender(t *testing.T) {
	t.Run("with default options", func(t *testing.T) {
		sender := webhook.NewWebhookSender()
		require.NotNil(t, sender, "webhook sender should not be nil")
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		client := &http.Client{Timeout: 15 * time.Second}
		sender := webhook.NewWebhookSender(webhook.WithHTTPClient(client))
		require.NotNil(t, sender, "webhook sender should not be nil")
	})

	t.Run("with multiple options", func(t *testing.T) {
		sender := webhook.NewWebhookSender(
			webhook.WithDefaultMethod(http.MethodPut),
			webhook.WithDefaultTimeout(15*time.Second),
			webhook.WithDefaultHeaders(map[string]string{
				"X-Custom-Header": "test",
			}),
		)
		require.NotNil(t, sender, "webhook sender should not be nil")
	})
}

func TestWebhookSender_Send(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check method
			assert.Equal(t, http.MethodPost, r.Method)
			// Check headers
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-value", r.Header.Get("X-Test-Header"))

			// Write response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}))
		defer server.Close()

		// Create webhook sender
		sender := webhook.NewWebhookSender()

		// Send webhook
		ctx := context.Background()
		params := map[string]string{"key": "value"}
		resp, err := sender.Send(ctx, server.URL, params,
			webhook.WithHeader("X-Test-Header", "test-value"))

		// Check response
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "success")
		assert.True(t, resp.IsSuccessful())
	})

	t.Run("GET request with query params from map", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check method
			assert.Equal(t, http.MethodGet, r.Method)

			// Check that params were converted to query string
			assert.Equal(t, "value1", r.URL.Query().Get("key1"))
			assert.Equal(t, "value2", r.URL.Query().Get("key2"))

			// Write response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data":"received"}`))
		}))
		defer server.Close()

		// Create webhook sender
		sender := webhook.NewWebhookSender()

		// Send webhook with params that should be converted to query string
		ctx := context.Background()
		params := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		resp, err := sender.Send(ctx, server.URL, params,
			webhook.WithMethod(http.MethodGet))

		// Check response
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "data")
	})

	t.Run("GET request with struct params", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check method
			assert.Equal(t, http.MethodGet, r.Method)

			// Check that struct fields were converted to query string
			assert.Equal(t, "123", r.URL.Query().Get("id"))
			assert.Equal(t, "test user", r.URL.Query().Get("name"))

			// Write response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}))
		defer server.Close()

		// Create webhook sender
		sender := webhook.NewWebhookSender()

		// Define a struct with json tags
		type User struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		// Send webhook with struct params
		ctx := context.Background()
		params := User{ID: 123, Name: "test user"}
		resp, err := sender.Send(ctx, server.URL, params,
			webhook.WithMethod(http.MethodGet))

		// Check response
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "success")
	})

	t.Run("DELETE request with query params", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check method
			assert.Equal(t, http.MethodDelete, r.Method)

			// Check that params were converted to query string
			assert.Equal(t, "123", r.URL.Query().Get("id"))

			// Write response
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"deleted":true}`))
		}))
		defer server.Close()

		// Create webhook sender
		sender := webhook.NewWebhookSender()

		// Send webhook with params that should be converted to query string
		ctx := context.Background()
		params := map[string]int{"id": 123}
		resp, err := sender.Send(ctx, server.URL, params,
			webhook.WithMethod(http.MethodDelete))

		// Check response
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Contains(t, string(resp.Body), "deleted")
	})

	t.Run("existing query params should be preserved", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that both existing and new params are present
			assert.Equal(t, "existing", r.URL.Query().Get("original"))
			assert.Equal(t, "new value", r.URL.Query().Get("added"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create webhook sender
		sender := webhook.NewWebhookSender()

		// Send webhook to URL with existing query params
		ctx := context.Background()
		params := map[string]string{"added": "new value"}
		resp, err := sender.Send(ctx, server.URL+"?original=existing", params,
			webhook.WithMethod(http.MethodGet))

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("error - invalid URL", func(t *testing.T) {
		sender := webhook.NewWebhookSender()

		resp, err := sender.Send(context.Background(), "", nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorIs(t, err, webhook.ErrInvalidURL)
	})

	t.Run("error - request timeout", func(t *testing.T) {
		// Create a slow server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		sender := webhook.NewWebhookSender()

		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		resp, err := sender.Send(ctx, server.URL, nil)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("custom headers override default headers", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check that custom headers override defaults
			assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create webhook sender with default headers
		sender := webhook.NewWebhookSender(
			webhook.WithDefaultHeaders(map[string]string{
				"Content-Type": "application/json",
			}),
		)

		// Send with custom headers that override defaults
		resp, err := sender.Send(
			context.Background(),
			server.URL,
			"plain text",
			webhook.WithHeader("Content-Type", "text/plain"),
		)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("retry on failure", func(t *testing.T) {
		attempts := 0

		// Create a server that fails on first attempt
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true}`))
		}))
		defer server.Close()

		// Create sender with retry
		sender := webhook.NewWebhookSender(
			webhook.WithMaxRetries(1),
			webhook.WithRetryInterval(10*time.Millisecond),
		)

		resp, err := sender.Send(context.Background(), server.URL, nil)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, attempts)
	})
}

func TestResponse_IsSuccessful(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"status 200", 200, true},
		{"status 201", 201, true},
		{"status 299", 299, true},
		{"status 300", 300, false},
		{"status 400", 400, false},
		{"status 500", 500, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &webhook.Response{
				StatusCode: tc.statusCode,
			}
			assert.Equal(t, tc.expected, resp.IsSuccessful())
		})
	}
}
