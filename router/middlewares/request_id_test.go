package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/gokit/router/middlewares"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRequestID(t *testing.T) {
	t.Run("context without request ID returns empty string", func(t *testing.T) {
		reqID := middlewares.GetRequestID(context.Background())
		assert.Equal(t, "", reqID)
	})

	t.Run("context with request ID returns the ID", func(t *testing.T) {
		expectedID := "test-request-id"
		ctx := middlewares.WithRequestID(context.Background(), expectedID)
		reqID := middlewares.GetRequestID(ctx)
		assert.Equal(t, expectedID, reqID)
	})
}

func TestWithRequestID(t *testing.T) {
	t.Run("adds request ID to context", func(t *testing.T) {
		originalCtx := context.Background()
		expectedID := "test-request-id"

		newCtx := middlewares.WithRequestID(originalCtx, expectedID)
		reqID := middlewares.GetRequestID(newCtx)

		assert.Equal(t, expectedID, reqID)
		assert.NotEqual(t, originalCtx, newCtx)
	})
}

func TestRequestID(t *testing.T) {
	t.Run("generates new request ID when header is empty", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middlewares.GetRequestID(r.Context())
			require.NotEmpty(t, reqID)

			// Validate that the generated ID is a valid UUID
			_, err := uuid.Parse(reqID)
			require.NoError(t, err)

			// Check response header
			assert.Equal(t, reqID, w.Header().Get(middlewares.RequestIDHeader))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handler := middlewares.RequestID(nextHandler)
		handler.ServeHTTP(rec, req)
	})

	t.Run("uses existing request ID from header", func(t *testing.T) {
		expectedID := "existing-request-id"

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middlewares.GetRequestID(r.Context())
			assert.Equal(t, expectedID, reqID)
			assert.Equal(t, expectedID, w.Header().Get(middlewares.RequestIDHeader))
		})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(middlewares.RequestIDHeader, expectedID)
		rec := httptest.NewRecorder()

		handler := middlewares.RequestID(nextHandler)
		handler.ServeHTTP(rec, req)
	})
}
