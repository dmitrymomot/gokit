package bus

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/redis/go-redis/v9"
)

// RedisBus implements the MessageBus interface using Redis PubSub.
type RedisBus struct {
	client       redis.UniversalClient
	mu           sync.RWMutex
	subscriptions map[string]map[chan sse.Event]struct{}
	closed       bool
	bufferSize   int
}

// NewRedisBus creates a new Redis SSE message bus.
// It accepts any redis.UniversalClient (e.g., *redis.Client, *redis.ClusterClient).
func NewRedisBus(client redis.UniversalClient) (*RedisBus, error) {
	return NewRedisBusWithConfig(client, 100)
}

// NewRedisBusWithConfig creates a new Redis SSE message bus with custom buffer size.
func NewRedisBusWithConfig(client redis.UniversalClient, bufferSize int) (*RedisBus, error) {
	// Ping the server to ensure connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Join(sse.ErrMessageBusClosed, err)
	}

	return &RedisBus{
		client:        client,
		mu:            sync.RWMutex{},
		subscriptions: make(map[string]map[chan sse.Event]struct{}),
		closed:        false,
		bufferSize:    bufferSize,
	}, nil
}

// Publish sends a message to the specified topic.
// The message will be encoded as JSON before publishing.
func (r *RedisBus) Publish(ctx context.Context, topic string, event sse.Event) error {
	if topic == "" {
		return sse.ErrTopicEmpty
	}

	r.mu.RLock()
	if r.closed {
		r.mu.RUnlock()
		return sse.ErrMessageBusClosed
	}
	r.mu.RUnlock()

	// Marshal event to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return errors.Join(sse.ErrMessageEmpty, err)
	}

	// Publish to Redis
	err = r.client.Publish(ctx, topic, eventData).Err()
	if err != nil {
		return errors.Join(sse.ErrMessageEmpty, err)
	}

	return nil
}

// Subscribe returns a channel that receives events for a specific topic.
func (r *RedisBus) Subscribe(ctx context.Context, topic string) (<-chan sse.Event, error) {
	if topic == "" {
		return nil, sse.ErrTopicEmpty
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, sse.ErrMessageBusClosed
	}

	// Create Redis PubSub subscription
	pubsub := r.client.Subscribe(ctx, topic)

	// Check the connection
	if _, err := pubsub.Receive(ctx); err != nil {
		pubsub.Close()
		return nil, errors.Join(sse.ErrMessageEmpty, err)
	}

	// Create a buffered channel to receive events
	eventCh := make(chan sse.Event, r.bufferSize)

	// Create or get the subscriptions map for this topic
	if _, exists := r.subscriptions[topic]; !exists {
		r.subscriptions[topic] = make(map[chan sse.Event]struct{})
	}

	// Store the channel in the subscriptions map
	r.subscriptions[topic][eventCh] = struct{}{}

	// Start a goroutine to receive messages from Redis and send them to the channel
	go func() {
		deferFunc := func() {
			pubsub.Close()
			r.mu.Lock()
			delete(r.subscriptions[topic], eventCh)
			// Clean up the topic if no more subscriptions
			if len(r.subscriptions[topic]) == 0 {
				delete(r.subscriptions, topic)
			}
			r.mu.Unlock()
			close(eventCh)
		}

		// Get the message channel from Redis pubsub
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				deferFunc()
				return

			case msg, ok := <-ch:
				if !ok {
					deferFunc()
					return
				}

				// Parse the message payload as Event
				var event sse.Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					// If parsing fails, create a simple event with the raw payload
					event = sse.Event{
						Event: msg.Channel,
						Data:  msg.Payload,
					}
				}

				// Send the event to the channel
				select {
				case eventCh <- event:
					// Event sent successfully
				case <-ctx.Done():
					deferFunc()
					return
				default:
					// Channel is full, skip this message
				}
			}
		}
	}()

	return eventCh, nil
}

// Unsubscribe removes a subscription for a specific topic.
func (r *RedisBus) Unsubscribe(ctx context.Context, topic string, ch <-chan sse.Event) error {
	if topic == "" {
		return sse.ErrTopicEmpty
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return sse.ErrMessageBusClosed
	}

	// Check if the topic exists in subscriptions
	if topicSubs, exists := r.subscriptions[topic]; exists {
		// Find and delete the channel
		for eventCh := range topicSubs {
			if ch == eventCh {
				delete(topicSubs, eventCh)
				// If this was the last subscription for the topic, remove the topic
				if len(topicSubs) == 0 {
					delete(r.subscriptions, topic)
				}
				return nil
			}
		}
	}

	// Subscription not found, but this is not an error
	return nil
}

// Close closes the Redis client connection and all subscriptions.
func (r *RedisBus) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	// Mark as closed to prevent new subscriptions
	r.closed = true

	// Close Redis client
	if err := r.client.Close(); err != nil {
		return errors.Join(sse.ErrMessageBusClosed, err)
	}

	return nil
}