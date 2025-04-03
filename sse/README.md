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

The `Message` struct represents an SSE event with fields for ID, event type, data payload, and targeting information. Messages can also have a TTL (Time-To-Live) to control their expiration.

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

// Send a message with TTL (Time-To-Live)
func sendWithTTL(ctx context.Context, broker sse.Broker) error {
	// Create a message that expires after 30 seconds
	msg := sse.NewMessage("time-sensitive", map[string]string{
		"message": "This message will expire!",
	}).WithTTL(30 * time.Second)
	
	return broker.Publish(ctx, msg)
}

// Create message with TTL directly in constructor
func createExpiringMessage(data any) sse.Message {
	// Alternative way to set TTL directly in constructor
	return sse.NewMessage("expiring-event", data, 5 * time.Minute)
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

## Broker Implementations

### Memory Broker

The memory broker is ideal for single-server deployments or for development and testing. It's simple to set up and use, but doesn't support horizontal scaling.

```go
import "github.com/dmitrymomot/gokit/sse/brokers/memory"

broker, err := memory.NewBroker()
if err != nil {
    // Handle error
}
```

### Redis Broker

The Redis broker enables horizontal scaling by using Redis pub/sub for message distribution. This allows multiple server instances to work together, with all messages being propagated to all connected clients.

```go
import (
    "github.com/dmitrymomot/gokit/sse/brokers/redis"
    redisClient "github.com/redis/go-redis/v9"
)

// Create Redis client
client := redisClient.NewClient(&redisClient.Options{
    Addr: "localhost:6379",
})

// Create broker with default options
broker, err := redis.NewBroker(client)
if err != nil {
    // Handle error
}

// Or with custom options (optional)
broker, err := redis.NewBroker(client, redis.Options{
    Channel: "custom_sse_channel",
})
```

### Redis Streams Broker

The Redis Streams broker provides enhanced reliability and automatic message expiration using Redis Streams instead of pub/sub. This is the recommended broker for production deployments with high message volumes or when message accumulation is a concern.

```go
import (
    "github.com/dmitrymomot/gokit/sse/brokers/redis"
    redisClient "github.com/redis/go-redis/v9"
)

// Create Redis client
client := redisClient.NewClient(&redisClient.Options{
    Addr: "localhost:6379",
})

// Create broker with default options
broker, err := redis.NewStreamsBroker(client)
if err != nil {
    // Handle error
}

// Or with custom options (for more control)
broker, err := redis.NewStreamsBroker(client, redis.StreamsOptions{
    StreamName:       "sse_messages",         // Redis stream name
    MaxStreamLength:  1000,                   // Maximum number of messages to keep
    MessageRetention: 24 * time.Hour,         // How long to keep messages
    GroupName:        "sse_consumer_group",   // Consumer group name
    ConsumerName:     "instance-1",           // Consumer name (unique per instance)
    BlockDuration:    100 * time.Millisecond, // Polling interval
})
```

Key benefits of the Streams broker:

- **Automatic message expiration**: Old messages are automatically removed based on the configured `MaxStreamLength` and `MessageRetention`
- **Reliable message delivery**: Uses Redis Streams consumer groups for reliable message processing
- **No message accumulation**: Messages are capped to the specified maximum stream length
- **Efficient retrieval**: Only gets new messages, with automatic acknowledgment
- **Manual cleanup**: Additional `CleanupStream` method available for explicit stream maintenance

For periodic cleanup of the stream (optional):

```go
// Set up periodic cleanup if needed
ticker := time.NewTicker(1 * time.Hour)
go func() {
    for {
        select {
        case <-ticker.C:
            if streamsBroker, ok := broker.(*redis.StreamsBroker); ok {
                ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
                _ = streamsBroker.CleanupStream(ctx)
                cancel()
            }
        case <-ctx.Done():
            ticker.Stop()
            return
        }
    }
}()
