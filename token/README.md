# Token Package

A lightweight and secure token generation and validation package for Go applications. This package provides functionality to create and validate tokens with HMAC-SHA256 signatures.

## Features

- Generic type support for payload data
- HMAC-SHA256 signature for security
- Base64URL encoding for URL-safe tokens
- Simple and easy-to-use API
- Zero external dependencies

## Installation

```bash
go get github.com/dmitrymomot/gokit/token
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/dmitrymomot/gokit/token"
)

// Define your payload structure
type UserPayload struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}

func main() {
    // Your secret key for signing tokens
    secret := "your-secret-key"

    // Create a payload
    payload := UserPayload{
        ID:    123,
        Email: "user@example.com",
    }

    // Generate a token
    tokenStr, err := token.GenerateToken(payload, secret)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Generated token: %s\n", tokenStr)

    // Parse and validate the token
    var parsedPayload UserPayload
    parsedPayload, err = token.ParseToken[UserPayload](tokenStr, secret)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Parsed payload: %+v\n", parsedPayload)
}
```

### Error Handling

```go
// Handle invalid token format
_, err := token.ParseToken[UserPayload]("invalid-token", secret)
if err == token.ErrInvalidToken {
    fmt.Println("Invalid token format")
}

// Handle signature mismatch
_, err = token.ParseToken[UserPayload](validToken, "wrong-secret")
if err == token.ErrSignatureInvalid {
    fmt.Println("Token signature is invalid")
}
```

## API Reference

### Functions

#### `GenerateToken[T any](payload T, secret string) (string, error)`

Generates a token by JSON encoding the payload and appending an 8-byte truncated HMAC-SHA256 signature.

Parameters:

- `payload T`: The data to be encoded in the token (must be JSON serializable)
- `secret string`: The secret key used for signing the token

Returns:

- `string`: The generated token in the format `payload.signature`
- `error`: Any error that occurred during token generation

#### `ParseToken[T any](token string, secret string) (T, error)`

Verifies the token's signature and decodes the JSON payload into the specified type.

Parameters:

- `token string`: The token to parse and validate
- `secret string`: The secret key used for verifying the signature

Returns:

- `T`: The decoded payload
- `error`: Any error that occurred during parsing or validation

### Errors

- `ErrInvalidToken`: Returned when the token format is invalid
- `ErrSignatureInvalid`: Returned when the token signature doesn't match

## Security Considerations

1. Keep your secret key secure and never expose it in your code or version control
2. Use a strong, random secret key
3. Consider implementing token expiration in your payload if needed
4. Always validate tokens before trusting their contents

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This package is part of the SaaSKit project.
