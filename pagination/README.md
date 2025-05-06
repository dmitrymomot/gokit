# Pagination Package

A flexible and feature-rich pagination library for Go web applications with immutable API design.

## Installation

```bash
go get github.com/dmitrymomot/gokit/pagination
```

## Overview

The pagination package provides tools for implementing pagination in web applications with a focus on immutability and flexibility. It supports offset/limit and page/size pagination styles, generates navigation links, and includes middleware for HTTP handlers. The package is designed to be thread-safe and easy to integrate with any Go web framework.

## Features

- Immutable pagination with a fluent API design
- Built-in HTTP middleware for easy integration
- Support for custom pagination parameters and configuration
- Automatic URL generation for navigation links (first, prev, next, last)
- JSON response formatting for RESTful APIs
- Thread-safe implementation for concurrent use
- Comprehensive pagination metadata (current page, total pages, etc.)
- UI pagination support with page number generation

## Usage

### Basic Example

```go
import (
    "net/http"
    "github.com/dmitrymomot/gokit/pagination"
)

func handler(w http.ResponseWriter, r *http.Request) {
    // Create a paginator from the request
    paginator := pagination.FromRequest(r, 100, nil)
    
    // Get offset and limit for database query
    ol := paginator.OffsetLimit()
    
    // Query your database using offset and limit
    items := fetchItems(ol.Offset, ol.Limit)
    
    // Respond with paginated JSON
    paginator.RespondWithJSON(w, items)
    // Returns a structured JSON response with items and pagination metadata
}
```

### With Custom Configuration

```go
// Create custom pagination configuration
config := &pagination.PaginatorConfig{
    PageParam:    "p",            // Use "p" instead of "page"
    SizeParam:    "limit",        // Use "limit" instead of "size"
    DefaultLimit: 25,             // Default items per page
    MaxLimit:     50,             // Maximum allowed items per page
    BaseURL:      "/api/items",   // Base URL for link generation
}

// Create paginator with custom config
paginator := pagination.NewPaginator(2, 25, 100, config)

// Get page information for templates
pageInfo := paginator.PageInfo()
// pageInfo contains page=2, size=25, total_items=100, total_pages=4
```

### Using Middleware

```go
import (
    "net/http"
    "github.com/dmitrymomot/gokit/pagination"
)

func main() {
    // Create pagination middleware with default configuration
    paginationMiddleware := pagination.Middleware(nil)
    
    // Apply middleware to your handler
    http.Handle("/api/items", paginationMiddleware(http.HandlerFunc(itemsHandler)))
    
    http.ListenAndServe(":8080", nil)
}

func itemsHandler(w http.ResponseWriter, r *http.Request) {
    // Get paginator from request context
    paginator := pagination.GetPaginatorFromContext(r.Context())
    
    // Get pagination parameters
    offset := paginator.Offset()
    limit := paginator.Limit()
    
    // Fetch items from database
    items, totalCount := fetchItemsWithCount(offset, limit)
    
    // Update paginator with total items count
    updatedPaginator := paginator.WithTotalItems(totalCount)
    
    // Respond with paginated data
    updatedPaginator.RespondWithJSON(w, items)
}
```

### Error Handling

```go
// Legacy function example with error handling
paginator, err := pagination.FromRequestWithOptions(r, totalItems)
if err != nil {
    switch {
    case errors.Is(err, pagination.ErrNegativeTotalItems):
        // Handle negative total items error
        http.Error(w, "Invalid total items count", http.StatusBadRequest)
    case errors.Is(err, pagination.ErrInvalidPage):
        // Handle invalid page number
        http.Error(w, "Invalid page number", http.StatusBadRequest)
    case errors.Is(err, pagination.ErrInvalidSize):
        // Handle invalid page size
        http.Error(w, "Invalid page size", http.StatusBadRequest)
    default:
        // Handle other errors
        http.Error(w, "Pagination error", http.StatusInternalServerError)
    }
    return
}
```

## Best Practices

1. **Configuration**:
   - Set appropriate default and maximum limits to prevent excessive database queries
   - Customize parameter names to match your API conventions
   - Include a descriptive base URL for proper link generation

2. **Performance**:
   - For large datasets, use COUNT queries optimized for your database
   - Consider caching total counts that change infrequently
   - Use database-specific optimizations for offset/limit queries

3. **API Design**:
   - Include pagination metadata in all list responses
   - Provide navigation links (first, prev, next, last) for client usability
   - Use consistent parameter naming across your API

4. **Error Handling**:
   - Handle invalid pagination parameters gracefully with sensible defaults
   - Provide clear error messages for debugging

## API Reference

### Configuration Variables

```go
const (
    DefaultLimit = 10  // Default number of items per page
    MaxLimit     = 100 // Maximum allowed items per page
    DefaultPage  = 1   // Default page number
)
```

### Types

```go
type PaginatorConfig struct {
    BaseURL      string     // Base URL for generating pagination links
    PageParam    string     // Name of the query parameter for page number
    SizeParam    string     // Name of the query parameter for page size
    DefaultLimit int        // Default number of items per page
    MaxLimit     int        // Maximum allowed number of items per page
    QueryParams  url.Values // Additional query parameters to include in links
}
```

```go
type PageInfo struct {
    Page       int `json:"page"`
    Size       int `json:"size"`
    TotalItems int `json:"total_items"`
    TotalPages int `json:"total_pages"`
}
```

```go
type OffsetLimit struct {
    Offset int
    Limit  int
}
```

```go
type PageNumber struct {
    Number  int    // The page number
    URL     string // The URL for this page
    Current bool   // Whether this is the current page
}
```

```go
type Link struct {
    Rel string // Relation type (first, prev, next, last)
    URL string // The URL for this link
}
```

```go
type PaginatedResponse[T any] struct {
    Items      []T              `json:"items"`
    Pagination PageInfo         `json:"pagination"`
    Links      map[string]string `json:"links"`
}
```

### Functions

```go
func DefaultConfig() *PaginatorConfig
```
Returns a default configuration for the paginator.

```go
func NewPaginator(page, size, totalItems int, config *PaginatorConfig) Paginator
```
Creates a new Paginator with explicit values and configuration.

```go
func FromRequest(r *http.Request, totalItems int, config *PaginatorConfig) Paginator
```
Creates a Paginator from an HTTP request with optional configuration.

```go
func FromCustomParams(r *http.Request, totalItems int, pageParam, sizeParam string) Paginator
```
Creates a Paginator from custom request parameters.

```go
func Middleware(config *PaginatorConfig) func(http.Handler) http.Handler
```
Creates middleware that extracts pagination parameters from requests.

```go
func GetPaginatorFromContext(ctx context.Context) Paginator
```
Retrieves the Paginator from the request context.

```go
func NewPaginatedResponse[T any](items []T, paginator Paginator) PaginatedResponse[T]
```
Creates a new paginated response with the given items and paginator.

```go
func RespondWithJSON[T any](paginator Paginator, items []T, w http.ResponseWriter) error
```
Helper function to write a paginated JSON response.

### Methods

```go
func (p Paginator) Offset() int
```
Returns the calculated offset for database queries.

```go
func (p Paginator) Limit() int
```
Returns the number of items per page.

```go
func (p Paginator) OffsetLimit() OffsetLimit
```
Returns both offset and limit in a single struct.

```go
func (p Paginator) CurrentPage() int
func (p Paginator) ItemsPerPage() int
func (p Paginator) TotalItems() int
func (p Paginator) TotalPages() int
```
Returns pagination state values.

```go
func (p Paginator) HasNext() bool
func (p Paginator) HasPrevious() bool
func (p Paginator) IsFirstPage() bool
func (p Paginator) IsLastPage() bool
```
Returns navigation state helpers.

```go
func (p Paginator) NextPageURL() string
func (p Paginator) PreviousPageURL() string
func (p Paginator) FirstPageURL() string
func (p Paginator) LastPageURL() string
```
Returns URLs for navigation.

```go
func (p Paginator) Links() map[string]string
func (p Paginator) LinksArray() []Link
```
Returns collection of navigation links.

```go
func (p Paginator) PageInfo() PageInfo
```
Returns pagination metadata for API responses.

```go
func (p Paginator) WithTotalItems(totalItems int) Paginator
```
Creates a new Paginator with updated total items count.

```go
func (p Paginator) RenderPageNumbers(before, after int) []PageNumber
```
Returns page numbers for template rendering.

```go
func (p Paginator) RespondWithJSON(w http.ResponseWriter, items any) error
```
Writes a standard paginated JSON response.

### Error Types

```go
var (
    ErrInvalidLimit       = errors.New("invalid limit value")
    ErrInvalidOffset      = errors.New("invalid offset value")
    ErrInvalidCursor      = errors.New("invalid cursor value")
    ErrInvalidPage        = errors.New("invalid page number")
    ErrInvalidSize        = errors.New("invalid page size")
    ErrMissingBaseURL     = errors.New("base URL is required for link generation")
    ErrNegativeTotalItems = errors.New("total items cannot be negative")
    ErrInvalidQueryParam  = errors.New("invalid query parameter")
)
```