# Token Package

A lightweight, secure token generation and validation library with HMAC signatures.

## Installation

```bash
go get github.com/dmitrymomot/gokit/token
```

## Overview

The `token` package provides a simple way to create and validate secure tokens with HMAC-SHA256 signatures. It's designed for applications that need a lightweight token solution without the complexity of JWT.

## Features

- Type-safe payload handling with Go generics
- HMAC-SHA256 signatures for security
- Base64URL encoding for URL-safe tokens
- Simple, intuitive API with minimal code
- Zero external dependencies
- Fast encoding and decoding

## Usage

### Generating Tokens

```go
import "github.com/dmitrymomot/gokit/token"

// Define your payload structure
type UserPayload struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
    Role  string `json:"role"`
}

// Create a payload
payload := UserPayload{
    ID:    123,
    Email: "user@example.com",
    Role:  "admin",
}

// Your secret key for signing tokens
secret := "your-secret-key"

// Generate a token
tokenStr, err := token.GenerateToken(payload, secret)
if err != nil {
    // Handle error
}

fmt.Printf("Generated token: %s\n", tokenStr)
```

### Validating Tokens

```go
// Parse and validate the token
var parsedPayload UserPayload
parsedPayload, err = token.ParseToken[UserPayload](tokenStr, secret)
if err != nil {
    if err == token.ErrInvalidToken {
        // Handle invalid token format
    } else if err == token.ErrSignatureInvalid {
        // Handle invalid signature
    } else {
        // Handle other errors
    }
    return
}

// Use the validated payload
fmt.Printf("User ID: %d, Email: %s, Role: %s\n", 
    parsedPayload.ID, parsedPayload.Email, parsedPayload.Role)
```

### With Expiration Time

```go
// Add expiration time to your payload
type TokenWithExpiry struct {
    UserID int       `json:"user_id"`
    Exp    time.Time `json:"exp"`
}

// Create token with expiration
expToken := TokenWithExpiry{
    UserID: 123,
    Exp:    time.Now().Add(24 * time.Hour), // Expires in 24 hours
}

token, _ := token.GenerateToken(expToken, secret)

// When validating, check expiration
var parsed TokenWithExpiry
parsed, err = token.ParseToken[TokenWithExpiry](token, secret)
if err != nil {
    // Handle token validation error
}

if time.Now().After(parsed.Exp) {
    // Token has expired
}
```

## API Reference

### Functions

```go
// GenerateToken encodes and signs a payload into a token
func GenerateToken[T any](payload T, secret string) (string, error)

// ParseToken validates and extracts the payload from a token
func ParseToken[T any](token string, secret string) (T, error)
```

### Errors

```go
// ErrInvalidToken indicates the token format is invalid
var ErrInvalidToken = errors.New("invalid token format")

// ErrSignatureInvalid indicates the token signature doesn't match
var ErrSignatureInvalid = errors.New("invalid token signature")
```

## Implementation Details

Tokens are generated in the format `payload.signature` where:

1. `payload` is the Base64URL-encoded JSON representation of your data
2. `signature` is the truncated (8 bytes) HMAC-SHA256 signature of the payload, also Base64URL-encoded

This creates compact tokens that can be safely used in URLs and cookies.

## Security Best Practices

1. **Secret Management**:
   - Store secrets securely, not in code or version control
   - Use environment variables or a secret management service
   - Use different secrets for different environments

2. **Token Design**:
   - Include expiration time in your payload
   - Consider adding a unique token ID for revocation capability
   - Include only necessary data in the payload

3. **Application Security**:
   - Validate all tokens before trusting their contents
   - Use HTTPS when transmitting tokens
   - Consider token rotation for long-lived sessions

4. **Error Handling**:
   - Use specific error checks rather than string comparisons
   - Don't reveal details about token validation failures to clients
