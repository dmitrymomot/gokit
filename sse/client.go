package sse

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Client represents a connected SSE client
type Client struct {
	// ID uniquely identifies this client
	ID string

	// Channels this client is subscribed to
	channels map[string]struct{}

	// Response writer for sending events
	writer http.ResponseWriter

	// Mutex for thread safety
	mu sync.RWMutex

	// Notify when client disconnects
	done chan struct{}
}

// newClient creates a new SSE client with the given ID
func newClient(id string, w http.ResponseWriter) *Client {
	return &Client{
		ID:       id,
		writer:   w,
		channels: make(map[string]struct{}),
		done:     make(chan struct{}),
	}
}

// Send writes a message to the client's event stream
func (c *Client) Send(ctx context.Context, msg Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return ErrClientNotConnected
	default:
		// Convert message to SSE format
		eventStr, err := msg.ToEventString()
		if err != nil {
			return fmt.Errorf("failed to format message: %w", err)
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		// Write to response
		_, err = fmt.Fprint(c.writer, eventStr)
		if err != nil {
			// If we can't write to the client, assume they've disconnected
			close(c.done)
			return fmt.Errorf("failed to send to client: %w", err)
		}

		// Flush to ensure the message is sent immediately
		if flusher, ok := c.writer.(http.Flusher); ok {
			flusher.Flush()
		}

		return nil
	}
}

// AddChannel subscribes the client to a channel
func (c *Client) AddChannel(channel string) {
	if channel == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels[channel] = struct{}{}
}

// RemoveChannel unsubscribes the client from a channel
func (c *Client) RemoveChannel(channel string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.channels, channel)
}

// IsSubscribedTo checks if the client is subscribed to a channel
func (c *Client) IsSubscribedTo(channel string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, subscribed := c.channels[channel]
	return subscribed
}

// Close marks the client as disconnected
func (c *Client) Close() {
	select {
	case <-c.done:
		// Already closed
		return
	default:
		close(c.done)
	}
}

// IsClosed checks if the client is closed
func (c *Client) IsClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

// SendKeepAlive sends a comment to keep the connection alive
func (c *Client) SendKeepAlive(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return ErrClientNotConnected
	default:
		c.mu.Lock()
		defer c.mu.Unlock()
		
		// Send a comment as a keep-alive
		_, err := fmt.Fprintf(c.writer, ": %s\n\n", time.Now().Format(time.RFC3339))
		if err != nil {
			close(c.done)
			return fmt.Errorf("failed to send keep-alive: %w", err)
		}

		if flusher, ok := c.writer.(http.Flusher); ok {
			flusher.Flush()
		}

		return nil
	}
}
