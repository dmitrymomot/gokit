package jwt

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// ContextKey is a type for context keys used by the jwt middleware
type ContextKey string

// TokenExtractorFunc defines a function that extracts a token from an HTTP request
type TokenExtractorFunc func(r *http.Request) (string, error)

// SkipFunc defines a function that determines whether to skip the middleware
type SkipFunc func(r *http.Request) bool

// MiddlewareConfig contains configuration for the JWT middleware
type MiddlewareConfig struct {
	// Service is the JWT service to use for parsing tokens
	Service *Service
	
	// ContextKey is the key to use for storing claims in the request context
	ContextKey ContextKey
	
	// Extractor is a function that extracts a token from an HTTP request
	// If not specified, DefaultTokenExtractor is used
	Extractor TokenExtractorFunc
	
	// Skip is a function that determines whether to skip the middleware
	// If not specified, the middleware is never skipped
	Skip SkipFunc
}

// DefaultTokenExtractor extracts a JWT token from the Authorization header
// It expects the format "Bearer <token>"
func DefaultTokenExtractor(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrInvalidToken
	}
	
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidToken
	}
	
	return parts[1], nil
}

// Middleware returns a new JWT middleware handler
func Middleware(config MiddlewareConfig) func(next http.Handler) http.Handler {
	// Use default token extractor if none is provided
	if config.Extractor == nil {
		config.Extractor = DefaultTokenExtractor
	}
	
	// Return the middleware handler
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if we should skip the middleware
			if config.Skip != nil && config.Skip(r) {
				next.ServeHTTP(w, r)
				return
			}
			
			// Extract the token from the request
			tokenString, err := config.Extractor(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			
			// Parse the token
			claims := make(map[string]interface{})
			if err := config.Service.Parse(tokenString, &claims); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			
			// Add the claims to the request context
			ctx := context.WithValue(r.Context(), config.ContextKey, claims)
			
			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts JWT claims from the request context
func GetClaims(ctx context.Context, key ContextKey) (map[string]interface{}, bool) {
	claims, ok := ctx.Value(key).(map[string]interface{})
	return claims, ok
}

// GetClaimsAs extracts JWT claims from the request context and parses them into a struct
func GetClaimsAs(ctx context.Context, key ContextKey, dest interface{}) error {
	claims, ok := ctx.Value(key).(map[string]interface{})
	if !ok {
		return ErrInvalidClaims
	}
	
	// Convert claims to JSON and then unmarshal into the destination struct
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	
	if err := json.Unmarshal(claimsJSON, dest); err != nil {
		return err
	}
	
	return nil
}

// WithClaims is a helper function to create a middleware with a specific claims type
func WithClaims[T any](service *Service, key ContextKey) func(next http.Handler) http.Handler {
	return Middleware(MiddlewareConfig{
		Service:    service,
		ContextKey: key,
	})
}

// WithClaimsAndExtractor is a helper function to create a middleware with a specific 
// claims type and token extractor
func WithClaimsAndExtractor[T any](
	service *Service,
	key ContextKey,
	extractor TokenExtractorFunc,
) func(next http.Handler) http.Handler {
	return Middleware(MiddlewareConfig{
		Service:    service,
		ContextKey: key,
		Extractor:  extractor,
	})
}

// WithClaimsAndSkip is a helper function to create a middleware with a specific
// claims type and skip function
func WithClaimsAndSkip[T any](
	service *Service,
	key ContextKey,
	skip SkipFunc,
) func(next http.Handler) http.Handler {
	return Middleware(MiddlewareConfig{
		Service:    service,
		ContextKey: key,
		Skip:       skip,
	})
}

// CookieTokenExtractor creates a token extractor that extracts tokens from a cookie
func CookieTokenExtractor(cookieName string) TokenExtractorFunc {
	return func(r *http.Request) (string, error) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return "", ErrInvalidToken
		}
		return cookie.Value, nil
	}
}

// QueryTokenExtractor creates a token extractor that extracts tokens from a query parameter
func QueryTokenExtractor(paramName string) TokenExtractorFunc {
	return func(r *http.Request) (string, error) {
		token := r.URL.Query().Get(paramName)
		if token == "" {
			return "", ErrInvalidToken
		}
		return token, nil
	}
}

// HeaderTokenExtractor creates a token extractor that extracts tokens from a header
func HeaderTokenExtractor(headerName string) TokenExtractorFunc {
	return func(r *http.Request) (string, error) {
		token := r.Header.Get(headerName)
		if token == "" {
			return "", ErrInvalidToken
		}
		return token, nil
	}
}
