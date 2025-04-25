# Google OAuth Package

A lightweight, type-safe Google OAuth 2.0 client for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/oauth/google
```

## Overview

The `google` package provides a simple interface for authenticating users with Google OAuth 2.0. It handles the complete authentication flow, from generating consent URLs to exchanging codes for tokens and retrieving user profiles.

## Features

- Complete Google OAuth 2.0 authentication flow
- User profile retrieval with type safety
- Optional verification filtering to ensure only verified accounts
- Environment-based configuration
- Context-aware for cancellation support
- Comprehensive error handling

## Usage

### Basic Authentication Flow

```go
import (
    "context"
    "net/http"
    "github.com/dmitrymomot/gokit/oauth/google"
)

// Initialize the client
client, err := google.New(google.Config{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURL:  "https://your-app.com/auth/callback",
}, logger) // Pass your logger implementation

// Step 1: Redirect to Google's consent page
http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    // Generate a secure random state to prevent CSRF
    state := generateSecureRandomString()
    
    // Store state in session
    saveToSession(r, "google_oauth_state", state)
    
    // Get authorization URL
    authURL, _ := client.RedirectURL(state)
    
    // Redirect user to Google
    http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
})

// Step 2: Handle the callback from Google
http.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
    // Get state and code from Google's response
    state := r.URL.Query().Get("state")
    code := r.URL.Query().Get("code")
    
    // Verify state to prevent CSRF attacks
    expectedState := getFromSession(r, "google_oauth_state")
    if state != expectedState {
        http.Error(w, "Invalid state", http.StatusBadRequest)
        return
    }
    
    // Exchange code for token and get user profile
    profile, err := client.Auth(r.Context(), code)
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusInternalServerError)
        return
    }
    
    // Use profile data (profile.Email, profile.Name, etc.)
    // ...
})
```

### Environment Configuration

```go
import (
    "github.com/dmitrymomot/gokit/config"
    "github.com/dmitrymomot/gokit/oauth/google"
)

// Load config from environment variables
cfg, err := config.Load[google.Config]()
if err != nil {
    // Handle error
}

// Create client with loaded config
client, err := google.New(cfg, logger)
```

### User Profile Information

The `Profile` struct contains information about the authenticated user:

```go
type Profile struct {
    ID            string // Google user ID
    Email         string // User's email address
    VerifiedEmail bool   // Whether email is verified
    Picture       string // Profile picture URL
    Name          string // Full name
    FamilyName    string // Last name
    GivenName     string // First name
    Locale        string // User's locale
}
```

## Configuration

```go
type Config struct {
    ClientID     string   `env:"GOOGLE_OAUTH_CLIENT_ID,required"`
    ClientSecret string   `env:"GOOGLE_OAUTH_CLIENT_SECRET,required"`
    RedirectURL  string   `env:"GOOGLE_OAUTH_REDIRECT_URL,required"`
    Scopes       []string `env:"GOOGLE_OAUTH_SCOPES" envDefault:"openid,profile,email"`
    StateKey     string   `env:"GOOGLE_OAUTH_STATE_KEY" envDefault:"google_oauth_state"`
    VerifiedOnly bool     `env:"GOOGLE_OAUTH_VERIFIED_ONLY" envDefault:"true"`
}
```

| Option | Description | Default |
|--------|-------------|---------|
| `ClientID` | OAuth client ID from Google | Required |
| `ClientSecret` | OAuth client secret from Google | Required |
| `RedirectURL` | URL to redirect after authentication | Required |
| `Scopes` | OAuth scopes to request | `openid,profile,email` |
| `StateKey` | Key used for storing state | `google_oauth_state` |
| `VerifiedOnly` | Only allow verified Google accounts | `true` |

## API Reference

### Client Creation

- `New(cfg Config, log logger) (*Client, error)`: Create a new Google OAuth client

### Authentication Flow

- `RedirectURL(state string) (string, error)`: Generate Google consent page URL
- `Auth(ctx context.Context, code string) (Profile, error)`: Exchange auth code for profile
- `GetProfile(ctx context.Context, token string) (Profile, error)`: Get profile from access token

### Error Handling

```go
// Check for specific errors
if errors.Is(err, google.ErrAccountNotVerified) {
    // Handle unverified account
}

if errors.Is(err, google.ErrFailedToExchangeCode) {
    // Handle code exchange failure
}
```

## Security Best Practices

1. **Store client secret securely**: Use environment variables instead of hardcoding
2. **Generate secure random state**: Use a cryptographically secure random generator
3. **Validate state parameter**: Always verify to prevent CSRF attacks
4. **Use HTTPS for redirect URLs**: OAuth 2.0 requires secure communication
5. **Verify email addresses**: Keep `VerifiedOnly` enabled to prevent fake accounts
