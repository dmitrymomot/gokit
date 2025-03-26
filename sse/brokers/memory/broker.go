package memory

import (
	"context"
	"sync"

	"github.com/dmitrymomot/gokit/sse"
)

// Broker implements an in-memory broker for SSE messages
// This is useful for development and testing, or for
// single-instance deployments where scaling is not required.
type Broker struct {
	// Subscribers are channels that receive messages
	subscribers map[chan sse.Message]struct{}

	// Mutex for thread safety
	mu sync.RWMutex

	// Context and cancellation for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBroker creates a new in-memory broker
func NewBroker() (*Broker, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Broker{
		subscribers: make(map[chan sse.Message]struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Publish sends a message to all subscribers
func (b *Broker) Publish(ctx context.Context, message sse.Message) error {
	// Validate message
	if err := message.Validate(); err != nil {
		return err
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	// Check if broker is closed
	select {
	case <-b.ctx.Done():
		return sse.ErrBrokerClosed
	default:
		// Continue
	}

	// Send to all subscribers
	for ch := range b.subscribers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-b.ctx.Done():
			return sse.ErrBrokerClosed
		case ch <- message:
			// Message sent
		default:
			// Skip slow subscribers
		}
	}

	return nil
}

// Subscribe returns a channel that receives all messages
func (b *Broker) Subscribe(ctx context.Context) (<-chan sse.Message, error) {
	// Check if broker is closed
	select {
	case <-b.ctx.Done():
		return nil, sse.ErrBrokerClosed
	default:
		// Continue
	}

	// Create a buffered channel for this subscriber
	ch := make(chan sse.Message, 100)

	// Register subscriber
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	// Create a done channel to signal when this subscription ends
	done := make(chan struct{})

	// Start goroutine to handle context cancellation
	go func() {
		select {
		case <-ctx.Done():
			// Context canceled
		case <-b.ctx.Done():
			// Broker closed
		}

		// Unregister subscriber
		b.mu.Lock()
		delete(b.subscribers, ch)
		b.mu.Unlock()

		// Close the channel
		close(ch)
		close(done)
	}()

	return ch, nil
}

// Close terminates the broker and all subscriptions
func (b *Broker) Close() error {
	// Signal shutdown
	b.cancel()

	// Clear subscribers
	b.mu.Lock()
	b.subscribers = make(map[chan sse.Message]struct{})
	b.mu.Unlock()

	return nil
}
