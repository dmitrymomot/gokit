package bus

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/dmitrymomot/gokit/sse"
)

// channelEntry represents a channel subscription with active state
type channelEntry struct {
	ch     chan sse.Event
	active atomic.Bool
}

// ChannelBus implements the MessageBus interface using Go channels
type ChannelBus struct {
	mu          sync.RWMutex
	subscribers map[string][]*channelEntry
	closed      bool
}

// NewChannelBus creates a new ChannelBus
func NewChannelBus() *ChannelBus {
	return &ChannelBus{
		subscribers: make(map[string][]*channelEntry),
	}
}

// Publish sends a message to all subscribers of a topic
func (b *ChannelBus) Publish(ctx context.Context, topic string, event sse.Event) error {
	if topic == "" {
		return sse.ErrTopicEmpty
	}

	b.mu.RLock()
	subs, exists := b.subscribers[topic]
	closed := b.closed
	b.mu.RUnlock()

	if closed {
		return sse.ErrMessageBusClosed
	}

	if !exists || len(subs) == 0 {
		return nil // No subscribers for this topic
	}

	// Create a goroutine for each active subscriber
	for _, entry := range subs {
		if !entry.active.Load() {
			// Skip inactive subscriptions
			continue
		}

		entry := entry // Capture variable for goroutine
		go func() {
			// Check again if subscription is active
			if !entry.active.Load() {
				return
			}

			// Use non-blocking send to avoid deadlocks
			select {
			case entry.ch <- event:
				// Message sent successfully
			case <-ctx.Done():
				// Context canceled
			default:
				// Channel is full, skip this message
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

	// Create a new subscription entry
	entry := &channelEntry{ch: ch}
	entry.active.Store(true)

	// Create the topic if it doesn't exist
	b.subscribers[topic] = append(b.subscribers[topic], entry)

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
	if !exists || len(subs) == 0 {
		return nil // Topic doesn't exist, nothing to unsubscribe
	}

	// Find and deactivate the channel but DO NOT close it
	for _, entry := range subs {
		if entry.ch == ch {
			// Deactivate the subscription to prevent future messages
			entry.active.Store(false)
			break
		}
	}

	// Clean up the subscribers list (only removes entries, doesn't close channels)
	b.cleanupInactiveSubscriptions(topic)

	return nil
}

// cleanupInactiveSubscriptions removes inactive subscriptions from the subscribers list
func (b *ChannelBus) cleanupInactiveSubscriptions(topic string) {
	subs, exists := b.subscribers[topic]
	if !exists || len(subs) == 0 {
		return
	}

	active := make([]*channelEntry, 0, len(subs))
	for _, entry := range subs {
		if entry.active.Load() {
			active = append(active, entry)
		}
	}

	if len(active) == 0 {
		// No active subscriptions, remove the topic
		delete(b.subscribers, topic)
	} else {
		// Update with only active subscriptions
		b.subscribers[topic] = active
	}
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
		for _, entry := range subs {
			entry.active.Store(false)
			close(entry.ch)
		}
	}

	// Clear the subscribers map
	b.subscribers = nil

	return nil
}