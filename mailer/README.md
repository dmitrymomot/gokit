# Mailer Package

A type-safe email sending service with Postmark API integration.

## Installation

```bash
go get github.com/dmitrymomot/gokit/mailer
```

## Overview

The `mailer` package provides a clean interface for sending emails via the Postmark API. It offers context-aware operations, tracking capabilities for email opens and link clicks, and comprehensive error handling. The package is designed with thread-safety in mind, making it suitable for concurrent use in web services and APIs.

## Features

- Postmark API integration for reliable email delivery
- Email open and link click tracking
- Context-aware operations supporting cancellation and timeouts
- Type-safe configuration using environment variables
- Comprehensive error handling with semantic error types
- Thread-safe implementation for concurrent usage
- Simplified client interface with intuitive API

## Usage

### Basic Example

```go
import (
	"context"
	"github.com/dmitrymomot/gokit/mailer"
)

// Create a mailer client
client, err := mailer.NewClient(mailer.Config{
	PostmarkServerToken:  "your-postmark-server-token",
	PostmarkAccountToken: "your-postmark-account-token",
	SenderEmail:          "noreply@example.com",
	SupportEmail:         "support@example.com",
})
if err != nil {
	// Handle error
}

// Send an email
err = client.SendEmail(context.Background(), mailer.SendEmailParams{
	SendTo:   "user@example.com",
	Subject:  "Welcome!",
	BodyHTML: "<h1>Hello</h1><p>Welcome to our service!</p>",
	Tag:      "welcome",
})
// Email is sent with tracking enabled for opens and links
```

### Loading Config from Environment Variables

```go
import (
	"github.com/dmitrymomot/gokit/config"
	"github.com/dmitrymomot/gokit/mailer"
)

// Load configuration from environment variables
cfg, err := config.Load[mailer.Config]()
if err != nil {
	// Handle error
}

// Create client with loaded config
client, err := mailer.NewClient(cfg)
if err != nil {
	// Handle error
}
```

### Error Handling

```go
import (
	"errors"
	"github.com/dmitrymomot/gokit/mailer"
)

err := client.SendEmail(ctx, params)
if err != nil {
	if errors.Is(err, mailer.ErrFailedToSendEmail) {
		// Handle specific email sending failure
		// Check wrapped error for more details
	} else {
		// Handle other errors
	}
}
```

## Best Practices

1. **Context Usage**:
   - Always provide a context with appropriate timeout for email sending operations
   - Use context cancellation for gracefully handling service shutdowns

2. **Email Categorization**:
   - Use the `Tag` field to categorize emails for analytics and tracking
   - Keep tags consistent across similar email types

3. **Error Handling**:
   - Check for specific errors using `errors.Is()`
   - Log failed email attempts with appropriate context

4. **Configuration**:
   - Store API tokens in secure environment variables
   - Use the `config` package for type-safe loading of environment variables

## API Reference

### Configuration

```go
type Config struct {
	PostmarkServerToken  string `env:"POSTMARK_SERVER_TOKEN,required"`  // Postmark API server token
	PostmarkAccountToken string `env:"POSTMARK_ACCOUNT_TOKEN,required"` // Postmark API account token
	SenderEmail          string `env:"SENDER_EMAIL,required"`           // Email address of the sender
	SupportEmail         string `env:"SUPPORT_EMAIL,required"`          // Email address for customer support
}
```

### Types

```go
type SendEmailParams struct {
	SendTo   string `json:"send_to"`       // Email address of the recipient
	Subject  string `json:"subject"`       // Subject of the email
	BodyHTML string `json:"body_html"`     // HTML body of the email
	Tag      string `json:"tag,omitempty"` // Optional tag for categorization
}
```

### Interfaces

```go
type EmailSender interface {
	SendEmail(ctx context.Context, params SendEmailParams) error
}
```

### Functions

```go
func NewClient(cfg Config) (EmailSender, error)
```
Creates a new instance of the mailer client with the provided configuration.

```go
func MustNewClient(cfg Config) EmailSender
```
Creates a new instance of the mailer client, panics if initialization fails.

### Error Types

```go
var ErrFailedToSendEmail = errors.New("failed to send email")
