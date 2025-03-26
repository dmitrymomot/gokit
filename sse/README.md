# SSE (Server-Sent Events)

The SSE package provides a robust implementation of the [Server-Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) protocol for Go applications. This package enables real-time, one-way communication from server to client over a standard HTTP connection.

## Features

- Implements the `http.Handler` interface for seamless integration with Go's HTTP server
- Broker-centric architecture for message distribution, enabling horizontal scaling
- Support for client-specific, channel-based, and global message broadcasting
- Built-in keep-alive mechanism to maintain connections
- Clean separation between connection handling and message distribution
- Thread-safe operations for concurrent use

## Installation

```bash
go get github.com/dmitrymomot/gokit/sse
```

## Core Components

### Server

The `Server` handles client connections and message dispatching. It implements the `http.Handler` interface for easy integration with standard Go HTTP routers.

### Broker

The `Broker` interface defines the contract for message distribution. All messages flow through the broker, which enables scaling across multiple server instances.

### Message

The `Message` struct represents an SSE event with fields for ID, event type, data payload, and targeting information.

## Usage Examples

### Basic Server Setup

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	
	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/brokers/memory"
)

func main() {
	// Create a broker (in-memory for this example)
	broker, err := memory.NewBroker()
	if err != nil {
		slog.Error("Failed to create broker", "error", err)
		return
	}
	defer broker.Close()
	
	// Create SSE server
	server, err := sse.NewServer(broker)
	if err != nil {
		slog.Error("Failed to create SSE server", "error", err)
		return
	}
	defer server.Close()
	
	// Register SSE endpoint
	http.Handle("/events", server)
	
	// Start HTTP server
	slog.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("Server failed", "error", err)
	}
}
```

### Publishing Messages

```go
// Send to specific client
func sendToClient(ctx context.Context, broker sse.Broker, clientID string) error {
	msg := sse.NewMessage("notification", map[string]string{
		"message": "Hello, specific client!",
	}).ForClient(clientID)
	
	return broker.Publish(ctx, msg)
}

// Broadcast to channel
func broadcastToChannel(ctx context.Context, broker sse.Broker, channel string) error {
	msg := sse.NewMessage("update", map[string]string{
		"message": "Channel update!",
	}).ForChannel(channel)
	
	return broker.Publish(ctx, msg)
}

// Broadcast to all clients
func broadcastToAll(ctx context.Context, broker sse.Broker) error {
	msg := sse.NewMessage("alert", map[string]string{
		"message": "System-wide alert!",
	})
	
	return broker.Publish(ctx, msg)
}
```

### Client-Side JavaScript

```javascript
// Connect to SSE endpoint
const clientID = "user-123";
const channel = "updates";
const eventSource = new EventSource(`/events?client_id=${clientID}&channel=${channel}`);

// Listen for specific events
eventSource.addEventListener('notification', (e) => {
    const data = JSON.parse(e.data);
    console.log('Notification:', data);
});

eventSource.addEventListener('update', (e) => {
    const data = JSON.parse(e.data);
    console.log('Update:', data);
});

eventSource.addEventListener('alert', (e) => {
    const data = JSON.parse(e.data);
    console.log('Alert:', data);
});

// Listen for all events
eventSource.onmessage = (e) => {
    console.log('Generic message:', e.data);
};

// Handle connection events
eventSource.onopen = () => {
    console.log('Connection established');
};

eventSource.onerror = (e) => {
    console.error('Connection error:', e);
    // Reconnect logic
};

// Close connection when needed
function closeConnection() {
    eventSource.close();
}
```

### Custom Broker Implementation

To implement your own broker (e.g., for Redis, RabbitMQ, etc.):

```go
// Implement the Broker interface
type RedisBroker struct {
    // Redis client
    client redis.Client
    // Other fields
}

// Implement Publish method
func (b *RedisBroker) Publish(ctx context.Context, message sse.Message) error {
    // Publish message to Redis
    // ...
}

// Implement Subscribe method
func (b *RedisBroker) Subscribe(ctx context.Context) (<-chan sse.Message, error) {
    // Subscribe to Redis channel
    // ...
}

// Implement Close method
func (b *RedisBroker) Close() error {
    // Close Redis connection
    // ...
}
```

## Thread Safety

All operations in the SSE package are thread-safe and can be used concurrently from multiple goroutines.

## Scalability

The broker-centric design allows for horizontal scaling across multiple server instances. By implementing a distributed broker (e.g., using Redis, NATS, Kafka), messages can be propagated to all connected clients regardless of which server instance they're connected to.
