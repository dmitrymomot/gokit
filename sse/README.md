# SSE Package

A flexible, scalable Server-Sent Events implementation for real-time web applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/sse
```

## Overview

The `sse` package provides a lightweight, thread-safe implementation of Server-Sent Events (SSE) that enables real-time communication from server to client. It supports both single-server and distributed architectures with pluggable message bus implementations, making it suitable for applications of any scale.

## Features

- Topic-based subscriptions for targeted event delivery
- Multiple message bus implementations (in-memory, Redis)
- Automatic client connection management with configurable heartbeats
- Support for structured data with automatic JSON serialization
- Compatible with any Go HTTP router or framework
- Thread-safe for concurrent operations

## Usage

### Basic Server Setup

```go
import (
    "net/http"
    "time"

    "github.com/dmitrymomot/gokit/sse"
    "github.com/dmitrymomot/gokit/sse/bus"
)

func main() {
    // Create message bus (in-memory for single server)
    msgBus := bus.NewChannelBus()

    // Create SSE server with 15-second heartbeat
    server := sse.NewServer(msgBus, sse.WithHeartbeat(15*time.Second))
    defer server.Close()

    // Set up SSE endpoint with topic extraction
    http.HandleFunc("/events", server.Handler(func(r *http.Request) string {
        return r.URL.Query().Get("topic") // Extract topic from query string
    }))

    // Set up publish endpoint
    http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        topic := r.URL.Query().Get("topic")
        message := r.URL.Query().Get("message")

        err := server.Publish(r.Context(), topic, sse.Event{
            Event: "message",
            Data:  message,
        })

        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
    })

    http.ListenAndServe(":8080", nil)
}
```

### Distributed Setup with Redis Message Bus

```go
import (
    "github.com/dmitrymomot/gokit/sse/bus"
    "github.com/redis/go-redis/v9"
)

// Initialize Redis client
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Create Redis-backed message bus with buffer size 200
msgBus, err := bus.NewRedisBusWithConfig(redisClient, 200)
if err != nil {
    log.Fatalf("Failed to create Redis message bus: %v", err)
}

// Create SSE server with Redis message bus
server := sse.NewServer(msgBus)
```

### Creating and Publishing Events

```go
// Create a simple text event
textEvent := sse.Event{
    Event: "message",
    Data:  "Hello, world!",
}

// Create a structured data event (automatically JSON-encoded)
dataEvent := sse.Event{
    ID:    "msg-123",         // Optional, generated if empty
    Event: "user_update",
    Data:  map[string]any{
        "id":        123,
        "username":  "johndoe",
        "status":    "online",
        "timestamp": time.Now(),
    },
    Retry: 3000,              // Reconnection time in ms (optional)
}

// Publish to a topic
err := server.Publish(ctx, "user:123", dataEvent)
if err != nil {
    // Handle error
    if errors.Is(err, sse.ErrTopicEmpty) {
        log.Println("Topic cannot be empty")
    } else if errors.Is(err, sse.ErrServerClosed) {
        log.Println("Server is closed")
    } else {
        log.Printf("Failed to publish event: %v", err)
    }
}
```

### Client-Side JavaScript

```javascript
// Connect to SSE endpoint
const eventSource = new EventSource("/events?topic=user:123");

// Listen for events
eventSource.addEventListener("user_update", (event) => {
    const userData = JSON.parse(event.data);
    console.log("User update:", userData);
    updateUserInterface(userData);
});

// Handle connection events
eventSource.addEventListener("open", () => {
    console.log("Connection established");
});

eventSource.addEventListener("error", (error) => {
    console.error("Connection error:", error);
    // EventSource will automatically try to reconnect
});

// Close connection when done (if needed)
// eventSource.close();
```

## Best Practices

1. **Topic Design**:
   - Use hierarchical topics (e.g., `user:123`, `chat:room1`) for organization
   - Keep topic names concise but descriptive
   - Consider using namespaces to prevent collisions

2. **Security and Authentication**:
   - Validate user authentication in your topic extractor function
   - Only allow clients to subscribe to topics they have access to
   - Sanitize and validate all topic names and event data

3. **Performance Optimization**:
   - Adjust message bus buffer sizes based on expected throughput
   - Keep event data compact; use notifications + REST for large data
   - Limit concurrent connections based on your server capacity
   - Choose heartbeat interval to balance connection stability and traffic

4. **Load Balancer Configuration**:
   - Use sticky sessions (client affinity) in production with multiple instances
   - Configure appropriate timeouts for long-lived connections
   - Use Redis message bus for distributed deployments

## API Reference

### Server

```go
// Create a new SSE server
func NewServer(bus MessageBus, opts ...ServerOption) *Server

// Server options
func WithHeartbeat(d time.Duration) ServerOption  // Set heartbeat interval (default: 30s)
func WithHostname(hostname string) ServerOption   // Set hostname for event IDs

// Server methods
func (s *Server) Handler(topicExtractor func(r *http.Request) string) http.HandlerFunc
func (s *Server) Publish(ctx context.Context, topic string, event Event) error
func (s *Server) Close() error
```

### Event Structure

```go
type Event struct {
    ID    string      // Event ID (auto-generated if empty)
    Event string      // Event type/name
    Data  any         // Event data (string, map, struct, etc.)
    Retry int         // Reconnection time in milliseconds (optional)
}

// Event methods
func (e Event) String() string
func (e Event) Write(w io.Writer) error
```

### Message Bus Interface

```go
type MessageBus interface {
    Publish(ctx context.Context, topic string, event Event) error
    Subscribe(ctx context.Context, topic string) (<-chan Event, error)
    Unsubscribe(ctx context.Context, topic string, ch <-chan Event) error
    Close() error
}
```

### Message Bus Implementations

```go
// In-memory (single server)
func NewChannelBus() MessageBus

// Redis-backed (distributed)
func NewRedisBus(redisClient *redis.Client) (MessageBus, error)
func NewRedisBusWithConfig(redisClient *redis.Client, bufferSize int) (MessageBus, error)
```

### Error Types

```go
var (
    ErrClientClosed    = errors.New("client is closed")
    ErrServerClosed    = errors.New("server is closed")
    ErrTopicEmpty      = errors.New("topic cannot be empty")
    ErrMessageEmpty    = errors.New("message cannot be empty")
    ErrInvalidEventID  = errors.New("invalid event ID")
    ErrMessageBusClosed = errors.New("message bus is closed")
    ErrNoFlusher       = errors.New("response writer does not implement http.Flusher")
)
```

## Examples

Full, working examples are available in the package:

- Chat application: `https://github.com/dmitrymomot/gokit/tree/main/sse/examples/chat`
- Real-time updates: `https://github.com/dmitrymomot/gokit/tree/main/sse/examples/realtime_update`
