package sse

import (
	"context"
)

// Broker defines the interface for message distribution
// All SSE messages flow through the broker, which enables
// the system to scale horizontally across multiple instances.
type Broker interface {
	// Publish sends a message through the message bus
	// Messages can target specific clients, channels, or be broadcast to all.
	Publish(ctx context.Context, message Message) error
	
	// Subscribe returns a channel that receives all messages from the broker
	// The channel will be closed when the context is canceled or the broker is closed
	Subscribe(ctx context.Context) (<-chan Message, error)
	
	// Close terminates the broker connection and cleans up resources
	Close() error
}
