package pagination

import (
	"math"
	"net/http"
	"strconv"
)

// Package pagination provides utilities for handling pagination in APIs.

const (
	DefaultLimit = 10
	MaxLimit     = 100
	DefaultPage  = 1
)

// PageInfo represents pagination details for offset-based pagination.
// It is designed to be embedded in API response structs.
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

// CalculateOffset calculates the database offset based on page number and size.
func CalculateOffset(page, size int) int {
	if page < 1 {
		page = DefaultPage
	}
	if size < 1 {
		size = DefaultLimit
	}
	return (page - 1) * size
}

// CalculateTotalPages calculates the total number of pages based on total items and page size.
func CalculateTotalPages(totalItems, size int) int {
	if totalItems <= 0 || size <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(size)))
}

// NewPageInfo creates a new PageInfo struct.
func NewPageInfo(page, size, totalItems int) PageInfo {
	if page < 1 {
		page = DefaultPage
	}
	if size < 1 {
		size = DefaultLimit
	}

	totalPages := CalculateTotalPages(totalItems, size)

	return PageInfo{
		Page:       page,
		Size:       size,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// ParseOffsetLimitFromRequest parses page and size from HTTP request query parameters
// and returns the calculated OffsetLimit struct.
// It applies default values and enforces maximum limits.
func ParseOffsetLimitFromRequest(r *http.Request) (OffsetLimit, error) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page := DefaultPage
	limit := DefaultLimit
	var err error

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return OffsetLimit{}, ErrInvalidPage
		}
	}

	if sizeStr != "" {
		limit, err = strconv.Atoi(sizeStr)
		if err != nil || limit <= 0 {
			return OffsetLimit{}, ErrInvalidSize
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}
	}

	offset := CalculateOffset(page, limit)

	return OffsetLimit{
		Offset: offset,
		Limit:  limit,
	}, nil
}

// GetPageFromRequest retrieves the 'page' query parameter from the request.
// Returns DefaultPage if the parameter is missing or invalid.
func GetPageFromRequest(r *http.Request) int {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		return DefaultPage
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		return DefaultPage
	}
	return page
}

// GetSizeFromRequest retrieves the 'size' query parameter from the request.
// Returns DefaultLimit if the parameter is missing or invalid.
// Enforces MaxLimit.
func GetSizeFromRequest(r *http.Request) int {
	sizeStr := r.URL.Query().Get("size")
	if sizeStr == "" {
		return DefaultLimit
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return DefaultLimit
	}
	if size > MaxLimit {
		size = MaxLimit
	}
	return size
}
