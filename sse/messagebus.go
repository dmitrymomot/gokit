package sse

import "context"

// MessageBus defines the interface for a message bus system used by the SSE server
// to distribute messages to connected clients. This interface allows for different
// implementations of message buses, such as in-memory channels, Redis, NATS, etc.
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
