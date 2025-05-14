package pagination_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/gokit/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	t.Run("Default Configuration", func(t *testing.T) {
		// Create middleware with default config
		middleware := pagination.Middleware(nil)
		require.NotNil(t, middleware)

		// Create test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get paginator from context
			paginator := pagination.GetPaginatorFromContext(r.Context())

			// Verify paginator properties
			assert.Equal(t, 1, paginator.CurrentPage())
			assert.Equal(t, pagination.DefaultLimit, paginator.ItemsPerPage())
			assert.Equal(t, 0, paginator.TotalItems()) // Should start at 0

			// Update total items and write response
			updatedPaginator := paginator.WithTotalItems(100)
			_ = updatedPaginator.RespondWithJSON(w, []string{"test1", "test2"})
		})

		// Create test request
		req := httptest.NewRequest("GET", "http://example.com/api", nil)
		w := httptest.NewRecorder()

		// Apply middleware and execute handler
		middleware(testHandler).ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"items":["test1","test2"]`)
		assert.Contains(t, w.Body.String(), `"total_items":100`)
	})

	t.Run("Custom Configuration", func(t *testing.T) {
		// Create custom config
		config := &pagination.PaginatorConfig{
			PageParam:    "p",
			SizeParam:    "limit",
			DefaultLimit: 15,
			MaxLimit:     50,
		}

		// Create middleware with custom config
		middleware := pagination.Middleware(config)
		require.NotNil(t, middleware)

		// Create test handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get paginator from context
			paginator := pagination.GetPaginatorFromContext(r.Context())

			// Verify paginator properties with custom parameters
			assert.Equal(t, 2, paginator.CurrentPage())
			assert.Equal(t, 25, paginator.ItemsPerPage())

			// Return simple response
			w.WriteHeader(http.StatusOK)
		})

		// Create test request with custom parameters
		req := httptest.NewRequest("GET", "http://example.com/api?p=2&limit=25", nil)
		w := httptest.NewRecorder()

		// Apply middleware and execute handler
		middleware(testHandler).ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetPaginatorFromContext(t *testing.T) {
	t.Run("Paginator Exists", func(t *testing.T) {
		// Setup middleware
		middleware := pagination.Middleware(nil)
		require.NotNil(t, middleware)

		// Create handler that extracts paginator
		var extractedPaginator pagination.Paginator

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			extractedPaginator = pagination.GetPaginatorFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		// Create test request
		req := httptest.NewRequest("GET", "http://example.com/api?page=3&size=10", nil)
		w := httptest.NewRecorder()

		// Apply middleware and execute handler
		middleware(testHandler).ServeHTTP(w, req)

		// Verify extracted paginator
		assert.Equal(t, 3, extractedPaginator.CurrentPage())
		assert.Equal(t, 10, extractedPaginator.ItemsPerPage())
	})

	t.Run("No Paginator", func(t *testing.T) {
		// Create empty context
		ctx := context.Background()

		// Try to extract paginator from empty context
		paginator := pagination.GetPaginatorFromContext(ctx)

		// Should return zero value of Paginator
		assert.Equal(t, 0, paginator.CurrentPage())
		assert.Equal(t, 0, paginator.ItemsPerPage())
		assert.Equal(t, 0, paginator.TotalItems())
	})
}
