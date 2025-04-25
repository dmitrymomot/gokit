# JWT Package

A simple, high-performance JWT (JSON Web Token) implementation for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/jwt
```

## Overview

The `jwt` package provides a minimalist JWT implementation focused on type safety, performance, and security. It supports token generation, validation, and HTTP middleware integration without external dependencies.

## Features

- Generate and parse JWT tokens with standard or custom claims
- Type-safe claims with proper validation
- HTTP middleware with flexible token extraction
- Support for token expiration and custom claims validation
- Minimal dependencies with optimized performance
- HMAC-SHA256 (HS256) signing method

## Usage

### Basic Token Generation and Parsing

```go
import (
    "github.com/dmitrymomot/gokit/jwt"
    "time"
)

// Create a JWT service
jwtService, err := jwt.New([]byte("your-secret-key"))
if err != nil {
    // Handle error
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
}

// Parse the token
var parsedClaims jwt.StandardClaims
err = jwtService.Parse(token, &parsedClaims)
if err != nil {
    // Handle error
}
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

// Generate and parse as usual
token, err := jwtService.Generate(claims)
// ...
var parsedClaims UserClaims
err = jwtService.Parse(token, &parsedClaims)
```

### Error Handling

```go
err := jwtService.Parse(token, &claims)
if err != nil {
    switch {
    case errors.Is(err, jwt.ErrExpiredToken):
        // Token has expired
    case errors.Is(err, jwt.ErrInvalidSignature):
        // Invalid signature
    case errors.Is(err, jwt.ErrInvalidToken):
        // Malformed token
    default:
        // Handle other errors
    }
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

// From a query parameter
middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.QueryTokenExtractor("token"),
})

// From a custom header
middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.HeaderTokenExtractor("X-API-Token"),
})
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
```

### Helper Functions

```go
// Type-safe middleware with generics
middleware := jwt.WithClaims[UserClaims](jwtService, JWTContextKey)

// With custom extractor
middleware := jwt.WithClaimsAndExtractor[UserClaims](
    jwtService, 
    JWTContextKey,
    jwt.CookieTokenExtractor("auth"),
)
