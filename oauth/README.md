# OAuth Package

Type-safe OAuth authentication clients for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/oauth
```

## Overview

The `oauth` package provides modular, type-safe OAuth client implementations for various providers. It simplifies the integration of OAuth authentication flows in Go applications with a focus on security, type safety, and developer experience.

## Available Providers

### Google OAuth 2.0

The Google OAuth provider allows simple integration of Google Sign-In into your application:

```go
import "github.com/dmitrymomot/gokit/oauth/google"

// Create a Google OAuth client
client, err := google.New(google.Config{
    ClientID:     "your-google-client-id",
    ClientSecret: "your-google-client-secret",
    RedirectURL:  "https://your-app.com/auth/google/callback",
}, logger)

// Generate authorization URL
authURL, _ := client.RedirectURL(state)

// Exchange code for user profile
profile, err := client.Auth(ctx, code)
```

For detailed documentation on Google OAuth, see the [Google OAuth README](/google/README.md).

## Common Features Across Providers

- Environment-based configuration
- Type-safe profile information
- Context support for cancellation and timeouts
- Standard error handling patterns
- Session state management helpers
- Security-focused defaults

## Best Practices

### State Parameter Security

Always use a cryptographically secure random string for the state parameter to prevent CSRF attacks:

```go
import (
    "crypto/rand"
    "encoding/base64"
)

func generateSecureState() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

### Secure Configuration Storage

Store OAuth client secrets securely using environment variables rather than hardcoding them in your application:

```go
import (
    "github.com/dmitrymomot/gokit/config"
    "github.com/dmitrymomot/gokit/oauth/google"
)

// Load from environment variables
cfg, err := config.Load[google.Config]()
```

### HTTPS for Redirect URLs

Always use HTTPS for redirect URLs in production environments. OAuth 2.0 recommends secure communication to prevent token interception.

## Error Handling

Each provider implements specific error types for common failure scenarios:

```go
import (
    "errors"
    "github.com/dmitrymomot/gokit/oauth/google"
)

// Handle specific errors
if errors.Is(err, google.ErrAccountNotVerified) {
    // Prompt user to verify their account
}
```

## License

This package is part of the gokit project and follows its licensing terms.
