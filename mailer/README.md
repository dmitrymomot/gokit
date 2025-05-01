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
- HTML templating with reusable components

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
	components.PrimaryButton("Get Started", "https://example.com/start"),
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

### Creating Custom Templates

You can create your own email templates using the `templ` library. Here's an example of creating a welcome email template:

1. Create a file named `welcome_email.templ`:

```go
package templates

import (
	"github.com/dmitrymomot/gokit/mailer/templates/components"
)

// WelcomeEmail creates a welcome email template with user's name and confirmation link
templ WelcomeEmail(userName, confirmationLink string) {
	@components.Layout() {
		@components.Header("Welcome to Our Service") 
		@components.Text("Hello " + userName + ",") 
		@components.Text("Thank you for joining our service! We're excited to have you on board.") 
		@components.Text("Please confirm your account to get started:") 
		@components.ButtonGroup() {
			@components.PrimaryButton("Confirm Account", confirmationLink) 
		}
		@components.Text("If you have any questions, feel free to contact our support team.") 
		@components.Footer(" 2025 Our Company") 
	}
}
```

2. After saving the template, run the `templ generate` command to generate the Go code:

```bash
templ generate
```

3. Use the template in your application code:

```go
import (
	"context"
	
	"github.com/dmitrymomot/gokit/mailer"
	"github.com/dmitrymomot/gokit/mailer/templates"
	"your-package-path/templates" // Import your custom templates package
)

func sendWelcomeEmail(ctx context.Context, client mailer.EmailSender, userEmail, userName string) error {
	// Create the template with user-specific data
	emailTemplate := mytemplates.WelcomeEmail(
		userName,
		"https://example.com/confirm?token=abc123",
	)
	
	// Render the template to HTML
	htmlBody, err := templates.Render(ctx, emailTemplate)
	if err != nil {
		return err
	}
	
	// Send the email with the rendered HTML
	return client.SendEmail(ctx, mailer.SendEmailParams{
		SendTo:   userEmail,
		Subject:  "Welcome to Our Service!",
		BodyHTML: htmlBody,
		Tag:      "welcome",
	})
}
```

This pattern allows you to create reusable, type-safe email templates with strong separation of concerns between the template design and the email sending logic.

## Template Components

The package includes reusable email template components built with the `templ` library:

```go
import "github.com/dmitrymomot/gokit/mailer/templates/components"
```

### Available Components

- `Layout`: Base container with responsive styling for email content
  ```go
  components.Layout(/* child components */)
  ```

- `Header`: Section headers with standardized styling
  ```go
  components.Header("Welcome to Our Service")
  ```

- `Text`: Formatted text paragraphs with customizable styling
  ```go
  components.Text("This is a paragraph of text.")
  ```

- `Button`: Call-to-action buttons with various styles
  ```go
  components.PrimaryButton("Get Started", "https://example.com/start")
  components.SuccessButton("Confirm", "https://example.com/confirm")
  components.DangerButton("Delete", "https://example.com/delete")
  ```

- `Link`: Styled hyperlinks for navigation
  ```go
  components.Link("Visit our website", "https://example.com")
  ```

- `Logo`: Display company logo in emails
  ```go
  components.Logo("https://example.com/logo.png", "Company Name")
  ```

- `OTP`: One-time password/code formatting
  ```go
  components.OTP("123456")
  ```

- `Footer`: Standard footer with customizable content
  ```go
  components.Footer(" 2025 Company Name")
  ```

### Template Rendering

The `templates` package provides a rendering function to convert template components to HTML:

```go
func Render(ctx context.Context, tpl templ.Component) (string, error)
```

This function takes a template component and renders it to an HTML string that can be used with the `SendEmail` method.

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

5. **Templates**:
   - Compose complex templates from smaller, reusable components
   - Test email rendering in multiple email clients
   - Use responsive design components for mobile compatibility

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

```go
func templates.Render(ctx context.Context, tpl templ.Component) (string, error)
```
Renders a template component to an HTML string.

### Error Types

```go
var ErrFailedToSendEmail = errors.New("failed to send email")
```
