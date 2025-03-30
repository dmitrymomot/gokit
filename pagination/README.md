# Pagination Package

This package provides utilities for implementing pagination in Go applications, supporting both offset/limit and cursor-based strategies.

## Features

- Offset/Limit Pagination
- Cursor-Based Pagination (Planned)
- Helper functions for parsing request parameters
- Standardized response structures
- Customizable defaults (e.g., `DefaultLimit`, `MaxLimit`, `DefaultPage`)

## Installation

```bash
go get github.com/dmitrymomot/gokit/pagination
```

## Offset/Limit Pagination

### Parsing Request Parameters

You can parse `page` and `size` query parameters from an `http.Request` to get the calculated `OffsetLimit` for your database queries.

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/pagination"
)

func handler(w http.ResponseWriter, r *http.Request) {
	// Parse page and size from query parameters (e.g., /items?page=2&size=20)
	offsetLimit, err := pagination.ParseOffsetLimitFromRequest(r)
	if err != nil {
		// Handle invalid pagination parameters (e.g., return 400 Bad Request)
		if errors.Is(err, pagination.ErrInvalidPage) || errors.Is(err, pagination.ErrInvalidSize) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// Handle other potential errors
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Query Offset: %d, Limit: %d", offsetLimit.Offset, offsetLimit.Limit)

	// Use offsetLimit.Offset and offsetLimit.Limit in your database query
	// results, totalItems := database.GetItems(offsetLimit.Offset, offsetLimit.Limit)
}
```

Alternatively, you can parse `page` and `size` individually:

```go
page := pagination.GetPageFromRequest(r) // Returns DefaultPage (1) if invalid/missing
size := pagination.GetSizeFromRequest(r) // Returns DefaultLimit (10), enforces MaxLimit (100) if invalid/missing
offset := pagination.CalculateOffset(page, size)

// Use offset and size in your query
```

### Creating Response Structures

Use `NewPageInfo` to create a `PageInfo` struct, which can be embedded in your API response to provide pagination metadata.

```go
package main

import (
	"encoding/json"
	"net/http"

	"github.com/dmitrymomot/gokit/pagination"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Example API response structure
type ItemsResponse struct {
	Data       []Item            `json:"data"`
	Pagination pagination.PageInfo `json:"pagination"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	page := pagination.GetPageFromRequest(r)
	size := pagination.GetSizeFromRequest(r)
	offset := pagination.CalculateOffset(page, size)

	// Assume these come from your database query
	items := []Item{{ID: 1, Name: "Item 1"}, {ID: 2, Name: "Item 2"}}
	totalItems := 105 // Total count from the database (without limit/offset)

	// Create pagination info
	pageInfo := pagination.NewPageInfo(page, size, totalItems)

	// Construct the response
	response := ItemsResponse{
		Data:       items,
		Pagination: pageInfo,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

/*
Example JSON Response:
{
  "data": [
    { "id": 1, "name": "Item 1" },
    { "id": 2, "name": "Item 2" }
  ],
  "pagination": {
    "page": 1,
    "size": 10,
    "total_items": 105,
    "total_pages": 11
  }
}
```

## Cursor-Based Pagination (Planned)

*(Documentation and examples will be added here once implemented)*
