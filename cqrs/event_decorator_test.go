package cqrs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample event for testing
type TestEvent struct {
	ID      string
	Payload string
}

// Simple error for testing
var ErrEventFailed = errors.New("event handling failed")

func TestApplyEventDecorators(t *testing.T) {
	t.Run("single decorator", func(t *testing.T) {
		// Create a flag to track if the decorator was called
		decoratorCalled := false

		// Create a base handler
		baseHandler := func(ctx context.Context, event *TestEvent) error {
			return nil
		}

		// Create a simple decorator that sets a flag
		decorator := func(next cqrs.EventHandlerFunc[TestEvent]) cqrs.EventHandlerFunc[TestEvent] {
			return func(ctx context.Context, event *TestEvent) error {
				decoratorCalled = true
				return next(ctx, event)
			}
		}

		// Apply the decorator
		decoratedHandler := cqrs.ApplyEventDecorators(baseHandler, decorator)

		// Handle the event
		err := decoratedHandler(context.Background(), &TestEvent{ID: "1", Payload: "test"})

		// Verify the decorator was called and no error occurred
		require.NoError(t, err)
		assert.True(t, decoratorCalled)
	})

	t.Run("multiple decorators in order", func(t *testing.T) {
		// Track the order of decorator calls
		executionOrder := []string{}

		// Create a base handler
		baseHandler := func(ctx context.Context, event *TestEvent) error {
			executionOrder = append(executionOrder, "base")
			return nil
		}

		// Create decorators that append to the execution order
		firstDecorator := func(next cqrs.EventHandlerFunc[TestEvent]) cqrs.EventHandlerFunc[TestEvent] {
			return func(ctx context.Context, event *TestEvent) error {
				executionOrder = append(executionOrder, "first")
				return next(ctx, event)
			}
		}

		secondDecorator := func(next cqrs.EventHandlerFunc[TestEvent]) cqrs.EventHandlerFunc[TestEvent] {
			return func(ctx context.Context, event *TestEvent) error {
				executionOrder = append(executionOrder, "second")
				return next(ctx, event)
			}
		}

		// Apply decorators
		decoratedHandler := cqrs.ApplyEventDecorators(baseHandler, firstDecorator, secondDecorator)

		// Handle the event
		err := decoratedHandler(context.Background(), &TestEvent{ID: "1", Payload: "test"})

		// Verify execution order and no error
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "second", "base"}, executionOrder)
	})

	t.Run("error handling in decorators", func(t *testing.T) {
		// Track if error handler was called
		errorHandlerCalled := false

		// Create a base handler that returns an error
		baseHandler := func(ctx context.Context, event *TestEvent) error {
			return ErrEventFailed
		}

		// Create a decorator that handles errors
		errorHandlingDecorator := func(next cqrs.EventHandlerFunc[TestEvent]) cqrs.EventHandlerFunc[TestEvent] {
			return func(ctx context.Context, event *TestEvent) error {
				err := next(ctx, event)
				if err != nil {
					errorHandlerCalled = true
					// We could transform or wrap the error here
					return err
				}
				return nil
			}
		}

		// Apply decorator
		decoratedHandler := cqrs.ApplyEventDecorators(baseHandler, errorHandlingDecorator)

		// Handle the event
		err := decoratedHandler(context.Background(), &TestEvent{ID: "1", Payload: "test"})

		// Verify error was passed through and handler was called
		require.Error(t, err)
		assert.Equal(t, ErrEventFailed, err)
		assert.True(t, errorHandlerCalled)
	})

	t.Run("decorator modifying event", func(t *testing.T) {
		// Create a base handler that verifies the event payload
		var receivedPayload string
		baseHandler := func(ctx context.Context, event *TestEvent) error {
			receivedPayload = event.Payload
			return nil
		}

		// Create a decorator that modifies the event
		modifyingDecorator := func(next cqrs.EventHandlerFunc[TestEvent]) cqrs.EventHandlerFunc[TestEvent] {
			return func(ctx context.Context, event *TestEvent) error {
				// Modify the event before passing it to the handler
				event.Payload = event.Payload + "-modified"
				return next(ctx, event)
			}
		}

		// Apply decorator
		decoratedHandler := cqrs.ApplyEventDecorators(baseHandler, modifyingDecorator)

		// Handle the event
		err := decoratedHandler(context.Background(), &TestEvent{ID: "1", Payload: "test"})

		// Verify the event was modified
		require.NoError(t, err)
		assert.Equal(t, "test-modified", receivedPayload)
	})
}

func TestWithEventDecorators(t *testing.T) {
	t.Run("create event handler with decorators", func(t *testing.T) {
		// Create call tracking variables
		handlerCalled := false
		decoratorCalled := false

		// Create a base handler function
		baseHandler := func(ctx context.Context, event *TestEvent) error {
			handlerCalled = true
			return nil
		}

		// Create a decorator
		decorator := func(next cqrs.EventHandlerFunc[TestEvent]) cqrs.EventHandlerFunc[TestEvent] {
			return func(ctx context.Context, event *TestEvent) error {
				decoratorCalled = true
				return next(ctx, event)
			}
		}

		// Create an event handler with the decorator
		handler := cqrs.WithEventDecorators(baseHandler, decorator)

		// Use the event handler
		err := handler.Handle(context.Background(), &TestEvent{ID: "1", Payload: "test"})

		// Verify everything was called correctly
		require.NoError(t, err)
		assert.True(t, handlerCalled, "Base handler should be called")
		assert.True(t, decoratorCalled, "Decorator should be called")
		assert.NotEmpty(t, handler.HandlerName(), "Handler should have a name")
	})
}
