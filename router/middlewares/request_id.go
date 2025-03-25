package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is the name of the header that contains the request ID
const RequestIDHeader = "X-Request-Id"

// requestIDKey is the context key for the request ID
var requestIDKey = newContextKey("request_id")

// GetRequestID returns the request ID from the context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// WithRequestID returns a new context with the provided request ID
func WithRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, reqID)
}

// RequestID is a middleware that injects a request ID into the context of each
// request. A request ID is a randomly generated UUID. If the incoming request has
// a X-Request-Id header, then that value is used instead.
// The request ID is returned as a response header for easy debugging.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from header or generate a new one
		reqID := r.Header.Get(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		// Set the request ID in the response header
		w.Header().Set(RequestIDHeader, reqID)

		// Add the request ID to the context
		ctx := WithRequestID(r.Context(), reqID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
