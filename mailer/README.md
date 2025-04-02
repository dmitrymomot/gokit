# Mailer

A package for sending emails with support for Postmark API, HTML templates, and tracking capabilities.

## Overview

The mailer package provides a flexible and type-safe interface for sending emails in Go applications. It integrates with the Postmark API for reliable email delivery and includes a templating system based on [templ](https://github.com/a-h/templ) for creating beautiful, responsive email templates.

## Features

- Type-safe configuration using environment variables
- Integration with Postmark API for reliable email delivery
- HTML email format with tracking for opens and links
- Context support for cancellation and timeouts
- Comprehensive error handling
- Reusable email components for consistent design
- Email template rendering utilities

## Installation

```bash
go get github.com/dmitrymomot/gokit/mailer
```

## Usage

### Basic Setup

```go
package main

import (
	"context"
	"log"

	"github.com/dmitrymomot/gokit/mailer"
)

func main() {
	// Create a new mailer client with configuration
	config := mailer.Config{
		PostmarkServerToken:  "your-postmark-server-token",
		PostmarkAccountToken: "your-postmark-account-token",
		SenderEmail:          "noreply@example.com",
		SupportEmail:         "support@example.com",
	}

	// Create a new client
	client, err := mailer.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create mailer client: %v", err)
	}

	// Send an email
	err = client.SendEmail(context.Background(), mailer.SendEmailParams{
		SendTo:   "user@example.com",
		Subject:  "Welcome to Our Service",
		BodyHTML: "<h1>Welcome!</h1><p>Thank you for signing up.</p>",
		Tag:      "welcome",
	})
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
}
```

### Using Email Templates

The package includes a templating system based on `templ` for creating beautiful, responsive email templates:

```go
package main

import (
	"context"
	"log"

	"github.com/dmitrymomot/gokit/mailer"
	"github.com/dmitrymomot/gokit/mailer/templates"
	"github.com/dmitrymomot/gokit/mailer/templates/components"
)

func main() {
	// Create a new mailer client
	client := mailer.MustNewClient(mailer.Config{
		PostmarkServerToken:  "your-postmark-server-token",
		PostmarkAccountToken: "your-postmark-account-token",
		SenderEmail:          "noreply@example.com",
		SupportEmail:         "support@example.com",
	})

	// Create a context
	ctx := context.Background()

	// Create an email template using the provided components
	emailTemplate := components.Layout(
		"Welcome to Our Service",
		components.Header("Welcome!"),
		components.Text("Thank you for signing up for our service."),
		components.Button("Get Started", "https://example.com/start"),
	)

	// Render the template to HTML
	htmlBody, err := templates.Render(ctx, emailTemplate)
	if err != nil {
		log.Fatalf("Failed to render email template: %v", err)
	}

	// Send the email with the rendered template
	err = client.SendEmail(ctx, mailer.SendEmailParams{
		SendTo:   "user@example.com",
		Subject:  "Welcome to Our Service",
		BodyHTML: htmlBody,
		Tag:      "welcome",
	})
	if err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}
}
```

### Available Template Components

The package provides several reusable components for building email templates:

- `Layout`: The main layout component that wraps all other components
- `Header`: For section headers in the email
- `Text`: For paragraphs of text
- `Button`: For call-to-action buttons
- `Link`: For hyperlinks
- `Logo`: For displaying your company logo
- `Footer`: For email footers with company information
- `OTP`: For one-time password display

## Error Handling

The package defines specific error types for handling email-related errors:

```go
var (
	ErrFailedToSendEmail = errors.New("failed to send email")
)
```

Errors from the Postmark API are joined with the above error for more context.

## Configuration

The `Config` struct provides essential configuration options for the mailer:

```go
type Config struct {
	PostmarkServerToken  string `env:"POSTMARK_SERVER_TOKEN,required"`  // Postmark API server token
	PostmarkAccountToken string `env:"POSTMARK_ACCOUNT_TOKEN,required"` // Postmark API account token
	SenderEmail          string `env:"SENDER_EMAIL,required"`           // Email address of the sender
	SupportEmail         string `env:"SUPPORT_EMAIL,required"`          // Email address for customer support
}
```

These configuration options can be loaded from environment variables using the `config` package in gokit.
