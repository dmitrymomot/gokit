# JWT

A simple, high-performance JWT (JSON Web Token) package for Go applications.

## Overview

This package provides a minimalist implementation of JWT with focus on:
- Type safety: Strongly typed interfaces with proper error handling
- Performance: Optimized for speed with minimal allocations
- Simplicity: Clean API with just what you need
- Security: Uses HMAC-SHA256 (HS256) signing method

The package implements token generation and parsing with customizable claims and doesn't rely on any external JWT libraries.

## Features

- Generate JWT tokens with standard or custom claims
- Parse and validate JWT tokens
- Support for expiration validation
- Custom claims validation
- Simple service-based API

## Installation

```bash
go get github.com/dmitrymomot/gokit/jwt
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "time"

    "github.com/dmitrymomot/gokit/jwt"
)

func main() {
    // Create a new JWT service with a signing key
    jwtService, err := jwt.New([]byte("your-secret-key"))
    if err != nil {
        panic(err)
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
        panic(err)
    }
    fmt.Println("Token:", token)

    // Parse the token
    var parsedClaims jwt.StandardClaims
    err = jwtService.Parse(token, &parsedClaims)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Parsed claims: %+v\n", parsedClaims)
}
```

### Custom Claims

```go
package main

import (
    "fmt"
    "time"

    "github.com/dmitrymomot/gokit/jwt"
)

// Define custom claims
type UserClaims struct {
    jwt.StandardClaims
    Name  string   `json:"name,omitempty"`
    Email string   `json:"email,omitempty"`
    Roles []string `json:"roles,omitempty"`
}

func main() {
    // Create a new JWT service from string
    jwtService, err := jwt.NewFromString("your-secret-key")
    if err != nil {
        panic(err)
    }

    // Create custom claims
    claims := UserClaims{
        StandardClaims: jwt.StandardClaims{
            Subject:   "user123",
            Issuer:    "myapp",
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
        },
        Name:  "John Doe",
        Email: "john@example.com",
        Roles: []string{"admin", "user"},
    }

    // Generate a token
    token, err := jwtService.Generate(claims)
    if err != nil {
        panic(err)
    }

    // Parse the token
    var parsedClaims UserClaims
    err = jwtService.Parse(token, &parsedClaims)
    if err != nil {
        panic(err)
    }
    fmt.Printf("User: %s (%s)\n", parsedClaims.Name, parsedClaims.Email)
    fmt.Printf("Roles: %v\n", parsedClaims.Roles)
}
```

### Error Handling

```go
package main

import (
    "errors"
    "fmt"
    "time"

    "github.com/dmitrymomot/gokit/jwt"
)

func main() {
    jwtService, _ := jwt.New([]byte("secret-key"))

    // Generate an expired token
    expiredClaims := jwt.StandardClaims{
        Subject:   "user123",
        ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
    }
    expiredToken, _ := jwtService.Generate(expiredClaims)

    // Try to parse the expired token
    var claims jwt.StandardClaims
    err := jwtService.Parse(expiredToken, &claims)
    
    if err != nil {
        if errors.Is(err, jwt.ErrExpiredToken) {
            fmt.Println("The token has expired")
        } else if errors.Is(err, jwt.ErrInvalidSignature) {
            fmt.Println("The token signature is invalid")
        } else {
            fmt.Printf("Error: %v\n", err)
        }
    }
}
```

## API Reference

### Service Creation

- `New(signingKey []byte) (*Service, error)` - Create a new JWT service with a byte slice signing key
- `NewFromString(signingKey string) (*Service, error)` - Create a new JWT service with a string signing key

### Token Management

- `Generate(claims any) (string, error)` - Generate a JWT token with provided claims
- `Parse(tokenString string, claims any) error` - Parse a JWT token into the provided claims struct

### Standard Claims

The `StandardClaims` struct implements common JWT claims:

- `ID` - JWT ID (jti)
- `Subject` - Subject (sub)
- `Issuer` - Issuer (iss)
- `Audience` - Audience (aud)
- `ExpiresAt` - Expiration time (exp)
- `NotBefore` - Not before time (nbf)
- `IssuedAt` - Issued at time (iat)

### Errors

The package defines multiple error types for different failure scenarios:

- `ErrInvalidToken` - Token format is invalid
- `ErrExpiredToken` - Token has expired
- `ErrInvalidSigningMethod` - Signing method is invalid
- `ErrMissingSigningKey` - Signing key is missing
- `ErrInvalidSigningKey` - Signing key is invalid
- `ErrInvalidClaims` - Claims are invalid
- `ErrMissingClaims` - Claims are missing
- `ErrInvalidSignature` - Signature is invalid
- `ErrUnexpectedSigningMethod` - Unexpected signing method

### Middleware

The package includes HTTP middleware for JWT authentication with the following features:

- Automatically extract JWT tokens from requests (customizable)
- Parse tokens and validate signatures
- Add claims to the request context
- Skip middleware based on custom conditions
- Helper functions for common use cases

#### Middleware Configuration

The `MiddlewareConfig` struct allows you to configure the middleware:

```go
type MiddlewareConfig struct {
    // Service is the JWT service to use for parsing tokens
    Service *Service
    
    // ContextKey is the key to use for storing claims in the request context
    ContextKey ContextKey
    
    // Extractor is a function that extracts a token from an HTTP request
    // If not specified, DefaultTokenExtractor is used (Authorization: Bearer <token>)
    Extractor TokenExtractorFunc
    
    // Skip is a function that determines whether to skip the middleware
    // If not specified, the middleware is never skipped
    Skip SkipFunc
}
```

#### Basic Usage

```go
package main

import (
    "log"
    "net/http"

    "github.com/dmitrymomot/gokit/jwt"
)

const (
    JWTContextKey jwt.ContextKey = "user_claims"
)

func main() {
    // Create a JWT service
    jwtService, err := jwt.New([]byte("your-secret-key"))
    if err != nil {
        log.Fatal(err)
    }

    // Create middleware with default configuration
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

    // Apply middleware to the handler
    http.Handle("/protected", jwtMiddleware(protectedHandler))
    http.ListenAndServe(":8080", nil)
}
```

#### Custom Token Extraction

```go
// Create middleware that extracts tokens from a cookie
cookieMiddleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.CookieTokenExtractor("auth_token"),
})

// Create middleware that extracts tokens from a query parameter
queryMiddleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.QueryTokenExtractor("token"),
})

// Create middleware that extracts tokens from a custom header
headerMiddleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor:  jwt.HeaderTokenExtractor("X-API-Token"),
})

// Create middleware with a completely custom extractor
customMiddleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Extractor: func(r *http.Request) (string, error) {
        // Extract token from wherever you want
        return r.Header.Get("X-Custom-Token"), nil
    },
})
```

#### Skip Middleware Conditionally

```go
// Create middleware that skips public endpoints
publicPaths := map[string]bool{
    "/api/public": true,
    "/health":     true,
    "/metrics":    true,
}

middleware := jwt.Middleware(jwt.MiddlewareConfig{
    Service:    jwtService,
    ContextKey: JWTContextKey,
    Skip: func(r *http.Request) bool {
        return publicPaths[r.URL.Path]
    },
})
```

#### Type-Safe Claims Access

```go
// Define custom claims type
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
    
    // Use typed claims
    if userClaims.Role != "admin" {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    
    // Process the request
    // ...
}
```

#### Helper Functions

The package provides helper functions for common middleware configurations:

```go
// Basic middleware with generic type parameter
middleware := jwt.WithClaims[UserClaims](jwtService, JWTContextKey)

// Middleware with custom extractor
middleware := jwt.WithClaimsAndExtractor[UserClaims](
    jwtService, 
    JWTContextKey,
    jwt.CookieTokenExtractor("auth"),
)

// Middleware with skip function
middleware := jwt.WithClaimsAndSkip[UserClaims](
    jwtService,
    JWTContextKey,
    func(r *http.Request) bool { return r.URL.Path == "/public" },
)
