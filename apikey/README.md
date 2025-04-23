# API Key Package

This package provides API key generation, hashing, and validation functionality for building secure API authentication systems.

## Overview

The `apikey` package offers tools to:

1. Generate cryptographically secure API keys
2. Hash API keys with a secret key for secure storage
3. Validate API keys against stored hashes
4. Compare strings in constant time to prevent timing attacks

## Installation

```go
import "github.com/dmitrymomot/gokit/apikey"
```

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
apiKey, err := apikey.GenerateTimeOrdered()
if err != nil {
    // Handle error
}
```

### Hashing API Keys

Always hash API keys before storing them:

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
5. **Include expiration logic** - Consider adding expiration timestamps to your API keys

## Error Handling

The package provides the following error types:

- `ErrEmptyInput` - Returned when the API key or secret key is empty
- `ErrGeneration` - Returned when API key generation fails
- `ErrInvalidHash` - Returned when the hash format is invalid

## Key Generation Methods

The package provides two distinct methods for generating API keys, each designed for different use cases:

### GenerateRandom

`GenerateRandom()` creates a fully random API key using cryptographically secure random bytes. This method is ideal for maximum security and unpredictability.

Benefits:
- Maximum entropy and unpredictability
- Simple implementation with minimal dependencies
- Hex-encoded output (64 characters)

### GenerateTimeOrdered

`GenerateTimeOrdered()` creates an API key based on a UUID V7, which embeds a timestamp. This makes the keys sortable by creation time.

Benefits:
- Keys can be chronologically ordered
- Still maintains good entropy and security
