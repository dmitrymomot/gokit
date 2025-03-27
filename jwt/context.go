package jwt

import (
	"context"
	"encoding/json"
	"fmt"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey struct{ name string }

// String returns the name of the context key.
func (c contextKey) String() string { return c.name }

var (
	jwtContextKey    = &contextKey{name: "jwt"}        // JWT string
	claimsContextKey = &contextKey{name: "jwt_claims"} // Parsed JWT claims
)

// SetToken sets the JWT token string in the context.
func SetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, jwtContextKey, token)
}

// SetClaims sets the JWT claims in the context.
func SetClaims(ctx context.Context, claims map[string]any) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// GetToken returns the JWT token string from the context.
// If no token is found, the second return value will be false.
func GetToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(jwtContextKey).(string)
	return token, ok
}

// GetClaims returns the JWT claims as a map[string]any from the context.
// If no claims are found, the second return value will be false.
func GetClaims(ctx context.Context) (map[string]any, bool) {
	claims, ok := ctx.Value(claimsContextKey).(map[string]any)
	return claims, ok
}

// GetClaimsAs parses the JWT claims from the context into the specified struct.
// If no claims are found or they cannot be parsed, an error is returned.
func GetClaimsAs[T any](ctx context.Context, claims *T) error {
	claimsMap, ok := ctx.Value(claimsContextKey).(map[string]any)
	if !ok {
		return ErrInvalidClaims
	}

	// Convert the map to JSON and then unmarshal to the provided struct
	claimsJSON, err := json.Marshal(claimsMap)
	if err != nil {
		return fmt.Errorf("failed to marshal claims: %w", err)
	}

	if err := json.Unmarshal(claimsJSON, claims); err != nil {
		return fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return nil
}
