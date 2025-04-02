package mailer

import (
	"context"
	"errors"
	"fmt"

	"github.com/mrz1836/postmark"
)

// client implements the EmailSender interface.
type client struct {
	client *postmark.Client
	config Config
}

// NewClient creates a new instance of the mailer client
// with the provided server token and account token from the config (see mailer.Config).
// The client is used to send emails synchronously using the Postmark API.
// For asynchronous email sending, use the email enqueuer.
func NewClient(cfg Config) (EmailSender, error) {
	return &client{
		client: postmark.NewClient(cfg.PostmarkServerToken, cfg.PostmarkAccountToken),
		config: cfg,
	}, nil
}

// MustNewClient creates a new instance of the mailer client
// with the provided server token and account token from the config (see mailer.Config).
// The client is used to send emails synchronously using the Postmark API.
// For asynchronous email sending, use the email enqueuer.
// Panics if the config cannot be loaded.
func MustNewClient(cfg Config) EmailSender {
	client, err := NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

// SendEmail sends an email using the Postmark API with tracking enabled for opens and links.
// It uses the configured sender email as the "From" address and support email as "Reply-To".
// Returns an error if the send fails or if Postmark returns an error response.
func (c *client) SendEmail(ctx context.Context, params SendEmailParams) error {
	resp, err := c.client.SendEmail(ctx, postmark.Email{
		From:       c.config.SenderEmail,
		ReplyTo:    c.config.SupportEmail,
		To:         params.SendTo,
		Subject:    params.Subject,
		Tag:        params.Tag,
		HTMLBody:   params.BodyHTML,
		TrackOpens: true,
		TrackLinks: "HtmlOnly",
	})
	if err != nil {
		return errors.Join(ErrFailedToSendEmail, err)
	}
	if resp.ErrorCode > 0 {
		return errors.Join(
			ErrFailedToSendEmail,
			fmt.Errorf("postmark error: %d - %s", resp.ErrorCode, resp.Message),
		)
	}
	return nil
}
