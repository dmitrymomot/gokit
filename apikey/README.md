# API Key Package

A secure API key generation, hashing, and validation package for authentication systems.

## Installation

```bash
go get github.com/dmitrymomot/gokit/apikey
```

## Overview

The `apikey` package provides tools for creating and managing secure API keys for authentication systems. It offers cryptographically secure key generation with multiple methods, secure hashing for storage, and constant-time validation to prevent timing attacks. This package is thread-safe and suitable for concurrent use.

## Features

- Cryptographically secure API key generation
- Time-ordered keys using UUID V7 for chronological sorting
- Secure hashing with HMAC-SHA256
- Constant-time comparison to prevent timing attacks
- Simple validation against stored hashes
- Thread-safe implementation for concurrent usage
- Comprehensive error handling

## Usage

### Generating API Keys

```go
import (
    "errors"
    "github.com/dmitrymomot/gokit/apikey"
)

// Generate a random API key (256 bits, hex-encoded)
apiKey, err := apikey.GenerateRandom()
if err != nil {
    // Handle error
    // Returns 64-character hexadecimal string or error
}
// apiKey = "a1b2c3d4..." (64-character hexadecimal string)

// Alternative: Generate a time-ordered API key using UUID V7
// Useful for keys that need to be sortable by creation time
orderedKey, err := apikey.GenerateTimeOrdered()
if err != nil {
    // Handle error
}
// orderedKey = "0188f8e8-..." (UUID v7 format)
```

### Hashing for Storage

```go
secretKey := "your-secret-key" // Store this securely

// Hash an API key for storage
hash, err := apikey.HashKey(apiKey, secretKey)
if err != nil {
    // Handle error
}
// hash = "a1b2c3..." (hex-encoded hash string)

// Store the hash in your database, not the API key itself
```

### Validating API Keys

```go
// When a client sends an API key, validate it against the stored hash
isValid := apikey.ValidateKey(apiKey, storedHash, secretKey)
if isValid {
    // Allow access
} else {
    // Deny access
}
```

### Secure String Comparison

```go
// Constant-time string comparison to prevent timing attacks
equal := apikey.SecureCompare(string1, string2)
// Returns true if strings are equal, false otherwise
```

### Error Handling

```go
apiKey, err := apikey.GenerateRandom()
if err != nil {
    switch {
    case errors.Is(err, apikey.ErrGeneration):
        // Handle generation error
    case errors.Is(err, apikey.ErrEmptyInput):
        // Handle empty input error (unlikely in this case)
    default:
        // Handle other errors
    }
}
```

## Best Practices

1. **Security**:
   - Never store raw API keys - only store hashed versions
   - Use a strong secret key - your secret key should be long, random, and kept secure
   - Rotate API keys periodically for sensitive systems

2. **Performance**:
   - Cache validation results for frequently used keys
   - For high-traffic systems, consider implementing a key cache

3. **Error Handling**:
   - Always check for errors when generating keys
   - Implement appropriate logging for failed validation attempts

## API Reference

### Constants

```go
const APIKeyLength = 32 // 256 bits - length of random API keys in bytes
```

### Functions

```go
func GenerateRandom() (string, error)
```
Creates a new API key with a secure random value. Returns a hex-encoded string of 64 characters (32 bytes) or an error if generation fails.

```go
func GenerateTimeOrdered() (string, error)
```
Creates a time-ordered API key using UUID V7 format. Returns the encoded API key as a string or an error if generation or encoding fails.

```go
func HashKey(apiKey, secretKey string) (string, error)
```
Hashes the API key using HMAC-SHA256 with a secret key. Both apiKey and secretKey must be non-empty strings. Returns the hex-encoded hash string or an error if inputs are invalid.

```go
func ValidateKey(apiKey, hash, secretKey string) bool
```
Checks if the API key matches the hash using the secret key. Returns true if the API key matches the hash, false otherwise.

```go
func SecureCompare(a, b string) bool
```
Performs a constant-time comparison of two strings to prevent timing attacks. Returns true if the strings are equal, false otherwise.

### Error Types

```go
var ErrEmptyInput = errors.New("empty api key or secret key")
var ErrGeneration = errors.New("failed to generate api key")
var ErrInvalidHash = errors.New("invalid hash format")
