package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/redis/go-redis/v9"
)

// Broker implements a Redis-backed broker for SSE messages
// This allows scaling SSE across multiple server instances
// by using Redis pub/sub as the message distribution mechanism.
type Broker struct {
	// Redis client connection
	client redis.UniversalClient

	// Channel name for pub/sub
	channel string

	// Subscribers map to manage active subscriptions
	subscribers map[chan sse.Message]struct{}

	// Mutex for thread safety
	mu sync.RWMutex

	// Redis subscription for receiving messages
	redisSubscription *redis.PubSub

	// Context and cancellation for shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// WaitGroup to track active goroutines
	wg sync.WaitGroup
}

// Options holds configuration options for the Redis broker
type Options struct {
	// Channel is the Redis pub/sub channel name
	// If empty, defaults to "sse_messages"
	Channel string
}

// NewBroker creates a new Redis broker
func NewBroker(client redis.UniversalClient, opts ...Options) (*Broker, error) {
	if client == nil {
		return nil, ErrNoRedisClient
	}

	// Set default options
	options := Options{
		Channel: "sse_messages",
	}

	// Apply custom options if provided
	if len(opts) > 0 && opts[0].Channel != "" {
		options.Channel = opts[0].Channel
	}

	ctx, cancel := context.WithCancel(context.Background())

	broker := &Broker{
		client:      client,
		channel:     options.Channel,
		subscribers: make(map[chan sse.Message]struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Set up Redis subscription
	if err := broker.setupSubscription(); err != nil {
		cancel() // Clean up context if setup fails
		return nil, err
	}

	return broker, nil
}

// setupSubscription creates the Redis pub/sub subscription
// and starts the message distribution goroutine
func (b *Broker) setupSubscription() error {
	// Create Redis subscription
	pubsub := b.client.Subscribe(b.ctx, b.channel)

	// Verify the connection
	if _, err := pubsub.Receive(b.ctx); err != nil {
		return err
	}

	b.redisSubscription = pubsub

	// Start goroutine to receive messages from Redis
	b.wg.Add(1)
	go b.receiveMessages()

	return nil
}

// receiveMessages listens for messages on the Redis subscription
// and distributes them to all active subscribers
func (b *Broker) receiveMessages() {
	defer b.wg.Done()

	// Get message channel from Redis subscription
	messageCh := b.redisSubscription.Channel()

	for {
		select {
		case <-b.ctx.Done():
			// Broker is closing
			return
		case redisMsg, ok := <-messageCh:
			if !ok {
				// Channel closed
				return
			}

			// Parse message from JSON
			var message sse.Message
			if err := json.Unmarshal([]byte(redisMsg.Payload), &message); err != nil {
				// Log error and continue
				continue
			}

			// Skip expired messages
			if message.IsExpired() {
				continue
			}

			// Distribute to all subscribers
			b.distributeMessage(message)
		}
	}
}

// distributeMessage sends a message to all active subscribers
func (b *Broker) distributeMessage(message sse.Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- message:
			// Message sent
		default:
			// Skip slow subscribers
		}
	}
}

// Publish sends a message through Redis pub/sub
func (b *Broker) Publish(ctx context.Context, message sse.Message) error {
	// Validate message
	if err := message.Validate(); err != nil {
		return err
	}

	// Skip expired messages
	if message.IsExpired() {
		return nil
	}

	// Check if broker is closed
	select {
	case <-b.ctx.Done():
		return sse.ErrBrokerClosed
	default:
		// Continue
	}

	// Marshal message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Publish to Redis
	return b.client.Publish(ctx, b.channel, data).Err()
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

	// Use a sync.Once to ensure we only unregister and close once
	var closeOnce sync.Once

	// Start goroutine to handle context cancellation
	go func() {
		select {
		case <-ctx.Done():
			// Context canceled
		case <-b.ctx.Done():
			// Broker closed
		}

		// Safely unregister subscriber and close channel
		closeOnce.Do(func() {
			// Unregister subscriber
			b.mu.Lock()
			delete(b.subscribers, ch)
			b.mu.Unlock()

			// Use a recover to handle any panics from closing already closed channel
			defer func() {
				if r := recover(); r != nil {
					// Channel was already closed, ignore the panic
				}
			}()
			
			// Close the channel
			close(ch)
		})
	}()

	return ch, nil
}

// Close terminates the broker and all subscriptions
func (b *Broker) Close() error {
	// Signal shutdown
	b.cancel()

	// Close Redis subscription
	if b.redisSubscription != nil {
		if err := b.redisSubscription.Close(); err != nil {
			return err
		}
	}

	// Clear subscribers - take a copy of the channels we need to close
	// to avoid race conditions where channels might get closed multiple times
	var channelsToClose []chan sse.Message
	
	b.mu.Lock()
	for ch := range b.subscribers {
		channelsToClose = append(channelsToClose, ch)
	}
	// Clear the subscribers map while we have the lock
	b.subscribers = make(map[chan sse.Message]struct{})
	b.mu.Unlock()
	
	// Now close the channels without holding the lock
	for _, ch := range channelsToClose {
		// Use a recover in case a channel was already closed by another goroutine
		func(c chan sse.Message) {
			defer func() {
				if r := recover(); r != nil {
					// Channel was already closed, ignore the panic
				}
			}()
			close(c)
		}(ch) // Immediately invoke the function with ch as the argument
	}

	// Wait for all goroutines to finish
	b.wg.Wait()

	return nil
}
