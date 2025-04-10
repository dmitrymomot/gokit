# SSE Package

A flexible and scalable Server-Sent Events (SSE) implementation for Go applications.

## Overview

This package provides a lightweight, framework-agnostic implementation of Server-Sent Events (SSE) that can be used with any Go HTTP router. It's designed to be simple to use while supporting a wide range of use cases including:

- Feed updates
- Live chat applications
- Real-time metrics updates
- Notifications

The package is built with horizontal scaling in mind and supports different message bus implementations through an adapter interface.

## Architecture

The SSE package follows a clean architecture with the following key components:

```
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  HTTP Handler │────▶│   SSE Server  │────▶│  Message Bus  │
└───────────────┘     └───────────────┘     └───────────────┘
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  HTTP Clients │◀────│  SSE Clients  │◀────│Implementation │
└───────────────┘     └───────────────┘     └───────────────┘
```

- **Server**: Manages client connections and topic subscriptions
- **Client**: Represents a connected SSE client with event delivery capability
- **Event**: Data structure for SSE messages
- **MessageBus**: Interface for different message delivery systems
- **Bus Implementations**: Currently includes Channel (in-memory) and Redis options

## Features

- Compatible with any Go HTTP router
- Topic-based subscription system
- Pluggable message bus architecture
- Built-in implementations:
    - Channel-based in-memory bus (for single server deployments)
    - Redis-backed distributed bus (for horizontal scaling)
- Automatic client disconnection handling
- Configurable heartbeat mechanism to maintain connections
- Simple, clean API

## Installation

```
go get github.com/dmitrymomot/gokit/sse
```

## API Reference

### Server

```go
// Create a new SSE server with a message bus
sseServer := sse.NewServer(messageBus, [options...])

// Available options
sse.WithHeartbeat(duration)  // Set heartbeat interval (default: 30s)
sse.WithHostname(hostname)   // Set hostname for event IDs

// Get an HTTP handler for SSE connections
handler := sseServer.Handler(topicExtractorFunc)

// Publish an event to a topic
err := sseServer.Publish(ctx, topic, event)

// Close the server and all connections
err := sseServer.Close()
```

### Event

```go
// Create a new event
event := sse.Event{
    ID:    "unique-id",        // Optional: Auto-generated if empty
    Event: "event-type",      // The event type (e.g., "message", "update")
    Data:  data,              // Any data type (string, map, struct, etc.)
    Retry: 3000,              // Optional: Reconnection time in milliseconds
}
```

### Message Bus Implementations

```go
// In-memory channel-based message bus (for single server)
msgBus := bus.NewChannelBus()

// Redis-based message bus (for distributed setup)
redisClient := redis.NewClient(&redis.Options{...})
msgBus, err := bus.NewRedisBus(redisClient)
// OR with custom buffer size
msgBus, err := bus.NewRedisBusWithConfig(redisClient, 200) // Buffer size of 200
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

	// Create an SSE server with custom heartbeat
	sseServer := sse.NewServer(msgBus, sse.WithHeartbeat(15*time.Second))

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
			log.Printf("Error publishing message: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Set up graceful shutdown
	go func() {
		// Listen for termination signal
		// ...
		// Close SSE server
		sseServer.Close()
	}()

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
			log.Printf("Error publishing feed update: %v", err)
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
			log.Printf("Error publishing chat message: %v", err)
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

	// Create a cancellable context for the metrics publisher
	ctx, cancel := context.WithCancel(context.Background())
	// Store cancel function for cleanup on server shutdown

	// Periodically publish metrics updates
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Context canceled, stop publishing
				return
			case <-ticker.C:
				// Get users who are currently online
				userIDs := getActiveUserIDs() // Implement this based on your system

				// Update metrics for each active user
				for _, userID := range userIDs {
					// Get metrics data for this user
					metricsData := generateMetricsData(userID) // Implement this based on your system

					// Publish metrics to the user's topic
					topic := "metrics:" + userID
					err := sseServer.Publish(ctx, topic, sse.Event{
						Event: "metrics_update",
						Data:  metricsData,
					})
					if err != nil {
						log.Printf("Error publishing metrics for user %s: %v", userID, err)
					}
				}
			}
		}
	}()
}
```

## Error Handling

The SSE package provides several predefined errors that you can check against:

```go
// Common errors
sse.ErrClientClosed    // When trying to send to a closed client
sse.ErrServerClosed    // When trying to use a closed server
sse.ErrTopicEmpty      // When the topic is empty
sse.ErrMessageEmpty    // When the message is empty
sse.ErrInvalidEventID  // When the event ID is invalid
sse.ErrMessageBusClosed // When the message bus is closed
sse.ErrNoFlusher       // When the ResponseWriter doesn't implement http.Flusher
```

Example of proper error handling:

```go
// Publishing with error handling
err := sseServer.Publish(ctx, topic, event)
if err != nil {
    if errors.Is(err, sse.ErrTopicEmpty) {
        log.Println("Topic cannot be empty")
        return
    } else if errors.Is(err, sse.ErrServerClosed) {
        log.Println("Server is closed, cannot publish")
        return
    }
    log.Printf("Unexpected error publishing event: %v", err)
    return
}
```

## Testing

Here's an example of testing an SSE endpoint using Go's testing package:

```go
func TestSSEEndpoint(t *testing.T) {
	// Create a test message bus
	msgBus := bus.NewChannelBus()

	// Create an SSE server
	sseServer := sse.NewServer(msgBus)

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(sseServer.Handler(func(r *http.Request) string {
		return "test-topic"
	})))
	defer ts.Close()

	// Prepare a test event
	testEvent := sse.Event{
		Event: "test-event",
		Data:  "test-data",
	}

	// Publish the event after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		err := sseServer.Publish(context.Background(), "test-topic", testEvent)
		if err != nil {
			t.Errorf("Failed to publish event: %v", err)
		}
	}()

	// Create a custom client that doesn't follow redirects
	client := &http.Client{
		Timeout: 1 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Make the request
	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check response headers
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("Expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}

	// Read and parse the events
	scanner := bufio.NewScanner(resp.Body)
	eventData := ""
	eventType := ""

	// Simple SSE parser
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// Event is complete
			if eventType == "test-event" && eventData == "test-data" {
				// Success
				return
			}
			eventData = ""
			eventType = ""
		} else if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			eventData = strings.TrimPrefix(line, "data: ")
		}
	}

	if scanner.Err() != nil {
		t.Fatalf("Scanner error: %v", scanner.Err())
	}

	t.Fatal("Did not receive expected event")
}
```

## Implementing Custom Message Bus Adapters

You can implement custom message bus adapters for any backend by implementing the `MessageBus` interface:

```go
type MessageBus interface {
    // Publish sends a message to a specific topic
    Publish(ctx context.Context, topic string, event Event) error

    // Subscribe returns a channel that receives events for a specific topic
    Subscribe(ctx context.Context, topic string) (<-chan Event, error)

    // Unsubscribe removes a subscription for a specific topic
    Unsubscribe(ctx context.Context, topic string, ch <-chan Event) error

    // Close shuts down the message bus
    Close() error
}
```

Here's a simplified example for a NATS-based message bus adapter:

```go
// NATSBus implements the MessageBus interface using NATS
type NATSBus struct {
    conn      *nats.Conn
    mu        sync.RWMutex
    subs      map[string][]*nats.Subscription
    channels  map[*nats.Subscription]chan sse.Event
    closed    bool
}

func NewNATSBus(url string) (*NATSBus, error) {
    // Connect to NATS server
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, err
    }

    return &NATSBus{
        conn:     nc,
        subs:     make(map[string][]*nats.Subscription),
        channels: make(map[*nats.Subscription]chan sse.Event),
        closed:   false,
    }, nil
}

// Implement MessageBus interface methods...
```

## Horizontal Scaling

To scale the SSE server horizontally:

1. **Use a distributed message bus**: Implement or use a message bus adapter that works across multiple servers:

    - The built-in Redis adapter (`bus.NewRedisBus`) works well for this purpose
    - Alternative options: NATS, RabbitMQ, Kafka, etc.

2. **Deploy multiple application instances**: Each with its own SSE server connected to the same distributed message bus

3. **Load balancing considerations**:

    - Use sticky sessions (client affinity) to ensure clients stay connected to the same server instance
    - Configure proper timeouts for the load balancer to avoid disconnecting long-lived SSE connections
    - Example nginx configuration:

    ```nginx
    http {
        upstream app_servers {
            ip_hash;  # Enable sticky sessions
            server app1:8080;
            server app2:8080;
        }

        server {
            listen 80;

            location /events {
                proxy_pass http://app_servers;
                proxy_http_version 1.1;
                proxy_set_header Connection "";
                proxy_buffering off;
                proxy_read_timeout 3600s;
                proxy_send_timeout 3600s;
            }
        }
    }
    ```

## Client-Side Usage

In the browser, connect to the SSE endpoint using the EventSource API:

```javascript
// Connect to the SSE endpoint
const eventSource = new EventSource("/events?topic=my-topic");

// Listen for specific event types
eventSource.addEventListener("message", (event) => {
    console.log("Received message:", event.data);
});

eventSource.addEventListener("chat_message", (event) => {
    const message = JSON.parse(event.data);
    console.log(`${message.username}: ${message.message}`);
});

// Handle connection established
eventSource.addEventListener("connected", (event) => {
    console.log("Connected to event stream:", event.data);
});

// Handle errors
eventSource.addEventListener("error", (error) => {
    console.error("SSE connection error:", error);
    // Implement reconnection logic if needed
    if (eventSource.readyState === EventSource.CLOSED) {
        // Connection was closed, reconnect after a delay
        setTimeout(() => {
            // Create a new EventSource instance
            // ...
        }, 3000);
    }
});

// Close the connection when done
// eventSource.close();
```

## Performance Considerations

- **Buffer sizing**: Adjust channel buffer sizes based on expected message volume

    - Default is 100 events for both ChannelBus and RedisBus
    - Increase for high-throughput applications

- **Goroutine management**: The server creates goroutines for each client connection

    - Make sure to properly close the server to clean up resources
    - For very high numbers of concurrent clients, monitor goroutine count

- **Message serialization**: Messages are JSON-serialized when using the Redis adapter

    - Keep event data structures efficient and avoid large payloads
    - Consider compression for large messages

- **Heartbeat interval**: Configure appropriate heartbeat intervals
    - Too frequent: Increased network traffic and CPU usage
    - Too infrequent: Risk of proxy timeouts closing idle connections

## Security Considerations

- **Authentication**: SSE doesn't support custom headers for authentication

    - Use query parameters or cookies for authentication
    - Validate authentication in the topic extractor function

- **Topic authorization**: Implement checks in the topic extractor function

    ```go
    sseServer.Handler(func(r *http.Request) string {
        userID := getUserFromSession(r)
        requestedTopic := r.URL.Query().Get("topic")

        // Check if user is authorized for this topic
        if !isAuthorized(userID, requestedTopic) {
            return "" // Empty topic means reject the connection
        }

        return requestedTopic
    })
    ```

- **Input validation**: Always validate topic names and event data

    - Prevent injection attacks in topic names
    - Sanitize user-provided content before publishing

- **Rate limiting**: Implement rate limiting for publishing endpoints
    - Prevent flooding the message bus
    - Protect against denial-of-service attacks

## Considerations

- **Connection Limits**: Browsers typically limit the number of concurrent connections to a domain (usually 6), which can affect SSE scalability. Consider domain sharding for many parallel SSE connections.
- **Proxy Support**: Some proxies don't support long-lived connections; set appropriate timeouts and use HTTP/1.1 with proper headers.
- **Reconnection**: The EventSource API has built-in reconnection logic, but you can customize it using the `retry` field.
- **Message Size**: Keep SSE messages small to avoid buffering issues. For large data transfers, consider sending a notification and having the client fetch the data separately.
- **Back-Pressure**: Implement throttling if clients can't keep up with message rates.
- **Memory Management**: Monitor memory usage when dealing with many concurrent connections.
