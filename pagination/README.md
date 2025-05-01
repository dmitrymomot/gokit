# Pagination Package

A lightweight pagination utility for Go applications with support for offset/limit pagination.

## Installation

```bash
go get github.com/dmitrymomot/gokit/pagination
```

## Overview

The pagination package provides utilities for implementing pagination in Go applications, particularly in REST APIs. It offers both request parameter parsing and response formatting to standardize pagination across your services. The package is thread-safe and designed for use in concurrent applications.

## Features

- Offset/limit pagination with configurable defaults
- HTTP request parameter parsing for `page` and `size` query parameters
- Response metadata generation (current page, total pages, etc.)
- Validation of pagination parameters with helpful error types
- Maximum limit enforcement to prevent excessive queries
- Thread-safe operations with no shared mutable state

## Usage

### Parsing Pagination from HTTP Request

```go
import (
	"context"
	"errors"
	"net/http"
	
	"github.com/dmitrymomot/gokit/pagination"
)

func ListHandler(w http.ResponseWriter, r *http.Request) {
	// Automatically parse page and size from query parameters
	// e.g., /api/items?page=2&size=20
	offsetLimit, err := pagination.ParseOffsetLimitFromRequest(r)
	if err != nil {
		// Handle validation errors
		if errors.Is(err, pagination.ErrInvalidPage) || errors.Is(err, pagination.ErrInvalidSize) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Use in database query
	ctx := r.Context()
	items, totalCount, err := fetchItems(ctx, offsetLimit.Offset, offsetLimit.Limit)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	// Create pagination info for response
	pageInfo := pagination.NewPageInfo(
		pagination.GetPageFromRequest(r),
		offsetLimit.Limit,
		totalCount,
	)
	
	// Return paginated response
	renderJSON(w, map[string]interface{}{
		"data":       items,
		"pagination": pageInfo,
	})
}
```

### Manual Pagination Parameter Extraction

```go
// Get individual pagination parameters
page := pagination.GetPageFromRequest(r) // Default: 1
size := pagination.GetSizeFromRequest(r) // Default: 10, Max: 100

// Calculate offset for database query
offset := pagination.CalculateOffset(page, size)

// Use in database query
items, totalCount, err := fetchItems(ctx, offset, size)
```

### Creating Pagination Response Metadata

```go
// Define your response structure
type PaginatedResponse struct {
	Items      []Item             `json:"items"`
	Pagination pagination.PageInfo `json:"pagination"`
}

// Create response with pagination metadata
func createResponse(items []Item, page, size, total int) PaginatedResponse {
	return PaginatedResponse{
		Items:      items,
		Pagination: pagination.NewPageInfo(page, size, total),
	}
}

// Example response format:
// {
//   "items": [
//     { "id": 1, "name": "Item 1" },
//     { "id": 2, "name": "Item 2" }
//   ],
//   "pagination": {
//     "page": 2,
//     "size": 10,
//     "total_items": 58,
//     "total_pages": 6
//   }
// }
```

## Best Practices

1. **Validation**:
   - Always validate pagination parameters before executing database queries
   - Use `errors.Is()` to check for specific pagination errors
   - Apply sensible defaults for missing parameters

2. **Performance**:
   - Consider adding database indexes on fields used for sorting in paginated queries
   - For large data sets, use keyset/cursor pagination instead of offset/limit
   - Limit maximum page size to prevent excessive resource usage

3. **API Design**:
   - Use consistent parameter names across your APIs (`page`, `size`)
   - Include pagination metadata in responses to help clients navigate
   - Provide documentation on pagination limits and defaults

## API Reference

### Constants

```go
var DefaultLimit = 10 // Default number of items per page
var MaxLimit = 100    // Maximum allowed items per page
var DefaultPage = 1   // Default page number
```

### Types

```go
// OffsetLimit represents offset and limit pagination parameters
type OffsetLimit struct {
	Offset int // Starting position for database query
	Limit  int // Number of items to retrieve
}

// PageInfo contains pagination metadata for responses
type PageInfo struct {
	Page       int `json:"page"`       // Current page number
	Size       int `json:"size"`       // Number of items per page
	TotalItems int `json:"total_items"` // Total number of items
	TotalPages int `json:"total_pages"` // Total number of pages
}
```

### Functions

```go
func ParseOffsetLimitFromRequest(r *http.Request) (OffsetLimit, error)
```
Parses page and size parameters from an HTTP request and returns calculated offset and limit values.

```go
func GetPageFromRequest(r *http.Request) int
```
Retrieves the page parameter from request with validation and default handling.

```go
func GetSizeFromRequest(r *http.Request) int
```
Retrieves the size parameter from request with validation, default handling, and maximum enforcement.

```go
func CalculateOffset(page, size int) int
```
Calculates the database offset based on page number and size.

```go
func CalculateTotalPages(totalItems, size int) int
```
Calculates the total number of pages based on total items and page size.

```go
func NewPageInfo(page, size, totalItems int) PageInfo
```
Creates a new PageInfo struct with calculated total pages.

### Error Types

```go
var ErrInvalidPage = errors.New("invalid page number")
var ErrInvalidSize = errors.New("invalid page size")
var ErrInvalidLimit = errors.New("invalid limit value")
var ErrInvalidOffset = errors.New("invalid offset value")
var ErrInvalidCursor = errors.New("invalid cursor value")
```
