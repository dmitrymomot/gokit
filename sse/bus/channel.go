package bus

import (
	"context"
	"sync"

	"github.com/dmitrymomot/gokit/sse"
)

// ChannelBus implements the MessageBus interface using Go channels
type ChannelBus struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan sse.Event]struct{}
	closed      bool
}

// NewChannelBus creates a new ChannelBus
func NewChannelBus() *ChannelBus {
	return &ChannelBus{
		subscribers: make(map[string]map[chan sse.Event]struct{}),
	}
}

// Publish sends a message to all subscribers of a topic
func (b *ChannelBus) Publish(ctx context.Context, topic string, event sse.Event) error {
	if topic == "" {
		return sse.ErrTopicEmpty
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return sse.ErrMessageBusClosed
	}

	subs, exists := b.subscribers[topic]
	if !exists || len(subs) == 0 {
		return nil // No subscribers for this topic
	}

	// Use a goroutine to avoid blocking and handle slow subscribers
	for ch := range subs {
		ch := ch // Capture variable for the goroutine
		go func() {
			// Use a non-blocking send to avoid deadlocks
			select {
			case ch <- event:
				// Successfully sent
			case <-ctx.Done():
				// Context canceled
			default:
				// Channel is full or closed, will be cleaned up on future Subscribe/Unsubscribe
			}
		}()
	}

	return nil
}

// Subscribe returns a channel that receives events for a specific topic
func (b *ChannelBus) Subscribe(ctx context.Context, topic string) (<-chan sse.Event, error) {
	if topic == "" {
		return nil, sse.ErrTopicEmpty
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, sse.ErrMessageBusClosed
	}

	// Create a buffered channel to avoid blocking publishers
	ch := make(chan sse.Event, 100)

	// Create the topic if it doesn't exist
	if _, exists := b.subscribers[topic]; !exists {
		b.subscribers[topic] = make(map[chan sse.Event]struct{})
	}

	// Add the channel to the subscribers for this topic
	b.subscribers[topic][ch] = struct{}{}

	// Clean up when context is done
	go func() {
		<-ctx.Done()
		b.Unsubscribe(context.Background(), topic, ch)
	}()

	return ch, nil
}

// Unsubscribe removes a subscription for a specific topic
func (b *ChannelBus) Unsubscribe(ctx context.Context, topic string, ch <-chan sse.Event) error {
	if topic == "" {
		return sse.ErrTopicEmpty
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return sse.ErrMessageBusClosed
	}

	// Check if the topic exists
	subs, exists := b.subscribers[topic]
	if !exists {
		return nil // Topic doesn't exist, nothing to unsubscribe
	}

	// Find and remove the channel
	for subCh := range subs {
		if ch == subCh {
			delete(subs, subCh)
			close(subCh) // Close the channel
			break
		}
	}

	// Remove the topic if there are no more subscribers
	if len(subs) == 0 {
		delete(b.subscribers, topic)
	}

	return nil
}

// Close shuts down the message bus
func (b *ChannelBus) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil // Already closed
	}

	b.closed = true

	// Close all subscriber channels
	for _, subs := range b.subscribers {
		for ch := range subs {
			close(ch)
		}
	}

	// Clear the subscribers map
	b.subscribers = nil

	return nil
}