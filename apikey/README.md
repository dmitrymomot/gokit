# API Key Package

A secure API key generation, hashing, and validation package for authentication systems.

## Installation

```bash
go get github.com/dmitrymomot/gokit/apikey
```

## Overview

The `apikey` package provides tools for creating and managing secure API keys for authentication systems. It offers cryptographically secure key generation with multiple methods, secure hashing for storage, and constant-time validation to prevent timing attacks.

## Features

- Cryptographically secure API key generation
- Time-ordered keys using UUID V7 for chronological sorting
- Secure hashing with HMAC-SHA256
- Constant-time comparison to prevent timing attacks
- Simple validation against stored hashes
- Comprehensive error handling

## Usage

### Generating API Keys

```go
// Generate a random API key (256 bits, hex-encoded)
apiKey, err := apikey.GenerateRandom()
if err != nil {
    // Handle error
}

// Alternative: Generate a time-ordered API key using UUID V7
// Useful for keys that need to be sortable by creation time
orderedKey, err := apikey.GenerateTimeOrdered()
if err != nil {
    // Handle error
}
```

### Hashing for Storage

```go
secretKey := "your-secret-key" // Store this securely

// Hash an API key for storage
hash, err := apikey.HashKey(apiKey, secretKey)
if err != nil {
    // Handle error
}

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
```

## Best Practices

1. **Never store raw API keys** - Only store hashed versions
2. **Use a strong secret key** - Your secret key should be long, random, and kept secure
3. **Rotate API keys periodically** - Implement key rotation for enhanced security
4. **Add rate limiting** - Protect your API from brute force attacks

## API Reference

### Functions

#### `GenerateRandom() (string, error)`

Creates a new API key with a secure random value. Returns a hex-encoded string of 64 characters.

#### `GenerateTimeOrdered() (string, error)`

Creates a new API key using a UUID V7. Generates a time-ordered key that can be chronologically sorted.

#### `HashKey(apiKey, secretKey string) (string, error)`

Hashes the API key using HMAC-SHA256 with a secret key.

#### `ValidateKey(apiKey, hash, secretKey string) bool`

Checks if the API key matches the hash using the secret key.

#### `SecureCompare(a, b string) bool`

Performs a constant-time comparison of two strings to prevent timing attacks.
