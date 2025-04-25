# Mailer Package

Type-safe email sending with Postmark API integration and HTML templates.

## Installation

```bash
go get github.com/dmitrymomot/gokit/mailer
```

## Overview

The `mailer` package provides a clean interface for sending emails via Postmark with HTML templating support. It focuses on type safety, tracking capabilities, and reusable components for consistent email design.

## Features

- Postmark API integration for reliable delivery
- HTML templating with reusable components
- Email open and link tracking
- Context-aware for cancellation support
- Type-safe configuration using environment variables
- Simple client interface with comprehensive error handling

## Usage

### Quick Start

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

// Send a simple email
err = client.SendEmail(context.Background(), mailer.SendEmailParams{
	SendTo:   "user@example.com",
	Subject:  "Welcome!",
	BodyHTML: "<h1>Hello</h1><p>Welcome to our service!</p>",
	Tag:      "welcome",
})
```

### With HTML Templates

```go
import (
	"context"
	"github.com/dmitrymomot/gokit/mailer"
	"github.com/dmitrymomot/gokit/mailer/templates"
	"github.com/dmitrymomot/gokit/mailer/templates/components"
)

// Create a mailer client
client := mailer.MustNewClient(mailer.Config{
	PostmarkServerToken:  "your-token",
	PostmarkAccountToken: "your-token",
	SenderEmail:          "noreply@example.com",
	SupportEmail:         "support@example.com",
})

// Build an email using components
emailTemplate := components.Layout(
	components.Header("Welcome to Our Platform"),
	components.Text("We're excited to have you join us."),
	components.Button("Get Started", "https://example.com/start"),
	components.Footer(" 2025 Example Inc."),
)

// Render template to HTML
htmlBody, err := templates.Render(context.Background(), emailTemplate)
if err != nil {
	// Handle error
}

// Send email with the rendered template
err = client.SendEmail(context.Background(), mailer.SendEmailParams{
	SendTo:   "user@example.com",
	Subject:  "Welcome!",
	BodyHTML: htmlBody,
	Tag:      "onboarding",
})
```

### Loading Config from Environment

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
```

## Template Components

The package includes reusable components for creating consistent email templates:

- `Layout`: Base container for email content
- `Header`: Section headers with appropriate styling
- `Text`: Formatted text paragraphs
- `Button`: Call-to-action buttons with tracking
- `Link`: Hyperlinks with tracking capabilities
- `Logo`: Display company logo
- `OTP`: Format for one-time passwords/codes
- `Footer`: Standard footer with unsubscribe options

## API Reference

### Client Creation

- `NewClient(cfg Config) (EmailSender, error)`: Create new email client
- `MustNewClient(cfg Config) EmailSender`: Create client, panics on error

### Configuration

```go
type Config struct {
	PostmarkServerToken  string `env:"POSTMARK_SERVER_TOKEN,required"`
	PostmarkAccountToken string `env:"POSTMARK_ACCOUNT_TOKEN,required"`
	SenderEmail          string `env:"SENDER_EMAIL,required"`
	SupportEmail         string `env:"SUPPORT_EMAIL,required"`
}
```

### Sending Emails

```go
type SendEmailParams struct {
	SendTo   string // Recipient email address
	Subject  string // Email subject
	BodyHTML string // HTML email content
	Tag      string // Optional tag for categorization
}
```

- `SendEmail(ctx context.Context, params SendEmailParams) error`: Send an email with tracking

### Template Rendering

- `templates.Render(ctx context.Context, tpl templ.Component) (string, error)`: Render a template to HTML

## Error Handling

```go
// Check for specific errors
if errors.Is(err, mailer.ErrFailedToSendEmail) {
	// Handle email sending failure
}
```

## Best Practices

1. Use tags to categorize emails for analytics
2. Enable tracking for open and link metrics
3. Provide both text and HTML versions for compatibility
4. Use responsive design in templates
5. Set appropriate timeout in context for email sending operations
