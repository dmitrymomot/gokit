package middlewares_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/gokit/router/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecovererMiddleware(t *testing.T) {
	t.Run("should recover from panic", func(t *testing.T) {
		// Handler that panics
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		// Wrap with recoverer
		handler := middlewares.Recoverer(panicHandler)

		// Create test request and response recorder
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		// Execute handler
		handler.ServeHTTP(res, req)

		// Verify response is 500 Internal Server Error
		require.Equal(t, http.StatusInternalServerError, res.Code)
		require.Equal(t, "Internal Server Error\n", res.Body.String())
	})

	t.Run("should not interfere with non-panicking handlers", func(t *testing.T) {
		// Normal handler
		normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		// Wrap with recoverer
		handler := middlewares.Recoverer(normalHandler)

		// Create test request and response recorder
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		// Execute handler
		handler.ServeHTTP(res, req)

		// Verify response is normal
		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "success", res.Body.String())
	})

	t.Run("should not recover http.ErrAbortHandler", func(t *testing.T) {
		// Handler that panics with http.ErrAbortHandler
		abortHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic(http.ErrAbortHandler)
		})

		// Wrap with recoverer
		handler := middlewares.Recoverer(abortHandler)

		// Create test request and response recorder
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		// Execute handler - should panic
		assert.Panics(t, func() {
			handler.ServeHTTP(res, req)
		})
	})
}

func TestRecovererWithHandlerMiddleware(t *testing.T) {
	t.Run("should use custom error handler", func(t *testing.T) {
		// Custom error handler
		customHandler := func(w http.ResponseWriter, r *http.Request, err any) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("custom error: " + err.(string)))
		}

		// Handler that panics
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("custom panic")
		})

		// Wrap with custom recoverer
		handler := middlewares.RecovererWithHandler(panicHandler, customHandler)

		// Create test request and response recorder
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		// Execute handler
		handler.ServeHTTP(res, req)

		// Verify custom response
		require.Equal(t, http.StatusServiceUnavailable, res.Code)
		require.Equal(t, "custom error: custom panic", res.Body.String())
	})

	t.Run("should use default handler when nil is provided", func(t *testing.T) {
		// Handler that panics
		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		})

		// Wrap with nil custom handler (should use default)
		handler := middlewares.RecovererWithHandler(panicHandler, nil)

		// Create test request and response recorder
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		// Execute handler
		handler.ServeHTTP(res, req)

		// Verify default response
		require.Equal(t, http.StatusInternalServerError, res.Code)
		require.Equal(t, "Internal Server Error\n", res.Body.String())
	})
}

func TestDefaultRecovererErrorHandler(t *testing.T) {
	t.Run("should return 500 Internal Server Error", func(t *testing.T) {
		// Create test request and response recorder
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()

		// Call the default handler directly
		middlewares.DefaultRecovererErrorHandler(res, req, "test error")

		// Verify response
		require.Equal(t, http.StatusInternalServerError, res.Code)
		require.Equal(t, "Internal Server Error\n", res.Body.String())
	})

	t.Run("should handle different error types", func(t *testing.T) {
		testCases := []struct {
			name  string
			value any
		}{
			{"string error", "string error message"},
			{"error interface", errors.New("error interface message")},
			{"integer error", 42},
			{"nil error", nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Create test request and response recorder
				req := httptest.NewRequest("GET", "/", nil)
				res := httptest.NewRecorder()

				// Call the default handler with the test error
				middlewares.DefaultRecovererErrorHandler(res, req, tc.value)

				// Verify response is always 500
				require.Equal(t, http.StatusInternalServerError, res.Code)
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				require.Equal(t, "Internal Server Error\n", string(body))
			})
		}
	})
}
