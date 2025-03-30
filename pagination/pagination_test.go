package pagination_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/pagination"
)

func TestPlaceholder(t *testing.T) {
	assert.True(t, true, "Placeholder test")
}

func TestCalculateOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		size     int
		expected int
	}{
		{"Valid page and size", 2, 10, 10},
		{"First page", 1, 20, 0},
		{"Invalid page (negative)", -1, 10, 0},
		{"Invalid page (zero)", 0, 10, 0},
		{"Invalid size (negative)", 2, -5, 10}, // Uses DefaultLimit
		{"Invalid size (zero)", 3, 0, 20},     // Uses DefaultLimit
		{"Large page number", 100, 10, 990},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pagination.CalculateOffset(tt.page, tt.size))
		})
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		size       int
		expected   int
	}{
		{"Exact multiple", 100, 10, 10},
		{"Remainder", 105, 10, 11},
		{"Less than size", 5, 10, 1},
		{"Zero items", 0, 10, 0},
		{"Invalid size (negative)", 100, -5, 0},
		{"Invalid size (zero)", 100, 0, 0},
		{"Large number of items", 999, pagination.MaxLimit, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pagination.CalculateTotalPages(tt.totalItems, tt.size))
		})
	}
}

func TestNewPageInfo(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		size       int
		totalItems int
		expected   pagination.PageInfo
	}{
		{
			name:       "Valid case",
			page:       2,
			size:       10,
			totalItems: 105,
			expected: pagination.PageInfo{
				Page:       2,
				Size:       10,
				TotalItems: 105,
				TotalPages: 11,
			},
		},
		{
			name:       "First page",
			page:       1,
			size:       20,
			totalItems: 50,
			expected: pagination.PageInfo{
				Page:       1,
				Size:       20,
				TotalItems: 50,
				TotalPages: 3,
			},
		},
		{
			name:       "Invalid page (negative)",
			page:       -1,
			size:       10,
			totalItems: 30,
			expected: pagination.PageInfo{
				Page:       pagination.DefaultPage, // Defaults to 1
				Size:       10,
				TotalItems: 30,
				TotalPages: 3,
			},
		},
		{
			name:       "Invalid size (zero)",
			page:       3,
			size:       0,
			totalItems: 45,
			expected: pagination.PageInfo{
				Page:       3,
				Size:       pagination.DefaultLimit, // Defaults to 10
				TotalItems: 45,
				TotalPages: 5, // 45 / 10 = 4.5 -> 5
			},
		},
		{
			name:       "Zero total items",
			page:       1,
			size:       10,
			totalItems: 0,
			expected: pagination.PageInfo{
				Page:       1,
				Size:       10,
				TotalItems: 0,
				TotalPages: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pagination.NewPageInfo(tt.page, tt.size, tt.totalItems))
		})
	}
}

func TestParseOffsetLimitFromRequest(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    pagination.OffsetLimit
		expectedErr error
	}{
		{
			name: "Valid page and size",
			url:  "/?page=2&size=20",
			expected: pagination.OffsetLimit{
				Offset: 20, // (2-1) * 20
				Limit:  20,
			},
			expectedErr: nil,
		},
		{
			name: "Default values",
			url:  "/",
			expected: pagination.OffsetLimit{
				Offset: 0, // (1-1) * 10
				Limit:  pagination.DefaultLimit,
			},
			expectedErr: nil,
		},
		{
			name: "Only page specified",
			url:  "/?page=3",
			expected: pagination.OffsetLimit{
				Offset: 20, // (3-1) * 10
				Limit:  pagination.DefaultLimit,
			},
			expectedErr: nil,
		},
		{
			name: "Only size specified",
			url:  "/?size=50",
			expected: pagination.OffsetLimit{
				Offset: 0, // (1-1) * 50
				Limit:  50,
			},
			expectedErr: nil,
		},
		{
			name: "Size exceeding max limit",
			url:  "/?size=200",
			expected: pagination.OffsetLimit{
				Offset: 0, // (1-1) * 100
				Limit:  pagination.MaxLimit,
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid page (string)",
			url:         "/?page=abc",
			expected:    pagination.OffsetLimit{}, // Expect zero value on error
			expectedErr: pagination.ErrInvalidPage,
		},
		{
			name:        "Invalid page (zero)",
			url:         "/?page=0",
			expected:    pagination.OffsetLimit{}, // Expect zero value on error
			expectedErr: pagination.ErrInvalidPage,
		},
		{
			name:        "Invalid size (string)",
			url:         "/?size=xyz",
			expected:    pagination.OffsetLimit{}, // Expect zero value on error
			expectedErr: pagination.ErrInvalidSize,
		},
		{
			name:        "Invalid size (negative)",
			url:         "/?size=-5",
			expected:    pagination.OffsetLimit{}, // Expect zero value on error
			expectedErr: pagination.ErrInvalidSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			result, err := pagination.ParseOffsetLimitFromRequest(req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Equal(t, pagination.OffsetLimit{}, result) // Check for zero value on error
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetPageFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"Valid page", "/?page=5", 5},
		{"Missing page", "/", pagination.DefaultPage},
		{"Invalid page (string)", "/?page=abc", pagination.DefaultPage},
		{"Invalid page (zero)", "/?page=0", pagination.DefaultPage},
		{"Invalid page (negative)", "/?page=-2", pagination.DefaultPage},
		{"Page with other params", "/?other=val&page=3", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			assert.Equal(t, tt.expected, pagination.GetPageFromRequest(req))
		})
	}
}

func TestGetSizeFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected int
	}{
		{"Valid size", "/?size=50", 50},
		{"Missing size", "/", pagination.DefaultLimit},
		{"Invalid size (string)", "/?size=xyz", pagination.DefaultLimit},
		{"Invalid size (zero)", "/?size=0", pagination.DefaultLimit},
		{"Invalid size (negative)", "/?size=-10", pagination.DefaultLimit},
		{"Size exceeding max limit", "/?size=150", pagination.MaxLimit},
		{"Size equal to max limit", "/?size=100", pagination.MaxLimit},
		{"Size with other params", "/?other=val&size=25", 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			assert.Equal(t, tt.expected, pagination.GetSizeFromRequest(req))
		})
	}
}
