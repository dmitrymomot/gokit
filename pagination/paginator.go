package pagination

import (
	"encoding/json"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

// Paginator provides an immutable interface for pagination operations.
// It contains all necessary pagination data and methods to interact with it.
type Paginator struct {
	// Core pagination state (immutable)
	currentPage  int
	itemsPerPage int
	totalItems   int
	totalPages   int

	// Configuration reference (immutable)
	config *PaginatorConfig
}

// calculateOffset calculates the database offset based on page number and size.
func calculateOffset(page, size int) int {
	if page < 1 {
		page = DefaultPage
	}
	if size < 1 {
		size = DefaultLimit
	}
	return (page - 1) * size
}

// calculateTotalPages calculates the total number of pages based on total items and page size.
func calculateTotalPages(totalItems, size int) int {
	if totalItems <= 0 || size <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(size)))
}

// Offset returns the calculated offset for database queries.
func (p Paginator) Offset() int {
	return calculateOffset(p.currentPage, p.itemsPerPage)
}

// Limit returns the number of items per page.
func (p Paginator) Limit() int {
	return p.itemsPerPage
}

// OffsetLimit returns both offset and limit in a single struct.
func (p Paginator) OffsetLimit() OffsetLimit {
	return OffsetLimit{
		Offset: p.Offset(),
		Limit:  p.Limit(),
	}
}

// CurrentPage returns the current page number.
func (p Paginator) CurrentPage() int {
	return p.currentPage
}

// ItemsPerPage returns the number of items per page.
func (p Paginator) ItemsPerPage() int {
	return p.itemsPerPage
}

// TotalItems returns the total number of items.
func (p Paginator) TotalItems() int {
	return p.totalItems
}

// TotalPages returns the total number of pages.
func (p Paginator) TotalPages() int {
	return p.totalPages
}

// HasNext returns true if there's a next page.
func (p Paginator) HasNext() bool {
	return p.currentPage < p.totalPages
}

// HasPrevious returns true if there's a previous page.
func (p Paginator) HasPrevious() bool {
	return p.currentPage > 1
}

// IsFirstPage returns true if this is the first page.
func (p Paginator) IsFirstPage() bool {
	return p.currentPage == 1
}

// IsLastPage returns true if this is the last page.
func (p Paginator) IsLastPage() bool {
	return p.currentPage == p.totalPages
}

// buildPageURL creates a URL for the specified page while preserving other query parameters.
func (p Paginator) buildPageURL(page int) string {
	if p.config.BaseURL == "" {
		return ""
	}

	// Copy existing query params
	params := url.Values{}
	for key, values := range p.config.QueryParams {
		for _, value := range values {
			params.Add(key, value)
		}
	}

	// Set the page and size params
	params.Set(p.config.PageParam, strconv.Itoa(page))
	params.Set(p.config.SizeParam, strconv.Itoa(p.itemsPerPage))

	// Build the URL
	u, err := url.Parse(p.config.BaseURL)
	if err != nil {
		return p.config.BaseURL
	}
	u.RawQuery = params.Encode()
	return u.String()
}

// NextPageURL returns the URL for the next page or empty string if none.
func (p Paginator) NextPageURL() string {
	if !p.HasNext() {
		return ""
	}
	return p.buildPageURL(p.currentPage + 1)
}

// PreviousPageURL returns the URL for the previous page or empty string if none.
func (p Paginator) PreviousPageURL() string {
	if !p.HasPrevious() {
		return ""
	}
	return p.buildPageURL(p.currentPage - 1)
}

// FirstPageURL returns the URL for the first page.
func (p Paginator) FirstPageURL() string {
	return p.buildPageURL(1)
}

// LastPageURL returns the URL for the last page.
func (p Paginator) LastPageURL() string {
	if p.totalPages < 1 {
		return p.FirstPageURL()
	}
	return p.buildPageURL(p.totalPages)
}

// Links returns a map of link relations to URLs.
func (p Paginator) Links() map[string]string {
	links := map[string]string{
		"self":  p.buildPageURL(p.currentPage),
		"first": p.FirstPageURL(),
	}

	if p.totalPages > 0 {
		links["last"] = p.LastPageURL()
	}

	if p.HasNext() {
		links["next"] = p.NextPageURL()
	}

	if p.HasPrevious() {
		links["prev"] = p.PreviousPageURL()
	}

	return links
}

// LinksArray returns an array of Link objects for structured API responses.
func (p Paginator) LinksArray() []Link {
	links := []Link{
		{Rel: "self", URL: p.buildPageURL(p.currentPage)},
		{Rel: "first", URL: p.FirstPageURL()},
	}

	if p.totalPages > 0 {
		links = append(links, Link{Rel: "last", URL: p.LastPageURL()})
	}

	if p.HasNext() {
		links = append(links, Link{Rel: "next", URL: p.NextPageURL()})
	}

	if p.HasPrevious() {
		links = append(links, Link{Rel: "prev", URL: p.PreviousPageURL()})
	}

	return links
}

// PageInfo returns a PageInfo struct for API responses.
func (p Paginator) PageInfo() PageInfo {
	return PageInfo{
		Page:       p.currentPage,
		Size:       p.itemsPerPage,
		TotalItems: p.totalItems,
		TotalPages: p.totalPages,
	}
}

// WithTotalItems creates a new Paginator with updated total items count.
// This is useful when you need to create the Paginator before knowing the total count.
// If totalItems is negative, it will be treated as 0.
func (p Paginator) WithTotalItems(totalItems int) Paginator {
	// Handle negative totalItems by using 0 instead
	if totalItems < 0 {
		totalItems = 0
	}

	// Calculate the new total pages
	totalPages := calculateTotalPages(totalItems, p.itemsPerPage)

	// Create a new paginator with updated values
	return Paginator{
		currentPage:  p.currentPage,
		itemsPerPage: p.itemsPerPage,
		totalItems:   totalItems,
		totalPages:   totalPages,
		config:       p.config,
	}
}

// RenderPageNumbers returns a slice of page numbers for template rendering
// with current page marked and configurable number of pages before/after.
func (p Paginator) RenderPageNumbers(before, after int) []PageNumber {
	if p.TotalPages() == 0 {
		return nil
	}

	result := make([]PageNumber, 0)

	// Calculate start and end page numbers
	start := p.currentPage - before
	if start < 1 {
		start = 1
	}

	end := p.currentPage + after
	if end > p.totalPages {
		end = p.totalPages
	}

	// Build page number objects
	for i := start; i <= end; i++ {
		result = append(result, PageNumber{
			Number:  i,
			URL:     p.buildPageURL(i),
			Current: i == p.currentPage,
		})
	}

	return result
}

// RespondWithJSON writes a standard paginated JSON response with the given items.
// This is a convenience method for REST APIs to write a consistent response format.
func (p Paginator) RespondWithJSON(w http.ResponseWriter, items any) error {
	response := map[string]any{
		"items":      items,
		"pagination": p.PageInfo(),
		"links":      p.Links(),
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}
