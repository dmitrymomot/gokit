# SSE Package

A flexible, scalable Server-Sent Events implementation for real-time web applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/sse
```

## Overview

The `sse` package provides a lightweight, framework-agnostic implementation of Server-Sent Events (SSE) that enables real-time communication from server to client. It supports both single-server and distributed architectures with pluggable message bus implementations.

## Features

- Clean, modular architecture with clear separation of concerns
- Topic-based subscriptions for targeted event delivery
- Multiple message bus implementations:
    - In-memory channel bus for single-server deployments
    - Redis-backed bus for distributed environments
- Configurable heartbeats to maintain long-lived connections
- Automatic client connection management
- Horizontal scaling support
- Compatible with any Go HTTP router

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

### With Redis Message Bus (for distributed setups)

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

## API Reference

### Server

```go
// Create a new SSE server
server := sse.NewServer(messageBus, [options...])

// Available options
sse.WithHeartbeat(duration)  // Set heartbeat interval (default: 30s)
sse.WithHostname(hostname)   // Set hostname for event IDs

// Get an HTTP handler for SSE connections
handler := server.Handler(topicExtractorFunc)

// Publish an event to a topic
err := server.Publish(ctx, topic, event)

// Close the server
err := server.Close()
```

### Event Structure

```go
type Event struct {
    ID    string      // Event ID (auto-generated if empty)
    Event string      // Event type/name
    Data  any         // Event data (string, map, struct, etc.)
    Retry int         // Reconnection time in milliseconds (optional)
}
```

### Message Bus Implementations

```go
// In-memory (single server)
bus := bus.NewChannelBus()

// Redis-backed (distributed)
bus, err := bus.NewRedisBus(redisClient)
// or with custom buffer size:
bus, err := bus.NewRedisBusWithConfig(redisClient, 200)
```

## Error Handling

The package provides predefined errors for specific failure cases:

```go
// Common errors to check for
if errors.Is(err, sse.ErrTopicEmpty) {
    // Handle empty topic error
}
if errors.Is(err, sse.ErrServerClosed) {
    // Handle server closed error
}
if errors.Is(err, sse.ErrClientClosed) {
    // Handle client closed error
}
```

## Scaling and Production Considerations

### Load Balancer Configuration

For production deployments with multiple instances:

1. Use the Redis message bus implementation
2. Configure sticky sessions (client affinity) in your load balancer
3. Set appropriate timeouts for long-lived connections

Example nginx configuration:

```nginx
upstream sse_servers {
    ip_hash;  // Enable sticky sessions
    server app1:8080;
    server app2:8080;
}

server {
    location /events {
        proxy_pass http://sse_servers;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_read_timeout 3600s;
    }
}
```

### Security Best Practices

1. **Authentication**: Validate user authentication in your topic extractor function
2. **Authorization**: Only allow clients to subscribe to topics they have access to
3. **Input validation**: Sanitize and validate all topic names and event data
4. **Rate limiting**: Implement rate limiting for publish endpoints

### Performance Optimization

1. **Buffer sizing**: Adjust message bus buffer sizes based on expected throughput
2. **Message size**: Keep event data compact; use notifications + REST for large data
3. **Connection pooling**: Monitor and limit concurrent connections
4. **Heartbeat interval**: Balance between connection stability and network traffic

## Examples

Full, working examples are available in the package:

- Chat application: `https://github.com/dmitrymomot/gokit/tree/main/sse/examples/chat`
- Real-time updates: `https://github.com/dmitrymomot/gokit/tree/main/sse/examples/realtime_update`
