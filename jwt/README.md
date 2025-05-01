# JWT Package

A simple, high-performance JWT (JSON Web Token) implementation for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/jwt
```

## Overview

The `jwt` package provides a minimalist JWT implementation focused on type safety, performance, and security. It supports token generation, validation, and HTTP middleware integration without external dependencies. The package is thread-safe and suitable for concurrent use in production applications.

## Features

- Generate and parse JWT tokens with standard or custom claims
- Type-safe claims with proper validation
- HTTP middleware with flexible token extraction
- Support for token expiration and custom claims validation
- Minimal dependencies with optimized performance
- HMAC-SHA256 (HS256) signing method
- Thread-safe implementation for concurrent usage

## Usage

### Basic Token Generation and Parsing

```go
import (
    "fmt"
    "github.com/dmitrymomot/gokit/jwt"
    "time"
)

// Create a JWT service
jwtService, err := jwt.New([]byte("your-secret-key"))
if err != nil {
    // Handle error
    panic(fmt.Sprintf("Failed to create JWT service: %v", err))
}

// Create standard claims
claims := jwt.StandardClaims{
    Subject:   "user123",
    Issuer:    "myapp",
    ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    IssuedAt:  time.Now().Unix(),
}

// Generate a token
token, err := jwtService.Generate(claims)
if err != nil {
    // Handle error
    fmt.Printf("Failed to generate token: %v\n", err)
    return
}
// Output: token is a string like "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiaXNzIjoibXlhcHAiLCJleHAiOjE2NTQ0NzI4MDAsImlhdCI6MTY1NDM4NjQwMH0.8Uj7PoJuDdnGoDei5XH6b7YjLdkDZ6Gv2eUDbAyRuYM"

// Parse the token
var parsedClaims jwt.StandardClaims
err = jwtService.Parse(token, &parsedClaims)
if err != nil {
    // Handle error
    fmt.Printf("Failed to parse token: %v\n", err)
    return
}
// parsedClaims now contains: {Subject:"user123", Issuer:"myapp", ExpiresAt:1654472800, IssuedAt:1654386400}

// Access individual claims
fmt.Println("User ID:", parsedClaims.Subject)
// Output: User ID: user123
fmt.Println("Token expires at:", time.Unix(parsedClaims.ExpiresAt, 0))
// Output: Token expires at: 2022-06-06 00:00:00 +0000 UTC
```

### Custom Claims

```go
// Define custom claims
type UserClaims struct {
    jwt.StandardClaims
    Name  string   `json:"name,omitempty"`
    Email string   `json:"email,omitempty"`
    Roles []string `json:"roles,omitempty"`
}

// Create custom claims
claims := UserClaims{
    StandardClaims: jwt.StandardClaims{
        Subject:   "user123",
        ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
    },
    Name:  "John Doe",
    Email: "john@example.com",
    Roles: []string{"admin", "user"},
}

// Generate token with custom claims
token, err := jwtService.Generate(claims)
if err != nil {
    fmt.Printf("Failed to generate token: %v\n", err)
    return
}
// Output: token contains all the custom claims encoded in JWT format

// Parse token with custom claims
var parsedClaims UserClaims
err = jwtService.Parse(token, &parsedClaims)
if err != nil {
    fmt.Printf("Failed to parse token: %v\n", err)
    return
}

// Access custom claims
fmt.Println("User:", parsedClaims.Name)
// Output: User: John Doe
fmt.Println("Roles:", parsedClaims.Roles)
// Output: Roles: [admin user]
```

### Error Handling

```go
import (
    "errors"
    "fmt"
    "github.com/dmitrymomot/gokit/jwt"
    "net/http"
    "time"
)

// Example 1: Handling expired tokens
expiredClaims := jwt.StandardClaims{
    Subject:   "user123",
    ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
}

expiredToken, _ := jwtService.Generate(expiredClaims)
var parsedClaims jwt.StandardClaims

err := jwtService.Parse(expiredToken, &parsedClaims)
if err != nil {
    switch {
    case errors.Is(err, jwt.ErrExpiredToken):
        // Token has expired
        fmt.Println("Please log in again, your session has expired")
        // Output: Please log in again, your session has expired
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}

// Example 2: Handling tampered tokens
tamperedToken := expiredToken + "tampered"
err = jwtService.Parse(tamperedToken, &parsedClaims)
if err != nil {
    switch {
    case errors.Is(err, jwt.ErrInvalidSignature):
        // Token signature is invalid (token was tampered with)
        fmt.Println("Security alert: Invalid token signature")
        // Output: Security alert: Invalid token signature
    case errors.Is(err, jwt.ErrInvalidToken):
        // Malformed token
        fmt.Println("Invalid token format")
        // Output: Invalid token format
    default:
        fmt.Printf("Unknown error: %v\n", err)
    }
}

// Example 3: Comprehensive error handling function
func validateUserToken(tokenString string) (*UserClaims, error) {
    var claims UserClaims
    err := jwtService.Parse(tokenString, &claims)
    if err != nil {
        switch {
        case errors.Is(err, jwt.ErrExpiredToken):
            return nil, fmt.Errorf("session expired: %w", err)
        case errors.Is(err, jwt.ErrInvalidSignature):
            return nil, fmt.Errorf("invalid token signature: %w", err)
        case errors.Is(err, jwt.ErrInvalidToken):
            return nil, fmt.Errorf("malformed token: %w", err)
        default:
            return nil, fmt.Errorf("token validation failed: %w", err)
        }
    }
    
    // Additional validation if needed
    if len(claims.Roles) == 0 {
        return nil, errors.New("token has no roles assigned")
    }
    
    return &claims, nil
}
```

### HTTP Middleware

```go
import (
    "net/http"
    "github.com/dmitrymomot/gokit/jwt"
)

const JWTContextKey jwt.ContextKey = "user_claims"

// Create JWT middleware
jwtMiddleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
})

// Create a protected handler
protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Get claims from context
    claims, ok := jwt.GetClaims(r.Context(), JWTContextKey)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Use the claims
    username, _ := claims["sub"].(string)
    w.Write([]byte("Hello, " + username))
    // Output: "Hello, user123" (if token's subject was "user123")
})

// Apply middleware
http.Handle("/protected", jwtMiddleware(protectedHandler))
```

### Type-Safe Claims in Handlers

```go
// Define your claims type
type UserClaims struct {
    jwt.StandardClaims
    Role string `json:"role"`
}

// In your handler
func handler(w http.ResponseWriter, r *http.Request) {
    var userClaims UserClaims
    if err := jwt.GetClaimsAs(r.Context(), JWTContextKey, &userClaims); err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    
    // Now you have strongly typed claims
    if userClaims.Role != "admin" {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    w.Write([]byte("Welcome, admin!"))
    // Output: "Welcome, admin!" (if token's role was "admin")
}
```

### Custom Token Extraction

```go
// From a cookie
middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.CookieTokenExtractor("auth_token"),
})
// Extracts token from the "auth_token" cookie

// From a query parameter
middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.QueryTokenExtractor("token"),
})
// Extracts token from the "token" query parameter (e.g., ?token=xyz)

// From a custom header
middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.HeaderTokenExtractor("X-API-Token"),
})
// Extracts token from the "X-API-Token" HTTP header
```

### Skip Middleware for Public Routes

```go
middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Skip: func(r *http.Request) bool {
        // Skip auth for public endpoints
        return r.URL.Path == "/api/public" || r.URL.Path == "/health"
    },
})
// The middleware will not check for tokens on /api/public or /health paths
```

### Helper Functions

```go
// Type-safe middleware with generics
middleware := jwt.WithClaims[UserClaims](jwtService, JWTContextKey)
// Creates middleware that strongly types the claims as UserClaims

// With custom extractor
middleware := jwt.WithClaimsAndExtractor[UserClaims](
    jwtService, 
    JWTContextKey,
    jwt.CookieTokenExtractor("auth"),
)
// Creates middleware with typed claims and custom token extraction
```

## Best Practices

1. **Security**:
   - Use strong, secret keys (at least 32 bytes) for signing tokens
   - Set appropriate expiration times on tokens
   - Regularly rotate signing keys for long-lived applications
   - Validate all claims before trusting token content

2. **Token Management**:
   - Keep tokens as short-lived as possible
   - Implement token refresh mechanisms for longer sessions
   - Store tokens securely on the client (HttpOnly cookies for web apps)
   - Implement token revocation for sensitive applications

3. **Error Handling**:
   - Always check for specific error types when parsing tokens
   - Return appropriate HTTP status codes (401 for expired/invalid tokens)
   - Log suspicious activity like invalid signatures (possible tampering)
   - Provide user-friendly messages without exposing internal details

4. **Performance**:
   - Cache parsed tokens in high-traffic scenarios
   - Use the WithClaims generic helper for type safety without runtime overhead
   - Keep claims minimal - tokens are passed with every request

## API Reference

### Types

```go
type Service interface {
    Generate(claims interface{}) (string, error)
    Parse(tokenString string, claims interface{}) error
}
```
Interface for the JWT service with methods for generating and parsing tokens.

```go
type StandardClaims struct {
    Audience  string `json:"aud,omitempty"`
    ExpiresAt int64  `json:"exp,omitempty"`
    ID        string `json:"jti,omitempty"`
    IssuedAt  int64  `json:"iat,omitempty"`
    Issuer    string `json:"iss,omitempty"`
    NotBefore int64  `json:"nbf,omitempty"`
    Subject   string `json:"sub,omitempty"`
}
```
Standard claims structure as per JWT specification.

```go
type ContextKey string
```
Type for JWT context keys.

```go
type MiddlewareConfig struct {
    Service    Service
    ContextKey ContextKey
    Extractor  TokenExtractor
    Skip       func(*http.Request) bool
}
```
Configuration for JWT middleware.

```go
type TokenExtractor func(*http.Request) string
```
Function type for extracting JWT tokens from HTTP requests.

### Functions

```go
func New(signingKey []byte) (Service, error)
```
Creates a new JWT service with the given signing key.

```go
func Middleware(config MiddlewareConfig) func(http.Handler) http.Handler
```
Creates HTTP middleware for JWT authentication.

```go
func WithClaims[T any](service Service, contextKey ContextKey) func(http.Handler) http.Handler
```
Creates type-safe JWT middleware with generics.

```go
func WithClaimsAndExtractor[T any](service Service, contextKey ContextKey, extractor TokenExtractor) func(http.Handler) http.Handler
```
Creates type-safe JWT middleware with custom token extractor.

```go
func GetClaims(ctx context.Context, key ContextKey) (map[string]interface{}, bool)
```
Gets claims from context as a map.

```go
func GetClaimsAs(ctx context.Context, key ContextKey, claims interface{}) error
```
Gets claims from context as a strongly typed structure.

```go
func HeaderTokenExtractor(header string) TokenExtractor
```
Creates a token extractor that gets tokens from an HTTP header.

```go
func AuthHeaderTokenExtractor() TokenExtractor
```
Creates a token extractor that gets tokens from the Authorization header.

```go
func CookieTokenExtractor(cookieName string) TokenExtractor
```
Creates a token extractor that gets tokens from an HTTP cookie.

```go
func QueryTokenExtractor(paramName string) TokenExtractor
```
Creates a token extractor that gets tokens from a query parameter.

### Error Types

```go
var ErrInvalidToken = errors.New("invalid token")
var ErrInvalidSignature = errors.New("invalid token signature")
var ErrExpiredToken = errors.New("token has expired")
var ErrInvalidClaims = errors.New("invalid claims")
var ErrMissingToken = errors.New("missing token")
