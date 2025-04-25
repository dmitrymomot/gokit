package cqrs_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/cqrs/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestNewCommandHandler(t *testing.T) {
	// Create a command handler
	handler := cqrs.NewCommandHandler(func(ctx context.Context, cmd *test.TestCommand) error {
		return nil
	})

	// Assertions
	require.NotNil(t, handler)
	assert.Equal(t, "TestCommand", handler.HandlerName())
}

func TestCommandProcessor(t *testing.T) {
	// Setup a logger
	logger := slog.Default()

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create test message bus using the channel-based approach
	bus, subscriberConstructor := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
		// Use a larger buffer to ensure messages don't get dropped
		BufferSize: 100,
	})

	// Command handler execution tracker
	var (
		handlerExecuted bool
		handlerMutex    sync.Mutex
		handlerErr      error
	)

	// Create a command handler that sets handlerExecuted to true
	handler := cqrs.NewCommandHandler(func(ctx context.Context, cmd *test.TestCommand) error {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()
		handlerExecuted = true

		// Validate command data
		assert.Equal(t, "test-id", cmd.ID)
		assert.Equal(t, "test-name", cmd.Name)

		return handlerErr
	})

	// Setup error handler to track errors
	var (
		errorHandlerCalled bool
		errorMutex         sync.Mutex
		errorReceived      error
	)

	errorHandler := func(ctx context.Context, err error) error {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		errorHandlerCalled = true
		errorReceived = err
		// Always return nil so we don't disrupt the test flow
		return nil
	}

	// Create error group to run the command processor
	eg, egCtx := errgroup.WithContext(ctx)

	// Start the command processor in a goroutine
	eg.Go(cqrs.CommandProcessorFunc(
		egCtx,
		cqrs.CommandProcessorConfig{
			Logger:                logger.With(slog.String("component", "command-processor")),
			SubscriberConstructor: subscriberConstructor,
			ErrorHandler:          errorHandler,
		},
		handler,
	))

	// Wait a moment for the processor to start
	time.Sleep(500 * time.Millisecond)

	// Test case 1: Successful command handling
	t.Run("Successful command handling", func(t *testing.T) {
		// Reset tracker
		handlerMutex.Lock()
		handlerExecuted = false
		handlerErr = nil
		handlerMutex.Unlock()

		errorMutex.Lock()
		errorHandlerCalled = false
		errorReceived = nil
		errorMutex.Unlock()

		// Create a command
		cmd := &test.TestCommand{
			ID:   "test-id",
			Name: "test-name",
		}

		// Use the command bus to send the command
		commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)
		require.NoError(t, err)
		err = commandBus.Send(ctx, cmd)
		require.NoError(t, err)

		// Wait for processing with retry logic
		assert.Eventually(t, func() bool {
			handlerMutex.Lock()
			defer handlerMutex.Unlock()
			return handlerExecuted
		}, 2*time.Second, 100*time.Millisecond, "Command handler should have been executed")

		// Ensure error handler was not called
		errorMutex.Lock()
		errCalled := errorHandlerCalled
		errReceived := errorReceived
		errorMutex.Unlock()
		assert.False(t, errCalled, "Error handler should not have been called")
		assert.Nil(t, errReceived)
	})

	// Test case 2: Command handler returning an error
	t.Run("Command handler returning an error", func(t *testing.T) {
		// Reset tracker
		handlerMutex.Lock()
		handlerExecuted = false
		handlerErr = errors.New("command handling failed")
		handlerMutex.Unlock()

		errorMutex.Lock()
		errorHandlerCalled = false
		errorReceived = nil
		errorMutex.Unlock()

		// Create a command
		cmd := &test.TestCommand{
			ID:   "test-id",
			Name: "test-name",
		}

		// Use the command bus to send the command
		commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)
		require.NoError(t, err)
		err = commandBus.Send(ctx, cmd)
		require.NoError(t, err)

		// Wait for processing with retry logic
		assert.Eventually(t, func() bool {
			handlerMutex.Lock()
			defer handlerMutex.Unlock()
			return handlerExecuted
		}, 2*time.Second, 100*time.Millisecond, "Command handler should have been executed")

		// Ensure error handler was called with the correct error
		assert.Eventually(t, func() bool {
			errorMutex.Lock()
			defer errorMutex.Unlock()
			return errorHandlerCalled && errorReceived != nil
		}, 2*time.Second, 100*time.Millisecond, "Error handler should have been called with an error")

		errorMutex.Lock()
		errReceived := errorReceived
		errorMutex.Unlock()
		if errReceived != nil {
			assert.Equal(t, "command handling failed", errReceived.Error())
		}
	})

	// Clean up
	cancel()

	// Wait for the processor to shut down gracefully
	err := eg.Wait()
	assert.NoError(t, err)
}

func TestCommandProcessorFunc(t *testing.T) {
	// Create a basic configuration
	logger := slog.Default()
	ctx := context.Background()

	// Create a test command handler
	handler := cqrs.NewCommandHandler(func(ctx context.Context, cmd *test.TestCommand) error {
		return nil
	})

	// Create test message bus
	bus, subscriberConstructor := test.NewChannelMessageBus(t, ctx, test.TestBusConfig{
		Logger: logger,
	})

	// Test if the function returns a non-nil function
	processorFunc := cqrs.CommandProcessorFunc(
		ctx,
		cqrs.CommandProcessorConfig{
			Logger:                logger,
			SubscriberConstructor: subscriberConstructor,
		},
		handler,
	)

	require.NotNil(t, processorFunc)
	assert.IsType(t, (func() error)(nil), processorFunc)

	// Optional: Verify the processor can actually be started (although we don't wait for it to complete)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(processorFunc)

	// Send a command to the processor
	commandBus, err := cqrs.NewCommandBus(bus.PubSub, logger)
	require.NoError(t, err)

	err = commandBus.Send(ctx, &test.TestCommand{
		ID:   "test-processor-func",
		Name: "processor-test",
	})
	require.NoError(t, err)
}
