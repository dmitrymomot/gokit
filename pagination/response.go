package pagination

import (
	"encoding/json"
	"net/http"
)

// PaginatedResponse represents a generic paginated response for REST APIs.
// It includes the items, pagination information, and navigation links.
type PaginatedResponse[T any] struct {
	// Items holds the paginated data.
	Items []T `json:"items"`

	// Pagination contains the pagination metadata.
	Pagination PageInfo `json:"pagination"`

	// Links contains navigation links for the paginated data.
	Links map[string]string `json:"links"`
}

// NewPaginatedResponse creates a new paginated response with the given items and paginator.
func NewPaginatedResponse[T any](items []T, paginator Paginator) PaginatedResponse[T] {
	return PaginatedResponse[T]{
		Items:      items,
		Pagination: paginator.PageInfo(),
		Links:      paginator.Links(),
	}
}

// RespondWithJSON is a helper function to write a paginated JSON response to an HTTP response writer.
func RespondWithJSON[T any](paginator Paginator, items []T, w http.ResponseWriter) error {
	response := NewPaginatedResponse(items, paginator)

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}
