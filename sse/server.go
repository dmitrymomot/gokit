package sse

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Server represents an SSE server that manages client connections
type Server struct {
	bus       MessageBus
	mu        sync.RWMutex
	clients   map[*Client]struct{}
	closed    bool
	hostname  string
	heartbeat time.Duration
}

// ServerOption is a function that configures a Server
type ServerOption func(*Server)

// WithHeartbeat sets the heartbeat interval for the server
func WithHeartbeat(d time.Duration) ServerOption {
	return func(s *Server) {
		s.heartbeat = d
	}
}

// WithHostname sets the hostname for the server
func WithHostname(hostname string) ServerOption {
	return func(s *Server) {
		s.hostname = hostname
	}
}

// NewServer creates a new SSE server with the provided message bus
// Panics if the bus parameter is nil
func NewServer(bus MessageBus, opts ...ServerOption) *Server {
	if bus == nil {
		panic("sse: message bus cannot be nil")
	}
	s := &Server{
		bus:       bus,
		clients:   make(map[*Client]struct{}),
		heartbeat: 30 * time.Second, // Default heartbeat interval
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Handler returns an HTTP handler function for SSE connections
// The topic is extracted from the provided extractor function
func (s *Server) Handler(topicExtractor func(r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only support GET requests
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Get the topic from the request
		topic := topicExtractor(r)
		if topic == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Set headers for SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no") // Disable buffering for Nginx

		// Ensure immediate flush
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Create a client with a context that cancels when the request is done
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		// Create a new client
		client := NewClient(w, s.hostname)

		// Register the client
		s.registerClient(client)
		defer s.removeClient(client)

		// Subscribe to the topic
		events, err := s.bus.Subscribe(ctx, topic)
		if err != nil {
			return
		}

		// Send a connected event
		client.Send(Event{
			Event: "connected",
			Data:  fmt.Sprintf("Connected to %s", topic),
		})

		// Set up heartbeat ticker
		ticker := time.NewTicker(s.heartbeat)
		defer ticker.Stop()

		// Event loop
		for {
			select {
			case <-ctx.Done():
				// Client disconnected
				return

			case event, ok := <-events:
				if !ok {
					// Channel closed
					return
				}
				// Send the event to the client
				err := client.Send(event)
				if err != nil {
					// Error sending event, client likely disconnected
					return
				}

			case <-ticker.C:
				// Send keep-alive to maintain the connection
				err := client.SendKeepAlive()
				if err != nil {
					// Error sending keep-alive, client likely disconnected
					return
				}
			}
		}
	}
}

// Publish sends an event to all clients subscribed to the specified topic
func (s *Server) Publish(ctx context.Context, topic string, event Event) error {
	return s.bus.Publish(ctx, topic, event)
}

// registerClient adds a client to the server
func (s *Server) registerClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client] = struct{}{}
}

// removeClient removes a client from the server
func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, client)
	client.Close()
}

// Close closes the server and all client connections
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Close all clients
	for client := range s.clients {
		client.Close()
		delete(s.clients, client)
	}

	// Close the message bus
	return s.bus.Close()
}