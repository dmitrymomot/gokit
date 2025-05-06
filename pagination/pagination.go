package pagination

import (
	"net/url"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
	DefaultPage  = 1
)

// PaginatorConfig holds configuration options for the paginator.
type PaginatorConfig struct {
	// BaseURL is the base URL for generating pagination links.
	BaseURL string

	// PageParam is the name of the query parameter for the page number.
	PageParam string

	// SizeParam is the name of the query parameter for the page size.
	SizeParam string

	// DefaultLimit is the default number of items per page.
	DefaultLimit int

	// MaxLimit is the maximum allowed number of items per page.
	MaxLimit int

	// QueryParams are additional query parameters to include in pagination links.
	QueryParams url.Values
}

// DefaultConfig returns a default configuration for the paginator.
func DefaultConfig() *PaginatorConfig {
	return &PaginatorConfig{
		BaseURL:      "",
		PageParam:    "page",
		SizeParam:    "size",
		DefaultLimit: DefaultLimit,
		MaxLimit:     MaxLimit,
		QueryParams:  url.Values{},
	}
}

// mergeWithDefaults merges the given configuration with the default configuration.
// The values from the provided configuration take precedence over the default values.
// If a field in the provided configuration is empty or zero, the corresponding value
// from the default configuration is used.
func mergeWithDefaults(config *PaginatorConfig) *PaginatorConfig {
	// If the config is nil, return default config
	if config == nil {
		return DefaultConfig()
	}

	// Start with default config
	defaults := DefaultConfig()

	// Create a new config with merged values
	merged := &PaginatorConfig{
		// Use current values if set, otherwise use defaults
		BaseURL:      valueOrDefault(config.BaseURL, defaults.BaseURL),
		PageParam:    valueOrDefault(config.PageParam, defaults.PageParam),
		SizeParam:    valueOrDefault(config.SizeParam, defaults.SizeParam),
		DefaultLimit: valueOrDefault(config.DefaultLimit, defaults.DefaultLimit),
		MaxLimit:     valueOrDefault(config.MaxLimit, defaults.MaxLimit),
	}

	// Handle QueryParams specially to merge them
	if config.QueryParams != nil {
		// Copy existing query params
		merged.QueryParams = url.Values{}
		for k, v := range config.QueryParams {
			for _, val := range v {
				merged.QueryParams.Add(k, val)
			}
		}
	} else {
		merged.QueryParams = url.Values{}
	}

	return merged
}

// PageInfo represents pagination details for API responses.
type PageInfo struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// OffsetLimit contains the calculated offset and limit for database queries.
type OffsetLimit struct {
	Offset int
	Limit  int
}

// PageNumber represents a page number for UI rendering in templates.
type PageNumber struct {
	Number  int    // The page number
	URL     string // The URL for this page
	Current bool   // Whether this is the current page
}

// Link represents a pagination link with rel attribute for API responses.
type Link struct {
	Rel string // Relation type (first, prev, next, last, etc.)
	URL string // The URL for this link
}

// valueOrDefault returns value if it's not zero, otherwise returns defaultValue.
// For strings, zero means empty string. For integers, zero means <= 0.
func valueOrDefault[T comparable](value, defaultValue T) T {
	switch v := any(value).(type) {
	case string:
		if v == "" {
			return defaultValue
		}
	case int:
		if v <= 0 {
			return defaultValue
		}
	}
	return value
}
