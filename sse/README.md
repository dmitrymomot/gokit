# SSE Package

A flexible and scalable Server-Sent Events (SSE) implementation for Go applications.

## Overview

This package provides a lightweight, framework-agnostic implementation of Server-Sent Events (SSE) that can be used with any Go HTTP router. It's designed to be simple to use while supporting a wide range of use cases including:

- Feed updates
- Live chat applications
- Real-time metrics updates
- Notifications

The package is built with horizontal scaling in mind and supports different message bus implementations through an adapter interface.

## Features

- Compatible with any Go HTTP router
- Topic-based subscription system
- Pluggable message bus architecture
- Built-in channel-based message bus implementation
- Automatic client disconnection handling
- Heartbeat mechanism to maintain connections
- Simple, clean API

## Installation

```
go get github.com/dmitrymomot/gokit/sse
```

## Usage

### Basic Example

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/bus"
)

func main() {
	// Create a channel-based message bus
	msgBus := bus.NewChannelBus()

	// Create an SSE server
	sseServer := sse.NewServer(msgBus)

	// Create an HTTP server
	http.HandleFunc("/events", sseServer.Handler(func(r *http.Request) string {
		// Extract topic from request, e.g., from query parameter
		return r.URL.Query().Get("topic")
	}))

	// Create an endpoint to publish events
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		topic := r.URL.Query().Get("topic")
		message := r.URL.Query().Get("message")

		err := sseServer.Publish(r.Context(), topic, sse.Event{
			Event: "message",
			Data:  message,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Feed Updates Example

```go
func setupFeedUpdates(sseServer *sse.Server) {
	// Handler for subscribing to feed updates
	http.HandleFunc("/feeds/subscribe", sseServer.Handler(func(r *http.Request) string {
		// Get feed ID from path or query parameters
		feedID := r.URL.Query().Get("feed_id")
		// Return the topic, e.g., "feed:{feedID}"
		return "feed:" + feedID
	}))

	// Handler for publishing new feed items
	http.HandleFunc("/feeds/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse the request body to get feed update
		// ...

		// Publish to the appropriate feed topic
		feedID := r.URL.Query().Get("feed_id")
		topic := "feed:" + feedID

		err := sseServer.Publish(r.Context(), topic, sse.Event{
			Event: "feed_update",
			Data:  "New feed item data", // Replace with actual feed data
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
```

### Chat Example

```go
func setupChat(sseServer *sse.Server) {
	// Handler for subscribing to a chat room
	http.HandleFunc("/chat/subscribe", sseServer.Handler(func(r *http.Request) string {
		// Get chat room ID from path or query parameters
		roomID := r.URL.Query().Get("room_id")
		// Return the topic, e.g., "chat:{roomID}"
		return "chat:" + roomID
	}))

	// Handler for sending chat messages
	http.HandleFunc("/chat/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Parse the request to get chat message
		roomID := r.URL.Query().Get("room_id")
		username := r.URL.Query().Get("username")
		message := r.URL.Query().Get("message")

		// Create a chat message in JSON format
		chatData := fmt.Sprintf(`{"username":"%s","message":"%s","timestamp":"%s"}`,
			username,
			message,
			time.Now().Format(time.RFC3339),
		)

		// Publish to the chat room topic
		topic := "chat:" + roomID
		err := sseServer.Publish(r.Context(), topic, sse.Event{
			Event: "chat_message",
			Data:  chatData,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
```

### Dashboard Metrics Example

```go
func setupDashboardMetrics(sseServer *sse.Server) {
	// Handler for subscribing to dashboard metrics
	http.HandleFunc("/metrics/subscribe", sseServer.Handler(func(r *http.Request) string {
		// Get user ID from request (e.g., from session or JWT)
		userID := getUserIDFromRequest(r) // Implement this function based on your auth system
		// Return the topic, e.g., "metrics:{userID}"
		return "metrics:" + userID
	}))

	// Periodically publish metrics updates
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Get users who are currently online
				userIDs := getActiveUserIDs() // Implement this based on your system

				// Update metrics for each active user
				for _, userID := range userIDs {
					// Get metrics data for this user
					metricsData := generateMetricsData(userID) // Implement this based on your system

					// Publish metrics to the user's topic
					topic := "metrics:" + userID
					sseServer.Publish(context.Background(), topic, sse.Event{
						Event: "metrics_update",
						Data:  metricsData,
					})
				}
			}
		}
	}()
}
```

## Implementing Custom Message Bus Adapters

To scale horizontally, you can implement custom message bus adapters. Here's an example of how to implement a Redis-based message bus:

```go
// Example pseudo-code for a Redis-based message bus implementation
type RedisBus struct {
	client *redis.Client
	// ... other fields
}

func NewRedisBus(redisURL string) (*RedisBus, error) {
	// Initialize Redis client
	// ...
	return &RedisBus{client: client}, nil
}

// Implement the MessageBus interface methods
func (b *RedisBus) Publish(ctx context.Context, topic string, event sse.Event) error {
	// Serialize the event
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish to Redis channel
	return b.client.Publish(ctx, topic, data).Err()
}

func (b *RedisBus) Subscribe(ctx context.Context, topic string) (<-chan sse.Event, error) {
	// Create a pubsub subscription
	pubsub := b.client.Subscribe(ctx, topic)

	// Create output channel
	ch := make(chan sse.Event, 100)

	// Handle messages in a goroutine
	go func() {
		defer close(ch)
		defer pubsub.Close()

		// Listen for messages
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pubsub.Channel():
				// Deserialize the event
				var event sse.Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					// Handle error
					continue
				}

				// Send event to the channel
				select {
				case ch <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (b *RedisBus) Unsubscribe(ctx context.Context, topic string, ch <-chan sse.Event) error {
	// Implementation depends on how you track subscriptions
	// Typically, this would involve context cancellation
	return nil
}

func (b *RedisBus) Close() error {
	return b.client.Close()
}
```

## Horizontal Scaling

To scale the SSE server horizontally:

1. Implement a distributed message bus adapter (like the Redis example above)
2. Deploy multiple instances of your application
3. Use a load balancer with sticky sessions (to maintain client connections)
4. Ensure each instance uses the same message bus configuration

This approach allows events published by any instance to be received by all connected clients, regardless of which instance they're connected to.

## Client-Side Usage

In the browser, connect to the SSE endpoint using the EventSource API:

```javascript
// Connect to the SSE endpoint
const eventSource = new EventSource('/events?topic=my-topic');

// Listen for specific event types
eventSource.addEventListener('message', (event) => {
  console.log('Received message:', event.data);
});

eventSource.addEventListener('chat_message', (event) => {
  const message = JSON.parse(event.data);
  console.log(`${message.username}: ${message.message}`);
});

// Handle connection established
eventSource.addEventListener('connected', (event) => {
  console.log('Connected to event stream:', event.data);
});

// Handle errors
eventSource.addEventListener('error', (error) => {
  console.error('SSE connection error:', error);
  // Reconnect logic if needed
});

// Close the connection when done
// eventSource.close();
```

## Considerations

- **Connection Limits**: Browsers typically limit the number of concurrent connections to a domain, which can affect SSE scalability
- **Proxy Support**: Some proxies don't support long-lived connections; set appropriate timeouts
- **Reconnection**: Implement client-side reconnection logic for robustness
- **Message Size**: Keep SSE messages small to avoid buffering issues

## License

MIT