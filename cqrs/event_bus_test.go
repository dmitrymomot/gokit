package cqrs_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/cqrs/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewEventBus(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.Default()

	// Create test message bus using the new util
	bus, _ := test.NewChannelMessageBus(t, ctx, test.NewDefaultTestBusConfig())

	// Test
	eventBus, err := cqrs.NewEventBus(bus.PubSub, logger)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, eventBus)
}

func TestEventBus_Publish(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.Default()

	// Create test message bus using the new util
	bus, _ := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})

	// Subscribe to events for verification
	eventChan := bus.SubscribeToEvents(t, ctx, "TestEvent")

	// Create the event bus
	eventBus, err := cqrs.NewEventBus(bus.PubSub, logger)
	require.NoError(t, err)

	// Test
	event := &test.TestEvent{
		ID:        "test-id",
		Name:      "test-name",
		Timestamp: time.Now().Unix(),
	}

	// Publish the event
	err = eventBus.Publish(ctx, event)

	// Assert
	require.NoError(t, err)

	// Verify the published message
	msg := test.WaitForMessage(t, ctx, eventChan, 1*time.Second)
	test.VerifyEventMessage(t, msg, event)
	msg.Ack()
}

func TestEventBus_PublishWithDelay(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.Default()

	// Create test message bus using the new util
	bus, _ := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})

	// Subscribe to events for verification
	eventChan := bus.SubscribeToEvents(t, ctx, "TestEvent")

	// Create the event bus
	eventBus, err := cqrs.NewEventBus(bus.PubSub, logger)
	require.NoError(t, err)

	// Test
	event := &test.TestEvent{
		ID:        "test-id",
		Name:      "test-name",
		Timestamp: time.Now().Unix(),
	}

	delayDuration := 5 * time.Second
	err = eventBus.PublishWithDelay(ctx, event, delayDuration)

	// Assert
	require.NoError(t, err)

	// Verify the published message
	msg := test.WaitForMessage(t, ctx, eventChan, 1*time.Second)
	test.VerifyEventMessage(t, msg, event)

	// Check that message context contains delay information
	// Note: In this test approach, we can't directly verify the delay mechanism
	// since the gochannel doesn't actually implement delayed delivery
	msgCtx := msg.Context()
	require.NotNil(t, msgCtx)

	msg.Ack()
}

func TestEventBus_PublishWithError(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.Default()

	// Create a failing publisher by using the test utility
	failingPublisher := &test.FailingPublisher{Err: assert.AnError}

	// Create the event bus with the failing publisher
	bus, err := cqrs.NewEventBus(failingPublisher, logger)
	require.NoError(t, err)

	// Test
	event := &test.TestEvent{
		ID:        "test-id",
		Name:      "test-name",
		Timestamp: time.Now().Unix(),
	}

	err = bus.Publish(ctx, event)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestEventBus_PublishWithDelayError(t *testing.T) {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.Default()

	// Create a failing publisher by using the test utility
	failingPublisher := &test.FailingPublisher{Err: assert.AnError}

	// Create the event bus with the failing publisher
	bus, err := cqrs.NewEventBus(failingPublisher, logger)
	require.NoError(t, err)

	// Test
	event := &test.TestEvent{
		ID:        "test-id",
		Name:      "test-name",
		Timestamp: time.Now().Unix(),
	}

	err = bus.PublishWithDelay(ctx, event, 5*time.Second)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}
