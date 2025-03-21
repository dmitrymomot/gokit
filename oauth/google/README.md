# Google OAuth 2.0 Client

A lightweight and easy-to-use Google OAuth 2.0 client for Go applications.

## Overview

This package provides a simple and intuitive API for integrating Google OAuth 2.0 authentication into your Go applications. It handles the entire OAuth flow from generating authorization URLs to retrieving user profiles.

## Features

- Simple API for Google OAuth 2.0 authentication
- User profile retrieval
- Email verification filtering
- Configurable via environment variables
- Error handling with specific error types

## Installation

```bash
go get github.com/dmitrymomot/gokit/oauth/google
```

## Configuration

The package can be configured using environment variables or directly via the `Config` struct:

```go
type Config struct {
    ClientID     string   // Google OAuth client ID
    ClientSecret string   // Google OAuth client secret
    RedirectURL  string   // OAuth redirect URL
    Scopes       []string // OAuth scopes (default: "openid,profile,email")
    StateKey     string   // Key to store the state in the session (default: "google_oauth_state")
    VerifiedOnly bool     // Only allow verified accounts (default: true)
}
```

Environment variable mapping:
- `GOOGLE_OAUTH_CLIENT_ID` - Google OAuth client ID (required)
- `GOOGLE_OAUTH_CLIENT_SECRET` - Google OAuth client secret (required)
- `GOOGLE_OAUTH_REDIRECT_URL` - OAuth redirect URL (required)
- `GOOGLE_OAUTH_SCOPES` - OAuth scopes (default: "openid,profile,email")
- `GOOGLE_OAUTH_STATE_KEY` - Key to store the state (default: "google_oauth_state")
- `GOOGLE_OAUTH_VERIFIED_ONLY` - Only allow verified accounts (default: true)

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    "github.com/dmitrymomot/gokit/oauth/google"
)

func main() {
    // Create a new Google OAuth client
    cfg := google.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/auth/google/callback",
    }
    
    client, err := google.New(cfg)
    if err != nil {
        panic(err)
    }

    // Handle OAuth login request
    http.HandleFunc("/auth/google", func(w http.ResponseWriter, r *http.Request) {
        // Generate a random state
        state := "random-state-string" // In production, use a secure random generator
        
        // Get the authorization URL
        authURL, err := client.RedirectURL(state)
        if err != nil {
            http.Error(w, "Failed to generate auth URL", http.StatusInternalServerError)
            return
        }
        
        // Store the state in the session (implementation depends on your session management)
        // saveSession(r, "google_oauth_state", state)
        
        // Redirect to Google's OAuth consent page
        http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
    })

    // Handle OAuth callback
    http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
        // Get the state and code from the request
        state := r.URL.Query().Get("state")
        code := r.URL.Query().Get("code")
        
        // Verify the state (implementation depends on your session management)
        // expectedState := getSession(r, "google_oauth_state")
        expectedState := "random-state-string" // For example only
        
        if state != expectedState {
            http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
            return
        }
        
        // Exchange the authorization code for a token and get the user profile
        profile, err := client.Auth(context.Background(), code)
        if err != nil {
            http.Error(w, "Authentication failed: "+err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Authentication successful, use the profile information
        fmt.Fprintf(w, "Authenticated as: %s (%s)", profile.Name, profile.Email)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Profile Information

The package returns a `Profile` struct with the following fields:

```go
type Profile struct {
    ID            string // User's unique ID
    Email         string // User's email address
    VerifiedEmail bool   // Whether the email is verified
    Picture       string // URL to user's profile picture
    Name          string // User's full name
    FamilyName    string // User's family name
    GivenName     string // User's given name
    Locale        string // User's locale
}
```

### Error Handling

The package provides specific error types for common failure scenarios:

```go
var (
    ErrFailedToGetProfile   = errors.New("failed to get profile")
    ErrInvalidState         = errors.New("invalid oauth state")
    ErrFailedToExchangeCode = errors.New("failed to exchange code")
    ErrAccountNotVerified   = errors.New("account is not verified")
    ErrFailedToSaveSession  = errors.New("failed to save session")
)
```

## Security Considerations

- Always generate secure random strings for the state parameter to prevent CSRF attacks
- Store the OAuth client secret securely (e.g., using environment variables)
- Set `VerifiedOnly` to true (default) to ensure only verified Google accounts can authenticate
