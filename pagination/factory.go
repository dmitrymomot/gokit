package pagination

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// PaginatorOption represents a configuration option for creating a Paginator.
// This is maintained for backward compatibility with the legacy API.
type PaginatorOption func(*Paginator)

// NewPaginator creates a new Paginator with explicit values and configuration.
// If config is nil, the default configuration will be used.
// Invalid inputs are handled internally by using defaults.
func NewPaginator(page, size, totalItems int, config *PaginatorConfig) Paginator {
	// Validate inputs
	if page < 1 {
		page = DefaultPage
	}

	if size < 1 {
		size = DefaultLimit
	}

	// Handle negative totalItems by treating it as 0
	if totalItems < 0 {
		totalItems = 0
	}

	// Merge with default config
	config = mergeWithDefaults(config)

	// Apply config limits
	if size > config.MaxLimit {
		size = config.MaxLimit
	}

	// Calculate total pages
	totalPages := calculateTotalPages(totalItems, size)

	// Create paginator
	return Paginator{
		currentPage:  page,
		itemsPerPage: size,
		totalItems:   totalItems,
		totalPages:   totalPages,
		config:       config,
	}
}

// FromRequest creates a Paginator from an HTTP request with optional configuration.
// If config is nil, the default configuration will be used.
// The totalItems parameter can be 0 if not known yet.
// Any parsing errors are handled internally by using default values.
func FromRequest(r *http.Request, totalItems int, config *PaginatorConfig) Paginator {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	// Extract pagination parameters from the request
	pageStr := r.URL.Query().Get(config.PageParam)
	sizeStr := r.URL.Query().Get(config.SizeParam)

	// Parse page parameter
	page := DefaultPage
	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	// Parse size parameter
	size := config.DefaultLimit
	if sizeStr != "" {
		parsedSize, err := strconv.Atoi(sizeStr)
		if err == nil && parsedSize > 0 {
			size = parsedSize
			if size > config.MaxLimit {
				size = config.MaxLimit
			}
		}
	}

	// Handle negative totalItems by treating it as 0
	if totalItems < 0 {
		totalItems = 0
	}

	// If config doesn't have a base URL, create one from the request
	localConfig := *config
	if localConfig.BaseURL == "" {
		localConfig.BaseURL = fmt.Sprintf("%s://%s%s",
			extractScheme(r),
			r.Host,
			r.URL.Path,
		)
	}

	// Copy query parameters if needed
	if localConfig.QueryParams == nil {
		localConfig.QueryParams = url.Values{}
	}

	// Add any query parameters not related to pagination
	for key, values := range r.URL.Query() {
		if key != localConfig.PageParam && key != localConfig.SizeParam {
			for _, value := range values {
				localConfig.QueryParams.Add(key, value)
			}
		}
	}

	// Create the paginator
	return NewPaginator(page, size, totalItems, &localConfig)
}

// FromCustomParams creates a Paginator from custom request parameters.
// It lets you define custom parameter names for page and size.
func FromCustomParams(r *http.Request, totalItems int, pageParam, sizeParam string) Paginator {
	// Create a config with custom parameter names
	config := DefaultConfig()
	config.PageParam = pageParam
	config.SizeParam = sizeParam

	// Use the FromRequest function with our custom config
	return FromRequest(r, totalItems, config)
}

// Deprecated: Use FromRequest instead.
// FromRequestWithOptions creates a Paginator from an HTTP request with custom options.
// This function is maintained for backward compatibility.
func FromRequestWithOptions(r *http.Request, totalItems int, options ...PaginatorOption) (*Paginator, error) {
	// Create a default paginator
	paginator, err := legacyFromRequest(r, totalItems)
	if err != nil {
		return nil, err
	}

	// Apply any custom options
	for _, option := range options {
		option(paginator)
	}

	return paginator, nil
}

// Deprecated: Use FromRequest instead.
// legacyFromRequest is an internal helper function for backward compatibility.
func legacyFromRequest(r *http.Request, totalItems int) (*Paginator, error) {
	return legacyFromCustomParams(r, totalItems, "page", "size")
}

// Deprecated: Use FromRequest instead.
// legacyFromCustomParams is an internal helper function for backward compatibility.
func legacyFromCustomParams(r *http.Request, totalItems int, pageParam, sizeParam string) (*Paginator, error) {
	if totalItems < 0 {
		return nil, ErrNegativeTotalItems
	}

	// Extract pagination parameters
	page, size, err := extractPaginationParams(r, pageParam, sizeParam)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := calculateTotalPages(totalItems, size)

	// Get base URL for link generation
	baseURL := fmt.Sprintf("%s://%s%s",
		extractScheme(r),
		r.Host,
		r.URL.Path,
	)

	// Create a copy of query parameters
	queryParams := url.Values{}
	for key, values := range r.URL.Query() {
		for _, value := range values {
			queryParams.Add(key, value)
		}
	}

	// Create legacy paginator
	return &Paginator{
		currentPage:  page,
		itemsPerPage: size,
		totalItems:   totalItems,
		totalPages:   totalPages,
		config: &PaginatorConfig{
			BaseURL:      baseURL,
			PageParam:    pageParam,
			SizeParam:    sizeParam,
			DefaultLimit: DefaultLimit,
			MaxLimit:     MaxLimit,
			QueryParams:  queryParams,
		},
	}, nil
}

// extractScheme determines the URL scheme from the request.
func extractScheme(r *http.Request) string {
	scheme := "http"

	// Check for X-Forwarded-Proto header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}

	// Check if the request is using TLS
	if r.TLS != nil {
		scheme = "https"
	}

	return scheme
}

// Deprecated: Used only by legacy functions.
// extractPaginationParams extracts pagination parameters from the request.
func extractPaginationParams(r *http.Request, pageParam, sizeParam string) (page, size int, err error) {
	// Set defaults
	page = DefaultPage
	size = DefaultLimit

	// Parse page parameter
	pageStr := r.URL.Query().Get(pageParam)
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return 0, 0, ErrInvalidPage
		}
	}

	// Parse size parameter
	sizeStr := r.URL.Query().Get(sizeParam)
	if sizeStr != "" {
		size, err = strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			return 0, 0, ErrInvalidSize
		}
		if size > MaxLimit {
			size = MaxLimit
		}
	}

	return page, size, nil
}
