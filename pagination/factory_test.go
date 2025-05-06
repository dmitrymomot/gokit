package pagination_test

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/gokit/pagination"
	"github.com/stretchr/testify/assert"
)

func TestNewPaginator(t *testing.T) {
	t.Run("Standard Parameters", func(t *testing.T) {
		p := pagination.NewPaginator(2, 20, 100, nil)
		assert.Equal(t, 2, p.CurrentPage())
		assert.Equal(t, 20, p.ItemsPerPage())
		assert.Equal(t, 100, p.TotalItems())
		assert.Equal(t, 5, p.TotalPages())
	})

	t.Run("Invalid Parameters", func(t *testing.T) {
		// Invalid page (should use default)
		p1 := pagination.NewPaginator(0, 20, 100, nil)
		assert.Equal(t, pagination.DefaultPage, p1.CurrentPage())

		// Invalid size (should use default)
		p2 := pagination.NewPaginator(2, 0, 100, nil)
		assert.Equal(t, pagination.DefaultLimit, p2.ItemsPerPage())

		// Negative total items (should treat as 0)
		p3 := pagination.NewPaginator(2, 20, -10, nil)
		assert.Equal(t, 0, p3.TotalItems())
	})

	t.Run("Custom Config", func(t *testing.T) {
		config := &pagination.PaginatorConfig{
			BaseURL:      "https://example.com/api",
			PageParam:    "p",
			SizeParam:    "limit",
			DefaultLimit: 25,
			MaxLimit:     50,
		}

		// Size within limits
		p1 := pagination.NewPaginator(2, 30, 100, config)
		assert.Equal(t, 30, p1.ItemsPerPage())

		// Size exceeds max limit
		p2 := pagination.NewPaginator(2, 80, 100, config)
		assert.Equal(t, 50, p2.ItemsPerPage()) // Should be capped at MaxLimit
	})
}

func TestFromRequest(t *testing.T) {
	t.Run("Parse Standard Parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/api?page=3&size=25", nil)
		p := pagination.FromRequest(req, 100, nil)

		assert.Equal(t, 3, p.CurrentPage())
		assert.Equal(t, 25, p.ItemsPerPage())
		assert.Equal(t, 100, p.TotalItems())
		assert.Equal(t, 4, p.TotalPages())
	})

	t.Run("Invalid Parameters", func(t *testing.T) {
		// Invalid page parameter
		req1 := httptest.NewRequest("GET", "http://example.com/api?page=invalid&size=25", nil)
		p1 := pagination.FromRequest(req1, 100, nil)
		assert.Equal(t, pagination.DefaultPage, p1.CurrentPage())

		// Invalid size parameter
		req2 := httptest.NewRequest("GET", "http://example.com/api?page=3&size=invalid", nil)
		p2 := pagination.FromRequest(req2, 100, nil)
		assert.Equal(t, pagination.DefaultLimit, p2.ItemsPerPage())

		// Negative page
		req3 := httptest.NewRequest("GET", "http://example.com/api?page=-1&size=25", nil)
		p3 := pagination.FromRequest(req3, 100, nil)
		assert.Equal(t, pagination.DefaultPage, p3.CurrentPage())

		// Size exceeds max
		req4 := httptest.NewRequest("GET", "http://example.com/api?page=3&size=200", nil)
		p4 := pagination.FromRequest(req4, 100, nil)
		assert.Equal(t, pagination.MaxLimit, p4.ItemsPerPage())
	})

	t.Run("Custom Configuration", func(t *testing.T) {
		config := &pagination.PaginatorConfig{
			PageParam:    "p",
			SizeParam:    "limit",
			DefaultLimit: 15,
			MaxLimit:     30,
		}

		req := httptest.NewRequest("GET", "http://example.com/api?p=2&limit=20", nil)
		p := pagination.FromRequest(req, 100, config)

		assert.Equal(t, 2, p.CurrentPage())
		assert.Equal(t, 20, p.ItemsPerPage())
		assert.Equal(t, 100, p.TotalItems())
	})

	t.Run("Automatic BaseURL Generation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/api/items?page=2&size=10&filter=active", nil)
		p := pagination.FromRequest(req, 100, nil)

		// Verify URL generation contains the correct host and path
		assert.Contains(t, p.FirstPageURL(), "http://example.com/api/items")

		// Verify query parameters are preserved
		assert.Contains(t, p.FirstPageURL(), "filter=active")
	})

	t.Run("HTTPS Detection", func(t *testing.T) {
		// Create request with TLS info
		req := httptest.NewRequest("GET", "http://example.com/api?page=2&size=10", nil)
		req.TLS = &tls.ConnectionState{}

		p := pagination.FromRequest(req, 100, nil)

		// URLs should use https scheme
		assert.Contains(t, p.FirstPageURL(), "https://")
	})

	t.Run("X-Forwarded-Proto Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/api?page=2&size=10", nil)
		req.Header.Set("X-Forwarded-Proto", "https")

		p := pagination.FromRequest(req, 100, nil)

		// URLs should use https scheme from header
		assert.Contains(t, p.FirstPageURL(), "https://")
	})
}

func TestFromCustomParams(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/api?p=3&limit=15&filter=active", nil)
	p := pagination.FromCustomParams(req, 100, "p", "limit")

	assert.Equal(t, 3, p.CurrentPage())
	assert.Equal(t, 15, p.ItemsPerPage())
	assert.Equal(t, 100, p.TotalItems())

	// Verify URL generation uses custom parameter names
	assert.Contains(t, p.FirstPageURL(), "p=1")
	assert.Contains(t, p.FirstPageURL(), "limit=15")

	// Other query parameters should be preserved
	assert.Contains(t, p.FirstPageURL(), "filter=active")
}

// Legacy API Tests
func TestFromRequestWithOptions(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/api?page=3&size=15", nil)
	p, err := pagination.FromRequestWithOptions(req, 100)

	assert.NoError(t, err)
	assert.Equal(t, 3, p.CurrentPage())
	assert.Equal(t, 15, p.ItemsPerPage())
	assert.Equal(t, 100, p.TotalItems())
}
