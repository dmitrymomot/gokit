package cqrs_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/cqrs/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestNewEventHandler(t *testing.T) {
	// Create an event handler
	handler := cqrs.NewEventHandler(func(ctx context.Context, event *test.TestEvent) error {
		return nil
	})

	// Assertions
	require.NotNil(t, handler)
	assert.Equal(t, "TestEvent", handler.HandlerName())
}

func TestEventProcessor(t *testing.T) {
	// Setup a logger
	logger := slog.Default()
	waterm := watermill.NewSlogLogger(logger.With(slog.String("test", "event_processor")))

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a GoChannel pubsub with explicit config to ensure consistent behavior
	goChannelConfig := gochannel.Config{
		OutputChannelBuffer:            10,
		Persistent:                     true,
		BlockPublishUntilSubscriberAck: false,
	}
	pubSub := gochannel.NewGoChannel(goChannelConfig, waterm)
	defer pubSub.Close()

	// Create subscriber constructor using the same pubsub
	subscriberConstructor := func(consumerGroup string) (message.Subscriber, error) {
		return pubSub, nil
	}

	// Event handler execution tracker
	var (
		handlerExecuted bool
		handlerMutex    sync.Mutex
		handlerErr      error
	)

	// Create an event handler that sets handlerExecuted to true
	handler := cqrs.NewEventHandler(func(ctx context.Context, event *test.TestEvent) error {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()
		handlerExecuted = true
		
		// Validate event data
		assert.Equal(t, "test-id", event.ID)
		assert.Equal(t, "test-name", event.Name)
		
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

	// Create error group to run the event processor
	eg, egCtx := errgroup.WithContext(ctx)

	// Start the event processor in a goroutine
	eg.Go(cqrs.EventProcessorFunc(
		egCtx,
		cqrs.EventProcessorConfig{
			Logger:                logger.With(slog.String("component", "event-processor")),
			SubscriberConstructor: subscriberConstructor,
			ErrorHandler:          errorHandler,
		},
		handler,
	))

	// Wait a moment for the processor to start
	time.Sleep(1 * time.Second)

	// Test case 1: Successful event handling
	t.Run("Successful event handling", func(t *testing.T) {
		// Reset tracker
		handlerMutex.Lock()
		handlerExecuted = false
		handlerErr = nil
		handlerMutex.Unlock()

		errorMutex.Lock()
		errorHandlerCalled = false
		errorReceived = nil
		errorMutex.Unlock()

		// Create an event
		event := &test.TestEvent{
			ID:        "test-id",
			Name:      "test-name",
			Timestamp: time.Now().Unix(),
		}
		
		// Use the event bus to publish the event
		eventBus, err := cqrs.NewEventBus(pubSub, logger)
		require.NoError(t, err)
		err = eventBus.Publish(ctx, event)
		require.NoError(t, err)

		// Wait for processing with retry logic
		assert.Eventually(t, func() bool {
			handlerMutex.Lock()
			defer handlerMutex.Unlock()
			return handlerExecuted
		}, 5*time.Second, 100*time.Millisecond, "Event handler should have been executed")

		// Ensure error handler was not called
		errorMutex.Lock()
		errCalled := errorHandlerCalled
		errReceived := errorReceived
		errorMutex.Unlock()
		assert.False(t, errCalled, "Error handler should not have been called")
		assert.Nil(t, errReceived)
	})

	// Test case 2: Event handler returning an error
	t.Run("Event handler returning an error", func(t *testing.T) {
		// Reset tracker
		handlerMutex.Lock()
		handlerExecuted = false
		handlerErr = errors.New("event handling failed")
		handlerMutex.Unlock()

		errorMutex.Lock()
		errorHandlerCalled = false
		errorReceived = nil
		errorMutex.Unlock()

		// Create an event
		event := &test.TestEvent{
			ID:        "test-id",
			Name:      "test-name",
			Timestamp: time.Now().Unix(),
		}
		
		// Use the event bus to publish the event
		eventBus, err := cqrs.NewEventBus(pubSub, logger)
		require.NoError(t, err)
		err = eventBus.Publish(ctx, event)
		require.NoError(t, err)

		// Wait for processing with retry logic
		assert.Eventually(t, func() bool {
			handlerMutex.Lock()
			defer handlerMutex.Unlock()
			return handlerExecuted
		}, 5*time.Second, 100*time.Millisecond, "Event handler should have been executed")

		// Ensure error handler was called with the correct error
		assert.Eventually(t, func() bool {
			errorMutex.Lock()
			defer errorMutex.Unlock()
			return errorHandlerCalled && errorReceived != nil
		}, 5*time.Second, 100*time.Millisecond, "Error handler should have been called with an error")
		
		errorMutex.Lock()
		errReceived := errorReceived
		errorMutex.Unlock()
		if errReceived != nil {
			assert.Equal(t, "event handling failed", errReceived.Error())
		}
	})

	// Clean up
	cancel()

	// Wait for the processor to shut down gracefully
	err := eg.Wait()
	assert.NoError(t, err)
}

func TestEventProcessorFunc(t *testing.T) {
	// Create a basic configuration
	logger := slog.Default()
	ctx := context.Background()
	handler := cqrs.NewEventHandler(func(ctx context.Context, event *test.TestEvent) error {
		return nil
	})

	// Test if the function returns a non-nil function
	processorFunc := cqrs.EventProcessorFunc(
		ctx,
		cqrs.EventProcessorConfig{
			Logger:                logger,
			SubscriberConstructor: cqrs.NewGoChannelSubscriber(ctx, logger),
		},
		handler,
	)

	require.NotNil(t, processorFunc)
	assert.IsType(t, (func() error)(nil), processorFunc)
}
