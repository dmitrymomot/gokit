package pagination

import (
	"context"
	"net/http"
)

// paginatorContextKey is the context key type for storing paginator in request context
type paginatorContextKey struct{}

// Middleware creates a new middleware handler that extracts pagination parameters
// from the request and stores a Paginator in the request context.
//
// The config parameter allows for custom pagination configuration.
// If config is nil, the default configuration will be used.
func Middleware(config *PaginatorConfig) func(http.Handler) http.Handler {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Create a basic paginator with zero total items
			// The total items will be updated later once the data is fetched
			paginator := FromRequest(r, 0, config)

			// Store paginator in context
			ctx = context.WithValue(ctx, paginatorContextKey{}, paginator)
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetPaginatorFromContext retrieves the Paginator from the request context.
// Returns the zero value of Paginator if not found.
func GetPaginatorFromContext(ctx context.Context) Paginator {
	p, ok := ctx.Value(paginatorContextKey{}).(Paginator)
	if !ok {
		return Paginator{}
	}
	return p
}
