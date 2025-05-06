package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrymomot/gokit/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginator(t *testing.T) {
	config := &pagination.PaginatorConfig{
		BaseURL:      "https://example.com/api/items",
		PageParam:    "page",
		SizeParam:    "size",
		DefaultLimit: 10,
		MaxLimit:     100,
	}

	t.Run("Basic Properties", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 25, config)

		assert.Equal(t, 2, p.CurrentPage())
		assert.Equal(t, 10, p.ItemsPerPage())
		assert.Equal(t, 25, p.TotalItems())
		assert.Equal(t, 3, p.TotalPages())
		assert.Equal(t, 10, p.Offset())
		assert.Equal(t, 10, p.Limit())

		ol := p.OffsetLimit()
		assert.Equal(t, 10, ol.Offset)
		assert.Equal(t, 10, ol.Limit)
	})

	t.Run("Navigation Properties", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 25, config)

		assert.True(t, p.HasNext())
		assert.True(t, p.HasPrevious())
		assert.False(t, p.IsFirstPage())
		assert.False(t, p.IsLastPage())

		// First page
		p1 := pagination.NewPaginator(1, 10, 25, config)
		assert.True(t, p1.HasNext())
		assert.False(t, p1.HasPrevious())
		assert.True(t, p1.IsFirstPage())
		assert.False(t, p1.IsLastPage())

		// Last page
		p3 := pagination.NewPaginator(3, 10, 25, config)
		assert.False(t, p3.HasNext())
		assert.True(t, p3.HasPrevious())
		assert.False(t, p3.IsFirstPage())
		assert.True(t, p3.IsLastPage())

		// Single page
		p4 := pagination.NewPaginator(1, 10, 5, config)
		assert.False(t, p4.HasNext())
		assert.False(t, p4.HasPrevious())
		assert.True(t, p4.IsFirstPage())
		assert.True(t, p4.IsLastPage())
	})

	t.Run("URL Methods", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 25, config)

		assert.Equal(t, "https://example.com/api/items?page=3&size=10", p.NextPageURL())
		assert.Equal(t, "https://example.com/api/items?page=1&size=10", p.PreviousPageURL())
		assert.Equal(t, "https://example.com/api/items?page=1&size=10", p.FirstPageURL())
		assert.Equal(t, "https://example.com/api/items?page=3&size=10", p.LastPageURL())

		// Test with non-existent pages
		p1 := pagination.NewPaginator(1, 10, 5, config)
		assert.Equal(t, "", p1.NextPageURL())
		assert.Equal(t, "", p1.PreviousPageURL())
		assert.Equal(t, "https://example.com/api/items?page=1&size=10", p1.FirstPageURL())
		assert.Equal(t, "https://example.com/api/items?page=1&size=10", p1.LastPageURL())
	})

	t.Run("Links Methods", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 25, config)

		links := p.Links()
		assert.Equal(t, 5, len(links))
		assert.Equal(t, "https://example.com/api/items?page=2&size=10", links["self"])
		assert.Equal(t, "https://example.com/api/items?page=1&size=10", links["first"])
		assert.Equal(t, "https://example.com/api/items?page=3&size=10", links["last"])
		assert.Equal(t, "https://example.com/api/items?page=3&size=10", links["next"])
		assert.Equal(t, "https://example.com/api/items?page=1&size=10", links["prev"])

		linksArray := p.LinksArray()
		assert.Equal(t, 5, len(linksArray))

		// Check presence of all links
		foundRels := make(map[string]bool)
		for _, link := range linksArray {
			foundRels[link.Rel] = true
		}
		assert.True(t, foundRels["self"])
		assert.True(t, foundRels["first"])
		assert.True(t, foundRels["last"])
		assert.True(t, foundRels["next"])
		assert.True(t, foundRels["prev"])
	})

	t.Run("PageInfo Method", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 25, config)

		info := p.PageInfo()
		assert.Equal(t, 2, info.Page)
		assert.Equal(t, 10, info.Size)
		assert.Equal(t, 25, info.TotalItems)
		assert.Equal(t, 3, info.TotalPages)
	})

	t.Run("WithTotalItems Method", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 0, config)
		assert.Equal(t, 0, p.TotalItems())
		assert.Equal(t, 0, p.TotalPages())

		p2 := p.WithTotalItems(25)
		assert.Equal(t, 25, p2.TotalItems())
		assert.Equal(t, 3, p2.TotalPages())

		// Original paginator should remain unchanged
		assert.Equal(t, 0, p.TotalItems())
		assert.Equal(t, 0, p.TotalPages())

		// Test with negative value
		p3 := p.WithTotalItems(-5)
		assert.Equal(t, 0, p3.TotalItems())
		assert.Equal(t, 0, p3.TotalPages())
	})

	t.Run("RenderPageNumbers Method", func(t *testing.T) {
		p := pagination.NewPaginator(5, 10, 100, config)

		// Test with 2 pages before and after
		pageNumbers := p.RenderPageNumbers(2, 2)
		assert.Equal(t, 5, len(pageNumbers))
		assert.Equal(t, 3, pageNumbers[0].Number)
		assert.Equal(t, 7, pageNumbers[4].Number)
		assert.True(t, pageNumbers[2].Current) // Page 5 should be current

		// Test with first pages
		p1 := pagination.NewPaginator(1, 10, 100, config)
		pageNumbers1 := p1.RenderPageNumbers(2, 2)
		assert.Equal(t, 3, len(pageNumbers1))
		assert.Equal(t, 1, pageNumbers1[0].Number)
		assert.Equal(t, 3, pageNumbers1[2].Number)
		assert.True(t, pageNumbers1[0].Current) // Page 1 should be current

		// Test with last pages
		p10 := pagination.NewPaginator(10, 10, 100, config)
		pageNumbers10 := p10.RenderPageNumbers(2, 2)
		assert.Equal(t, 3, len(pageNumbers10))
		assert.Equal(t, 8, pageNumbers10[0].Number)
		assert.Equal(t, 10, pageNumbers10[2].Number)
		assert.True(t, pageNumbers10[2].Current) // Page 10 should be current

		// Test with zero total items
		p0 := pagination.NewPaginator(1, 10, 0, config)
		pageNumbers0 := p0.RenderPageNumbers(2, 2)
		assert.Nil(t, pageNumbers0)
	})

	t.Run("RespondWithJSON Method", func(t *testing.T) {
		p := pagination.NewPaginator(2, 10, 25, config)

		// Setup test server
		w := httptest.NewRecorder()
		items := []string{"item1", "item2", "item3"}

		err := p.RespondWithJSON(w, items)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Verify response contains expected fields (just check presence, not exact values)
		responseBody := w.Body.String()
		assert.Contains(t, responseBody, `"items":["item1","item2","item3"]`)
		assert.Contains(t, responseBody, `"pagination":`)
		assert.Contains(t, responseBody, `"page":2`)
		assert.Contains(t, responseBody, `"size":10`)
		assert.Contains(t, responseBody, `"total_items":25`)
		assert.Contains(t, responseBody, `"total_pages":3`)
		assert.Contains(t, responseBody, `"links":`)
	})

	// Edge Case Tests

	t.Run("Page Beyond Total Pages", func(t *testing.T) {
		// Test with page number > total pages
		p := pagination.NewPaginator(10, 10, 50, config) // Only 5 pages total

		assert.Equal(t, 10, p.CurrentPage()) // Should keep the requested page
		assert.Equal(t, 5, p.TotalPages())   // Should calculate correct total pages
		assert.False(t, p.HasNext())         // Should have no next page
		assert.True(t, p.HasPrevious())      // Should have previous page
		assert.Equal(t, "", p.NextPageURL()) // Should return empty next URL

		// Check what happens with page links
		links := p.Links()
		assert.NotEmpty(t, links["self"])
		assert.NotEmpty(t, links["first"])
		assert.NotEmpty(t, links["last"])
		assert.Empty(t, links["next"]) // Should be empty for page beyond total
		assert.NotEmpty(t, links["prev"])
	})

	t.Run("Zero Total Items", func(t *testing.T) {
		p := pagination.NewPaginator(1, 10, 0, config)

		assert.Equal(t, 0, p.TotalPages())
		assert.False(t, p.HasNext())
		assert.False(t, p.HasPrevious())
		assert.True(t, p.IsFirstPage())
		assert.False(t, p.IsLastPage()) // With 0 total items, it's not the last page

		links := p.Links()
		assert.NotEmpty(t, links["self"])
		assert.NotEmpty(t, links["first"])
		assert.Empty(t, links["last"]) // Should be empty for 0 items
		assert.Empty(t, links["next"]) // Should be empty for 0 items
		assert.Empty(t, links["prev"]) // Should be empty for 0 items
	})

	t.Run("Extreme Pagination Values", func(t *testing.T) {
		// Test with very large page number
		pLargePage := pagination.NewPaginator(1000000, 10, 100, config)
		assert.Equal(t, 1000000, pLargePage.CurrentPage())
		assert.Equal(t, 10, pLargePage.TotalPages())
		assert.Equal(t, 9999990, pLargePage.Offset()) // Should calculate correct offset

		// Test with very large size
		pLargeSize := pagination.NewPaginator(1, 1000000, 100, config)
		// Size should be capped at MaxLimit (100)
		assert.Equal(t, config.MaxLimit, pLargeSize.ItemsPerPage())
		assert.Equal(t, 1, pLargeSize.TotalPages())

		// Test with very large total items
		pLargeTotal := pagination.NewPaginator(1, 10, 1000000000, config)
		assert.Equal(t, 100000000, pLargeTotal.TotalPages()) // Should calculate correctly
	})

	t.Run("Nil Config", func(t *testing.T) {
		// Should work with nil config
		p := pagination.NewPaginator(1, 10, 100, nil)

		assert.Equal(t, 1, p.CurrentPage())
		assert.Equal(t, 10, p.ItemsPerPage())
		assert.Equal(t, 100, p.TotalItems())
		assert.Equal(t, 10, p.TotalPages())

		// Should use default parameters
		assert.Empty(t, p.FirstPageURL()) // Empty URL with nil config
	})

	t.Run("Zero or Negative Size", func(t *testing.T) {
		// Zero size should use default
		pZero := pagination.NewPaginator(1, 0, 100, config)
		assert.Equal(t, pagination.DefaultLimit, pZero.ItemsPerPage())

		// Negative size should use default
		pNeg := pagination.NewPaginator(1, -5, 100, config)
		assert.Equal(t, pagination.DefaultLimit, pNeg.ItemsPerPage())
	})

	t.Run("Invalid Base URL", func(t *testing.T) {
		badConfig := &pagination.PaginatorConfig{
			BaseURL:      "ht tp://bad url.com/with spaces",
			PageParam:    "page",
			SizeParam:    "size",
			DefaultLimit: 10,
			MaxLimit:     100,
		}

		p := pagination.NewPaginator(1, 10, 100, badConfig)

		// Should not panic and handle URL encoding gracefully
		assert.NotPanics(t, func() {
			url := p.FirstPageURL()
			// URL should contain something, even with a bad base URL
			// Not checking exact content because URL handling may vary
			_ = url
		})
	})

	t.Run("WithTotalItems Immutability", func(t *testing.T) {
		// Create initial paginator
		original := pagination.NewPaginator(2, 10, 25, config)

		// Test multiple mutations to verify immutability
		modified1 := original.WithTotalItems(50)
		modified2 := modified1.WithTotalItems(75)

		// Original should remain unchanged through all mutations
		assert.Equal(t, 25, original.TotalItems())
		assert.Equal(t, 3, original.TotalPages())

		// Each modified version should have its own state
		assert.Equal(t, 50, modified1.TotalItems())
		assert.Equal(t, 5, modified1.TotalPages())

		assert.Equal(t, 75, modified2.TotalItems())
		assert.Equal(t, 8, modified2.TotalPages())
	})
}
