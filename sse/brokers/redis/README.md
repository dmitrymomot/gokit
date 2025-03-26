# Redis Broker for SSE

This package provides a Redis-backed broker implementation for Server-Sent Events (SSE). It enables horizontal scaling of SSE servers by using Redis pub/sub as the message distribution mechanism across multiple instances.

## Overview

The Redis broker implements the `sse.Broker` interface and uses Redis pub/sub to distribute messages across server instances. This allows for a scalable SSE implementation where messages published from one server instance can be received by clients connected to other instances.

## Features

- Horizontal scaling of SSE servers
- Fault tolerance through Redis pub/sub
- Configurable pub/sub channel
- Buffered message delivery
- Automatic handling of slow subscribers

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/sse"
	redisbroker "github.com/dmitrymomot/gokit/sse/brokers/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Create Redis broker for SSE
	broker, err := redisbroker.NewBroker(redisClient)
	if err != nil {
		log.Fatalf("Failed to create Redis broker: %v", err)
	}
	defer broker.Close()

	// Create SSE server with Redis broker
	server := sse.NewServer(broker)
	defer server.Close()

	// Set up HTTP handler for SSE connections
	http.HandleFunc("/events", server.ServeHTTP)

	// Start HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

### Publishing Messages

```go
// Function to publish a message from anywhere in your application
func publishMessage(ctx context.Context, broker sse.Broker) error {
	// Create a new message
	message := sse.NewMessage("user_updated", map[string]interface{}{
		"id":   123,
		"name": "John Doe",
	})

	// Publish to all clients
	return broker.Publish(ctx, message)

	// Or publish to a specific channel
	// return broker.Publish(ctx, message.ForChannel("user_updates"))

	// Or publish to a specific client
	// return broker.Publish(ctx, message.ForClient("client-123"))
}
```

### Custom Configuration

```go
// Create Redis broker with custom options
broker, err := redisbroker.NewBroker(
	redisClient,
	redisbroker.Options{
		Channel: "custom_sse_channel", // Use custom Redis channel
	},
)
```

### Using with Redis Cluster

```go
// Create Redis cluster client
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
	Addrs: []string{
		"redis-node1:6379",
		"redis-node2:6379",
		"redis-node3:6379",
	},
})

// Create Redis broker with cluster client
broker, err := redisbroker.NewBroker(clusterClient)
```

## Error Handling

The Redis broker defines the following errors:

- `ErrNoRedisClient`: Returned when attempting to create a broker without a Redis client
- `ErrRedisConnectionFailed`: Returned when the broker cannot connect to Redis

Additionally, it may return errors from the SSE package:

- `sse.ErrBrokerClosed`: Returned when attempting to use a closed broker
- `sse.ErrInvalidMessage`: Returned when a message is invalid

## Thread Safety

The Redis broker implementation is thread-safe and can be safely used from multiple goroutines.
