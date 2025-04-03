package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

// StreamsBroker implements a Redis Streams-backed broker for SSE messages
// This provides reliable message delivery and automatic expiration of old messages
// to prevent unlimited accumulation in Redis.
type StreamsBroker struct {
	// Redis client connection
	client redis.UniversalClient

	// Stream name for storing messages
	streamName string

	// Max stream length (0 for unlimited)
	maxStreamLength int

	// Message retention period (0 for no automatic expiry)
	messageRetention time.Duration

	// Group name for consumer groups
	groupName string

	// Consumer name (defaults to a unique ID)
	consumerName string

	// Read block timeout
	blockDuration time.Duration

	// Last read ID
	lastID string

	// Subscribers map to manage active subscriptions
	subscribers map[chan sse.Message]struct{}

	// Mutex for thread safety
	mu sync.RWMutex

	// Context and cancellation for shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// WaitGroup to track active goroutines
	wg sync.WaitGroup
}

// StreamsOptions holds configuration options for the Redis Streams broker
type StreamsOptions struct {
	// StreamName is the Redis stream name for messages
	// If empty, defaults to "sse_messages_stream"
	StreamName string

	// MaxStreamLength limits the number of messages in the stream
	// If 0, no limit is applied (not recommended for production)
	// If > 0, stream will be trimmed to this length after each publish
	MaxStreamLength int

	// MessageRetention defines how long messages are kept
	// If 0, messages are kept indefinitely (not recommended)
	// Only used when creating a stream or consumer group
	MessageRetention time.Duration

	// GroupName for consumer group
	// If empty, defaults to "sse_consumers"
	GroupName string

	// ConsumerName within the group
	// If empty, a unique name is generated
	ConsumerName string

	// BlockDuration is how long to block on XREAD calls
	// If 0, defaults to 100ms
	BlockDuration time.Duration
}

// NewStreamsBroker creates a new Redis Streams-backed broker
func NewStreamsBroker(client redis.UniversalClient, opts ...StreamsOptions) (*StreamsBroker, error) {
	if client == nil {
		return nil, ErrNoRedisClient
	}

	// Default options
	options := StreamsOptions{
		StreamName:       "sse_messages_stream",
		MaxStreamLength:  1000,                 // Default to 1000 messages
		MessageRetention: 24 * time.Hour,       // Default to 24 hours
		GroupName:        "sse_consumers",
		ConsumerName:     fmt.Sprintf("consumer-%d", time.Now().UnixNano()),
		BlockDuration:    100 * time.Millisecond,
	}

	// Apply custom options if provided
	if len(opts) > 0 {
		opt := opts[0]
		if opt.StreamName != "" {
			options.StreamName = opt.StreamName
		}
		if opt.MaxStreamLength > 0 {
			options.MaxStreamLength = opt.MaxStreamLength
		}
		if opt.MessageRetention > 0 {
			options.MessageRetention = opt.MessageRetention
		}
		if opt.GroupName != "" {
			options.GroupName = opt.GroupName
		}
		if opt.ConsumerName != "" {
			options.ConsumerName = opt.ConsumerName
		}
		if opt.BlockDuration > 0 {
			options.BlockDuration = opt.BlockDuration
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	broker := &StreamsBroker{
		client:           client,
		streamName:       options.StreamName,
		maxStreamLength:  options.MaxStreamLength,
		messageRetention: options.MessageRetention,
		groupName:        options.GroupName,
		consumerName:     options.ConsumerName,
		blockDuration:    options.BlockDuration,
		lastID:           "0",  // Start from the beginning
		subscribers:      make(map[chan sse.Message]struct{}),
		ctx:              ctx,
		cancel:           cancel,
	}

	// Set up stream and consumer group
	if err := broker.setupStream(ctx); err != nil {
		cancel() // Clean up context if setup fails
		return nil, err
	}

	// Start message consumer goroutine
	broker.wg.Add(1)
	go broker.consumeMessages()

	return broker, nil
}

// setupStream initializes the Redis stream and consumer group
func (b *StreamsBroker) setupStream(ctx context.Context) error {
	// Check if stream exists by attempting to get stream info
	info, err := b.client.XInfoStream(ctx, b.streamName).Result()
	if err != nil {
		if err != redis.Nil {
			// If error is not "key does not exist", it's a connection error
			return fmt.Errorf("%w: %v", ErrRedisConnectionFailed, err)
		}
		
		// Stream doesn't exist, create it with an initial message
		// This is necessary because we can't create a consumer group on a non-existent stream
		initMsg := map[string]interface{}{
			"type": "init",
			"time": time.Now().Format(time.RFC3339),
		}
		
		createCmd := b.client.XAdd(ctx, &redis.XAddArgs{
			Stream: b.streamName,
			ID:     "*", // Auto-generate ID
			Values: initMsg,
		})
		
		if err := createCmd.Err(); err != nil {
			return fmt.Errorf("failed to create stream: %w", err)
		}
		
		// Now we know the stream exists
	} else {
		// Stream exists, check length and possibly trim
		length := info.Length
		if b.maxStreamLength > 0 && length > int64(b.maxStreamLength) {
			if err := b.client.XTrimMaxLen(ctx, b.streamName, int64(b.maxStreamLength)).Err(); err != nil {
				// Log error but continue - this is not fatal
				fmt.Printf("failed to trim stream: %v\n", err)
			}
		}
	}

	// Try to create the consumer group
	err = b.client.XGroupCreate(ctx, b.streamName, b.groupName, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		// Only return error if it's not the "group already exists" error
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Set the last ID to special ">" value for new messages only
	b.lastID = ">"

	return nil
}

// consumeMessages reads messages from the Redis stream and distributes them to subscribers
func (b *StreamsBroker) consumeMessages() {
	defer b.wg.Done()

	for {
		select {
		case <-b.ctx.Done():
			// Broker is shutting down
			return
		default:
			// Continue reading
		}

		// Read from the stream with timeout
		streams, err := b.client.XReadGroup(b.ctx, &redis.XReadGroupArgs{
			Group:    b.groupName,
			Consumer: b.consumerName,
			Streams:  []string{b.streamName, b.lastID},
			Count:    10,
			Block:    b.blockDuration,
		}).Result()

		if err != nil {
			// Check if context canceled
			if b.ctx.Err() != nil {
				return
			}
			// If error is timeout, just try again
			if err == redis.Nil {
				continue
			}
			// For other errors, wait a bit and retry
			time.Sleep(time.Second)
			continue
		}

		// Process received messages
		if len(streams) > 0 && len(streams[0].Messages) > 0 {
			messages := streams[0].Messages
			
			// Use errgroup for concurrent processing
			g, ctx := errgroup.WithContext(b.ctx)
			
			// Process messages
			for _, xMsg := range messages {
				msgID := xMsg.ID
				
				// Track last message ID for acknowledgement
				ackID := msgID
				
				// Process the message
				valueMap := xMsg.Values
				
				// Skip initial message
				if typ, ok := valueMap["type"].(string); ok && typ == "init" {
					continue
				}
				
				// Check for SSE message data
				msgData, ok := valueMap["message"].(string)
				if !ok {
					continue
				}
				
				// Parse message
				var message sse.Message
				if err := json.Unmarshal([]byte(msgData), &message); err != nil {
					continue
				}
				
				// Skip expired messages
				if message.IsExpired() {
					continue
				}
				
				// Copy message to avoid capture issues
				msg := message
				
				// Distribute message to subscribers
				g.Go(func() error {
					b.distributeMessage(msg)
					return nil
				})
				
				// Acknowledge message in a separate goroutine to avoid blocking
				g.Go(func() error {
					return b.client.XAck(ctx, b.streamName, b.groupName, ackID).Err()
				})
			}
			
			// Wait for all goroutines to complete
			_ = g.Wait()
		}
	}
}

// distributeMessage sends a message to all active subscribers
func (b *StreamsBroker) distributeMessage(message sse.Message) {
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

// Publish sends a message through Redis Streams
func (b *StreamsBroker) Publish(ctx context.Context, message sse.Message) error {
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
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add message to the stream, with trimming if needed
	cmd := b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: b.streamName,
		ID:     "*", // Auto-generate ID
		Values: map[string]interface{}{
			"message": string(data),
			"time":    time.Now().Format(time.RFC3339),
		},
		MaxLen: int64(b.maxStreamLength),
		Approx: true, // Use approximate trimming for better performance
	})

	if err := cmd.Err(); err != nil {
		return fmt.Errorf("failed to publish message to stream: %w", err)
	}

	return nil
}

// Subscribe returns a channel that receives all messages
func (b *StreamsBroker) Subscribe(ctx context.Context) (<-chan sse.Message, error) {
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
func (b *StreamsBroker) Close() error {
	// Signal shutdown
	b.cancel()

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

// CleanupStream removes expired messages from the stream
// This can be called periodically if automatic trimming is not sufficient
func (b *StreamsBroker) CleanupStream(ctx context.Context) error {
	if b.maxStreamLength <= 0 {
		return nil // No cleanup needed if no max length set
	}

	// Trim the stream to max length
	if err := b.client.XTrimMaxLen(ctx, b.streamName, int64(b.maxStreamLength)).Err(); err != nil {
		return fmt.Errorf("failed to trim stream: %w", err)
	}

	return nil
}

// SetMaxStreamLength sets the maximum stream length for cleanup trimming.
func (b *StreamsBroker) SetMaxStreamLength(max int) {
	b.maxStreamLength = max
}

// Ensure StreamsBroker implements the sse.Broker interface
var _ sse.Broker = (*StreamsBroker)(nil)
