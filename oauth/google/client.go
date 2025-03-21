package google

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Client represents the Google OAuth 2.0 client.
type Client struct {
	oauth        *oauth2.Config
	stateKey     string // The key to store the state in the session.
	verifiedOnly bool   // If true, only verified accounts are allowed. Default is true.
	log          logger
}

type logger interface {
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

// Profile represents the user's profile from Google.
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

// New creates a new Google OAuth 2.0 client.
func New(cfg Config, log logger) (*Client, error) {
	return &Client{
		oauth: &oauth2.Config{
			RedirectURL:  cfg.RedirectURL,
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       cfg.Scopes,
			Endpoint:     google.Endpoint,
		},
		stateKey:     cfg.StateKey,
		verifiedOnly: cfg.VerifiedOnly,
		log:          log,
	}, nil
}

// RedirectURL returns the URL to redirect the user to Google's OAuth 2.0 consent page.
// The state parameter is used by the application to prevent CSRF attacks.
// The state parameter should be a random string.
func (c *Client) RedirectURL(state string) (string, error) {
	return c.oauth.AuthCodeURL(state), nil
}

// GetProfile retrieves the user's profile from Google using the provided oauth access token.
func (c *Client) GetProfile(ctx context.Context, token string) (Profile, error) {
	// Get the user's profile.
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return Profile{}, errors.Join(ErrFailedToGetProfile)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			c.log.ErrorContext(ctx, "Failed to close response body", "error", err)
		}
	}(response.Body)

	// Parse the user's profile.
	var profile Profile
	if err := json.NewDecoder(response.Body).Decode(&profile); err != nil {
		return Profile{}, errors.Join(ErrFailedToGetProfile, err)
	}

	return profile, nil
}

// Auth function is a final step of the Google authentication process.
// It exchanges the authorization code for an access token and retrieves the user's profile from Google.
// The function returns the user's profile and an error if any
func (c *Client) Auth(ctx context.Context, code string) (Profile, error) {
	// Exchange the authorization code for an access token.
	token, err := c.oauth.Exchange(ctx, code)
	if err != nil {
		return Profile{}, errors.Join(ErrFailedToExchangeCode, err)
	}

	// Get the user's profile.
	profile, err := c.GetProfile(ctx, token.AccessToken)
	if err != nil {
		return Profile{}, err
	}
	if c.verifiedOnly && !profile.VerifiedEmail {
		c.log.WarnContext(ctx, "Account is not verified", "email", profile.Email)
		return Profile{}, ErrAccountNotVerified
	}

	return profile, nil
}
