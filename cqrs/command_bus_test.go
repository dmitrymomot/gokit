package cqrs_test

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/cqrs/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewCommandBus(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()
	
	// Create test message bus
	bus, _ := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})

	// Test
	commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, commandBus)
}

func TestCommandBus_Send(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()
	
	// Create test message bus
	bus, _ := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})
	
	// Subscribe to commands for verification
	cmdChan := bus.SubscribeToCommands(t, ctx, "TestCommand")
	
	// Create the command bus
	commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)
	require.NoError(t, err)

	// Test
	cmd := &test.TestCommand{
		ID:   "test-id",
		Name: "test-name",
	}

	err = commandBus.Send(ctx, cmd)

	// Assert
	require.NoError(t, err)
	
	// Verify the published message
	msg := test.WaitForMessage(t, ctx, cmdChan, test.DefaultWaitDuration)
	test.VerifyCommandMessage(t, msg, cmd)
	msg.Ack()
}

func TestCommandBus_SendWithModifiedMessage(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()
	
	// Create test message bus
	bus, _ := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})
	
	// Subscribe to commands for verification
	cmdChan := bus.SubscribeToCommands(t, ctx, "TestCommand")
	
	// Create the command bus
	commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)
	require.NoError(t, err)

	// Test
	cmd := &test.TestCommand{
		ID:   "test-id",
		Name: "test-name",
	}

	err = commandBus.SendWithModifiedMessage(ctx, cmd, func(msg *message.Message) error {
		msg.Metadata.Set("priority", "high")
		return nil
	})

	// Assert
	require.NoError(t, err)
	
	// Verify the published message
	msg := test.WaitForMessage(t, ctx, cmdChan, test.DefaultWaitDuration)
	test.VerifyCommandMessage(t, msg, cmd)
	assert.Equal(t, "high", msg.Metadata.Get("priority"))
	msg.Ack()
}

func TestCommandBus_SendWithError(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()
	
	// Create a failing publisher
	failingPublisher := &test.FailingPublisher{Err: assert.AnError}
	
	// Create the command bus with the failing publisher
	commandBus, err := cqrs.NewCommandBus(failingPublisher, logger)
	require.NoError(t, err)

	// Test
	cmd := &test.TestCommand{
		ID:   "test-id",
		Name: "test-name",
	}

	err = commandBus.Send(ctx, cmd)

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestCommandBus_SendWithModifiedMessageError(t *testing.T) {
	// Setup
	ctx := context.Background()
	logger := slog.Default()
	
	// Create test message bus
	bus, _ := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})
	
	// Create the command bus
	commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)
	require.NoError(t, err)

	// Test
	cmd := &test.TestCommand{
		ID:   "test-id",
		Name: "test-name",
	}

	modifyError := assert.AnError
	err = commandBus.SendWithModifiedMessage(ctx, cmd, func(msg *message.Message) error {
		return modifyError
	})

	// Assert
	require.Error(t, err)
	assert.ErrorIs(t, err, modifyError)
}
