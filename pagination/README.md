# Pagination Package

A lightweight pagination utility for Go applications with support for offset/limit pagination.

## Installation

```bash
go get github.com/dmitrymomot/gokit/pagination
```

## Overview

The pagination package provides utilities for implementing pagination in Go applications, particularly in REST APIs. It offers both request parameter parsing and response formatting to standardize pagination across your services.

## Features

- Offset/Limit pagination with configurable defaults
- HTTP request parameter parsing (page, size)
- Response metadata generation (current page, total pages, etc.)
- Validation of pagination parameters
- Helpful error types for parameter validation
- Support for maximum limits to prevent excessive queries

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

func fetchItems(ctx context.Context, offset, limit int) ([]Item, int, error) {
	// Implementation of database query with pagination
	// ...
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

### Structuring Paginated Responses

```go
// Define your response structure
type PaginatedResponse struct {
	Items      []Item            `json:"items"`
	Pagination pagination.PageInfo `json:"pagination"`
}

// Create response with pagination metadata
func createResponse(items []Item, page, size, total int) PaginatedResponse {
	return PaginatedResponse{
		Items:      items,
		Pagination: pagination.NewPageInfo(page, size, total),
	}
}

// Example usage in handler
func ListHandler(w http.ResponseWriter, r *http.Request) {
	page := pagination.GetPageFromRequest(r)
	size := pagination.GetSizeFromRequest(r)
	offset := pagination.CalculateOffset(page, size)
	
	items, total, err := repository.FindItems(r.Context(), offset, size)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	response := createResponse(items, page, size, total)
	renderJSON(w, response)
}
```

### Example Response Structure

```json
{
  "items": [
    { "id": 1, "name": "Item 1" },
    { "id": 2, "name": "Item 2" }
  ],
  "pagination": {
    "page": 2,
    "size": 10,
    "total_items": 58,
    "total_pages": 6
  }
}
```

## API Reference

### Types

```go
// OffsetLimit represents offset and limit pagination parameters
type OffsetLimit struct {
	Offset int
	Limit  int
}

// PageInfo contains pagination metadata for responses
type PageInfo struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}
```

### Functions

```go
// Parse pagination from HTTP request
func ParseOffsetLimitFromRequest(r *http.Request) (OffsetLimit, error)

// Get individual parameters with defaults applied
func GetPageFromRequest(r *http.Request) int
func GetSizeFromRequest(r *http.Request) int

// Calculate offset from page and size
func CalculateOffset(page, size int) int

// Create pagination metadata for responses
func NewPageInfo(page, size, totalItems int) PageInfo
```

### Configuration Constants

```go
// Default values (can be modified)
var DefaultPage = 1
var DefaultLimit = 10
var MaxLimit = 100
```

### Error Types

```go
// Error types for validation
var ErrInvalidPage = errors.New("invalid page parameter")
var ErrInvalidSize = errors.New("invalid size parameter")
```

## Error Handling

- Use `errors.Is()` to check for specific pagination errors (e.g., `ErrInvalidPage`, `ErrInvalidSize`)
- Always validate pagination parameters before executing database queries
- Return appropriate HTTP status codes for pagination errors (usually HTTP 400 Bad Request)

## Best Practices

1. **Parameter Validation**:
   - Always validate pagination parameters before executing expensive database queries
   - Use the provided error types to return specific error messages

2. **Performance**:
   - Enforce a reasonable `MaxLimit` to prevent excessive database load
   - Use database-specific pagination techniques (e.g., keyset pagination for large datasets)
   - Consider adding caching for frequently accessed pages

3. **API Design**:
   - Keep parameter names consistent across endpoints (`page`, `size`)
   - Include pagination metadata in all paginated responses
   - Document pagination behavior in your API documentation

4. **Context Usage**:
   - Pass context to database queries to enable timeout and cancellation
   - Consider using context for additional pagination parameters if needed
