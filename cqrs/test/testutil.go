// Package test provides testing utilities for the cqrs package
package test

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DefaultWaitDuration is the default time to wait for messages in tests
const DefaultWaitDuration = 1 * time.Second

// TestCommand is a test command type used across test files
type TestCommand struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TestEvent is a test event type used across test files
type TestEvent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
}

// TestBusConfig contains configuration for a test message bus
type TestBusConfig struct {
	Logger       *slog.Logger
	BufferSize   int  // Default: 100
	Persistent   bool // Default: true
	BlockPublish bool // Default: false
}

// NewDefaultTestBusConfig creates a default test bus configuration
func NewDefaultTestBusConfig() TestBusConfig {
	return TestBusConfig{
		Logger:       slog.Default(),
		BufferSize:   100,
		Persistent:   true,
		BlockPublish: false,
	}
}

// ChannelMessageBus represents a channel-based message bus for testing
type ChannelMessageBus struct {
	PubSub    *gochannel.GoChannel
	Listeners sync.Map // Topic -> []chan *message.Message
	Logger    *slog.Logger
}

// NewChannelMessageBus creates a shared GoChannel for testing
func NewChannelMessageBus(t *testing.T, ctx context.Context, cfg TestBusConfig) (*ChannelMessageBus, func(consumerGroup string) (message.Subscriber, error)) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 100
	}

	// Create a consistent config for both publisher and subscriber
	goChannelConfig := gochannel.Config{
		// OutputChannelBuffer is an int64
		OutputChannelBuffer:            int64(cfg.BufferSize),
		Persistent:                     cfg.Persistent,
		BlockPublishUntilSubscriberAck: cfg.BlockPublish,
	}

	// Create the GoChannel pubsub
	pubSub := gochannel.NewGoChannel(goChannelConfig, watermill.NewSlogLogger(cfg.Logger))
	require.NotNil(t, pubSub, "GoChannel pubsub should not be nil")

	bus := &ChannelMessageBus{
		PubSub: pubSub,
		Logger: cfg.Logger,
	}

	// Return the bus and a subscriber constructor function that returns the same pubsub
	return bus, func(consumerGroup string) (message.Subscriber, error) {
		return pubSub, nil
	}
}

// SubscribeToEvents subscribes to events of a specific type
func (b *ChannelMessageBus) SubscribeToEvents(t *testing.T, ctx context.Context, eventType string) <-chan *message.Message {
	// The EventBus uses the event struct name as the topic
	return b.subscribe(t, ctx, eventType)
}

// SubscribeToCommands subscribes to commands of a specific type
func (b *ChannelMessageBus) SubscribeToCommands(t *testing.T, ctx context.Context, commandType string) <-chan *message.Message {
	// The CommandBus uses the command struct name as the topic
	return b.subscribe(t, ctx, commandType)
}

// subscribe creates a subscription to a specific topic
func (b *ChannelMessageBus) subscribe(t *testing.T, ctx context.Context, topic string) <-chan *message.Message {
	// Create a message channel to receive messages
	msgChan := make(chan *message.Message, 100)

	// Subscribe to the topic
	messages, err := b.PubSub.Subscribe(ctx, topic)
	require.NoError(t, err, "Failed to subscribe to topic %s", topic)

	// Process messages in a goroutine
	go func() {
		defer close(msgChan)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-messages:
				if !ok {
					return
				}

				// Forward the message to our test channel
				select {
				case msgChan <- msg:
					// Message forwarded successfully
				case <-ctx.Done():
					// Acknowledge the message since we're shutting down
					msg.Ack()
					return
				}
			}
		}
	}()

	return msgChan
}

// WaitForMessage waits for a message to be received on the given channel with a timeout
func WaitForMessage(t *testing.T, ctx context.Context, msgChan <-chan *message.Message, timeout time.Duration) *message.Message {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		require.Fail(t, "Timeout waiting for message")
		return nil
	case msg, ok := <-msgChan:
		if !ok {
			require.Fail(t, "Message channel closed unexpectedly")
			return nil
		}
		return msg
	}
}

// VerifyCommandMessage verifies that a message contains the expected command
func VerifyCommandMessage(t *testing.T, msg *message.Message, expectedCmd *TestCommand) {
	require.NotNil(t, msg, "Message should not be nil")

	// Unmarshal the payload
	var receivedCmd TestCommand
	err := json.Unmarshal(msg.Payload, &receivedCmd)
	require.NoError(t, err, "Failed to unmarshal command")

	// Verify the command contents
	assert.Equal(t, expectedCmd.ID, receivedCmd.ID, "Command ID should match")
	assert.Equal(t, expectedCmd.Name, receivedCmd.Name, "Command Name should match")
}

// VerifyEventMessage verifies that a message contains the expected event
func VerifyEventMessage(t *testing.T, msg *message.Message, expectedEvent *TestEvent) {
	require.NotNil(t, msg, "Message should not be nil")

	// Unmarshal the payload
	var receivedEvent TestEvent
	err := json.Unmarshal(msg.Payload, &receivedEvent)
	require.NoError(t, err, "Failed to unmarshal event")

	// Verify the event contents
	assert.Equal(t, expectedEvent.ID, receivedEvent.ID, "Event ID should match")
	assert.Equal(t, expectedEvent.Name, receivedEvent.Name, "Event Name should match")
	assert.Equal(t, expectedEvent.Timestamp, receivedEvent.Timestamp, "Event Timestamp should match")
}

// FailingPublisher is a publisher that always returns an error
type FailingPublisher struct {
	Err error
}

// Publish implements the message.Publisher interface
func (p *FailingPublisher) Publish(topic string, messages ...*message.Message) error {
	return p.Err
}

// Close implements the message.Publisher interface
func (p *FailingPublisher) Close() error {
	return nil
}
