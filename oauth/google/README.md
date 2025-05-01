# Google OAuth Package

A lightweight, type-safe Google OAuth 2.0 client for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/oauth/google
```

## Overview

The `google` package provides a simple interface for authenticating users with Google OAuth 2.0. It handles the complete authentication flow, from generating consent URLs to exchanging codes for tokens and retrieving user profiles. The package is thread-safe and designed for use in web applications with proper context propagation.

## Features

- Complete Google OAuth 2.0 authentication flow
- User profile retrieval with type safety
- Optional verification filtering to ensure only verified accounts
- Environment-based configuration with sensible defaults
- Context-aware for cancellation and timeout support
- Comprehensive error handling with specific error types

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
    // Optional: custom scopes, defaults to "openid,profile,email"
    Scopes:       []string{"openid", "profile", "email"},
    // Optional: only allow verified accounts, defaults to true
    VerifiedOnly: true,
}, logger) // Pass your logger implementation
if err != nil {
    // Handle initialization error
}

// Step 1: Redirect to Google's consent page
http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
    // Generate a secure random state to prevent CSRF
    state := generateSecureRandomString()
    
    // Store state in session
    saveToSession(r, "google_oauth_state", state)
    
    // Get authorization URL
    authURL, err := client.RedirectURL(state)
    if err != nil {
        http.Error(w, "Failed to generate auth URL", http.StatusInternalServerError)
        return
    }
    
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
        // Handle specific errors
        switch {
        case errors.Is(err, google.ErrAccountNotVerified):
            http.Error(w, "Account not verified", http.StatusForbidden)
        case errors.Is(err, google.ErrFailedToExchangeCode):
            http.Error(w, "Failed to exchange code", http.StatusBadRequest)
        default:
            http.Error(w, "Authentication failed", http.StatusInternalServerError)
        }
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
if err != nil {
    // Handle error
}
```

## Best Practices

1. **Security**:
   - Store client secret securely using environment variables
   - Generate cryptographically secure random state parameters
   - Always validate state parameter to prevent CSRF attacks
   - Use HTTPS for redirect URLs as OAuth 2.0 requires secure communication

2. **Account Verification**:
   - Keep `VerifiedOnly` enabled (default) to prevent fake accounts
   - Handle unverified account errors appropriately

3. **Error Handling**:
   - Check for specific error types using `errors.Is()`
   - Provide clear error messages to users when authentication fails
   - Log authentication failures for security monitoring

4. **Performance**:
   - Reuse the client instance throughout your application
   - Use context with timeouts for HTTP requests to prevent hanging

## API Reference

### Types

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

```go
type Profile struct {
    ID            string `json:"id"`
    Email         string `json:"email"`
    VerifiedEmail bool   `json:"verified_email"`
    Picture       string `json:"picture,omitempty"`
    Name          string `json:"name,omitempty"`
    FamilyName    string `json:"family_name,omitempty"`
    GivenName     string `json:"given_name,omitempty"`
    Locale        string `json:"locale,omitempty"`
}
```

```go
type logger interface {
    WarnContext(ctx context.Context, msg string, args ...any)
    ErrorContext(ctx context.Context, msg string, args ...any)
}
```

### Functions

```go
func New(cfg Config, log logger) (*Client, error)
```
Creates a new Google OAuth 2.0 client with the provided configuration and logger.

### Methods

```go
func (c *Client) RedirectURL(state string) (string, error)
```
Returns the URL to redirect the user to Google's OAuth 2.0 consent page with the provided state parameter.

```go
func (c *Client) GetProfile(ctx context.Context, token string) (Profile, error)
```
Retrieves the user's profile from Google using the provided oauth access token.

```go
func (c *Client) Auth(ctx context.Context, code string) (Profile, error)
```
Exchanges the authorization code for an access token and retrieves the user's profile from Google.

### Error Types

```go
var ErrFailedToGetProfile = errors.New("failed to get profile")
var ErrInvalidState = errors.New("invalid oauth state")
var ErrFailedToExchangeCode = errors.New("failed to exchange code")
var ErrAccountNotVerified = errors.New("account is not verified")
var ErrFailedToSaveSession = errors.New("failed to save session")
```
