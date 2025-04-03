package sse

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	// DefaultKeepAliveInterval is the default interval for sending keep-alive messages
	DefaultKeepAliveInterval = 30 * time.Second
)

// Server represents an SSE server that distributes messages to connected clients
type Server struct {
	// broker handles message distribution
	broker Broker

	// clients maps client IDs to connected clients
	clients map[string]*Client

	// channels maps channel names to sets of client IDs
	channels map[string]map[string]*Client

	// mutex for thread safety
	mu sync.RWMutex

	// ctx is the server context
	ctx context.Context

	// cancel cancels the server context
	cancel context.CancelFunc

	// wg tracks ongoing operations
	wg sync.WaitGroup

	// keepAliveInterval determines how often to send keep-alive messages
	keepAliveInterval time.Duration
}

// ServerOption defines a server configuration option
type ServerOption func(*Server)

// WithKeepAliveInterval sets the keep-alive interval
func WithKeepAliveInterval(interval time.Duration) ServerOption {
	return func(s *Server) {
		s.keepAliveInterval = interval
	}
}

// NewServer creates a new SSE server with the given broker
func NewServer(broker Broker, opts ...ServerOption) (*Server, error) {
	if broker == nil {
		return nil, ErrNoBrokerProvided
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		broker:            broker,
		clients:           make(map[string]*Client),
		channels:          make(map[string]map[string]*Client),
		ctx:               ctx,
		cancel:            cancel,
		keepAliveInterval: DefaultKeepAliveInterval,
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Start broker subscription
	s.wg.Add(1)
	go s.listenToBroker()

	// Start keep-alive routine
	s.wg.Add(1)
	go s.keepAliveRoutine()

	return s, nil
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only support GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract client ID from the request
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(w, "Missing client_id parameter", http.StatusBadRequest)
		return
	}

	// Optional channel subscription
	channel := r.URL.Query().Get("channel")

	// Create context that cancels when client disconnects
	ctx := r.Context()

	// Serve the client
	if err := s.ServeClient(ctx, w, clientID, channel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ServeClient handles an SSE connection for a specific client
func (s *Server) ServeClient(ctx context.Context, w http.ResponseWriter, clientID, channel string) error {
	// Ensure server is running
	select {
	case <-s.ctx.Done():
		return ErrServerClosed
	default:
		// Continue
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // For Nginx

	// Check if client already exists
	s.mu.Lock()
	if _, exists := s.clients[clientID]; exists {
		s.mu.Unlock()
		return fmt.Errorf("%w: client_id %s", ErrClientAlreadyExists, clientID)
	}

	// Create new client
	client := newClient(clientID, w)
	s.clients[clientID] = client

	// Add to channel if specified
	if channel != "" {
		if _, exists := s.channels[channel]; !exists {
			s.channels[channel] = make(map[string]*Client)
		}
		s.channels[channel][clientID] = client
		client.AddChannel(channel)
	}
	s.mu.Unlock()

	// Send initial event to confirm connection
	initialMsg := NewMessage("connected", map[string]string{
		"client_id": clientID,
		"channel":   channel,
		"time":      time.Now().Format(time.RFC3339),
	})
	if err := client.Send(ctx, initialMsg); err != nil {
		s.removeClient(clientID)
		return fmt.Errorf("failed to send initial message: %w", err)
	}

	// Wait for client disconnection
	select {
	case <-ctx.Done():
		// Client disconnected - clean up
	case <-s.ctx.Done():
		// Server shutting down
	case <-client.done:
		// Client connection closed
	}

	// Clean up client
	s.removeClient(clientID)
	return nil
}

// removeClient removes a client from the server
func (s *Server) removeClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client, exists := s.clients[clientID]
	if !exists {
		return
	}

	// Close client
	client.Close()

	// Remove from channels
	for channel := range s.channels {
		delete(s.channels[channel], clientID)
		// Clean up empty channels
		if len(s.channels[channel]) == 0 {
			delete(s.channels, channel)
		}
	}

	// Remove from clients map
	delete(s.clients, clientID)
}

// listenToBroker subscribes to the broker and dispatches messages to clients
func (s *Server) listenToBroker() {
	defer s.wg.Done()

	msgCh, err := s.broker.Subscribe(s.ctx)
	if err != nil {
		// If we can't subscribe, server is unusable
		s.cancel()
		return
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case msg, ok := <-msgCh:
			if !ok {
				// Channel closed, exit
				return
			}
			s.dispatchMessage(msg)
		}
	}
}

// dispatchMessage sends a message to targeted clients
func (s *Server) dispatchMessage(msg Message) {
	// Skip expired messages
	if msg.IsExpired() {
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Use errgroup to send concurrently while handling errors
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	if msg.ClientID != "" {
		// Single client target
		if client, exists := s.clients[msg.ClientID]; exists && !client.IsClosed() {
			client := client // Capture for goroutine
			g.Go(func() error {
				return client.Send(ctx, msg)
			})
		}
	} else if msg.Channel != "" {
		// Channel target
		if clientsMap, exists := s.channels[msg.Channel]; exists {
			for _, client := range clientsMap {
				if client.IsClosed() {
					continue
				}
				client := client // Capture for goroutine
				g.Go(func() error {
					return client.Send(ctx, msg)
				})
			}
		}
	} else {
		// Broadcast to all
		for _, client := range s.clients {
			if client.IsClosed() {
				continue
			}
			client := client // Capture for goroutine
			g.Go(func() error {
				return client.Send(ctx, msg)
			})
		}
	}

	// Wait but don't block forever
	_ = g.Wait() // We can ignore errors here as they're per-client issues
}

// keepAliveRoutine sends periodic keep-alive messages to all clients
func (s *Server) keepAliveRoutine() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.sendKeepAlive()
		}
	}
}

// sendKeepAlive sends a keep-alive message to all clients
func (s *Server) sendKeepAlive() {
	s.mu.RLock()
	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		if !client.IsClosed() {
			clients = append(clients, client)
		}
	}
	s.mu.RUnlock()

	// Use context with timeout for keep-alive
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	// Use errgroup to send concurrently
	g, ctx := errgroup.WithContext(ctx)

	for _, client := range clients {
		client := client // Capture for goroutine
		g.Go(func() error {
			return client.SendKeepAlive(ctx)
		})
	}

	// Wait for all keep-alives to complete or timeout
	_ = g.Wait() // Ignore individual client errors
}

// Close shuts down the SSE server
func (s *Server) Close() error {
	// Signal all goroutines to stop
	s.cancel()

	// Wait for all routines to finish
	s.wg.Wait()

	// Close the broker
	if err := s.broker.Close(); err != nil {
		return fmt.Errorf("failed to close broker: %w", err)
	}

	return nil
}
