package sse

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Client represents a connected SSE client
type Client struct {
	writer   http.ResponseWriter
	flusher  http.Flusher
	mu       sync.Mutex
	closed   bool
	hostname string
}

// NewClient creates a new SSE client
func NewClient(w http.ResponseWriter, hostname string) *Client {
	return &Client{
		writer:   w,
		flusher:  w.(http.Flusher),
		hostname: hostname,
	}
}

// Send sends an event to the client
func (c *Client) Send(event Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return ErrClientClosed
	}

	// Add timestamp and hostname if no ID is present
	if event.ID == "" {
		event.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), c.hostname)
	}

	// Write the event to the response
	err := event.Write(c.writer)
	if err != nil {
		c.closed = true
		return err
	}

	// Flush the response writer
	c.flusher.Flush()

	return nil
}

// Close marks the client as closed
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
}