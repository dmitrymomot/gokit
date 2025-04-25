package cqrs_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample command for testing
type TestCommand struct {
	ID      string
	Payload string
}

// Simple errors for testing purposes
var (
	ErrCommandFailed = errors.New("command failed")
	ErrValidation    = errors.New("validation failed")
)

func TestApplyCommandDecorators(t *testing.T) {
	t.Run("single decorator", func(t *testing.T) {
		// Create a flag to track if the decorator was called
		decoratorCalled := false

		// Create a base handler
		baseHandler := func(ctx context.Context, cmd *TestCommand) error {
			return nil
		}

		// Create a simple decorator that sets a flag
		decorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
			return func(ctx context.Context, cmd *TestCommand) error {
				decoratorCalled = true
				return next(ctx, cmd)
			}
		}

		// Apply the decorator
		decoratedHandler := cqrs.ApplyCommandDecorators(baseHandler, decorator)

		// Handle the command
		err := decoratedHandler(context.Background(), &TestCommand{ID: "1", Payload: "test"})

		// Verify the decorator was called and no error occurred
		require.NoError(t, err)
		assert.True(t, decoratorCalled)
	})

	t.Run("multiple decorators in order", func(t *testing.T) {
		// Track the order of decorator calls
		executionOrder := []string{}

		// Create a base handler
		baseHandler := func(ctx context.Context, cmd *TestCommand) error {
			executionOrder = append(executionOrder, "base")
			return nil
		}

		// Create decorators that append to the execution order
		firstDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
			return func(ctx context.Context, cmd *TestCommand) error {
				executionOrder = append(executionOrder, "first")
				return next(ctx, cmd)
			}
		}

		secondDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
			return func(ctx context.Context, cmd *TestCommand) error {
				executionOrder = append(executionOrder, "second")
				return next(ctx, cmd)
			}
		}

		// Apply decorators
		decoratedHandler := cqrs.ApplyCommandDecorators(baseHandler, firstDecorator, secondDecorator)

		// Handle the command
		err := decoratedHandler(context.Background(), &TestCommand{ID: "1", Payload: "test"})

		// Verify execution order and no error
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "second", "base"}, executionOrder)
	})

	t.Run("error handling in decorators", func(t *testing.T) {
		// Track if error handler was called
		errorHandlerCalled := false

		// Create a base handler that returns an error
		baseHandler := func(ctx context.Context, cmd *TestCommand) error {
			return ErrCommandFailed
		}

		// Create a decorator that handles errors
		errorHandlingDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
			return func(ctx context.Context, cmd *TestCommand) error {
				err := next(ctx, cmd)
				if err != nil {
					errorHandlerCalled = true
					// We could transform or wrap the error here
					return err
				}
				return nil
			}
		}

		// Apply decorator
		decoratedHandler := cqrs.ApplyCommandDecorators(baseHandler, errorHandlingDecorator)

		// Handle the command
		err := decoratedHandler(context.Background(), &TestCommand{ID: "1", Payload: "test"})

		// Verify error was passed through and handler was called
		require.Error(t, err)
		assert.Equal(t, ErrCommandFailed, err)
		assert.True(t, errorHandlerCalled)
	})

	t.Run("decorator modifying command", func(t *testing.T) {
		// Create a base handler that verifies the command payload
		var receivedPayload string
		baseHandler := func(ctx context.Context, cmd *TestCommand) error {
			receivedPayload = cmd.Payload
			return nil
		}

		// Create a decorator that modifies the command
		modifyingDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
			return func(ctx context.Context, cmd *TestCommand) error {
				// Modify the command before passing it to the handler
				cmd.Payload = cmd.Payload + "-modified"
				return next(ctx, cmd)
			}
		}

		// Apply decorator
		decoratedHandler := cqrs.ApplyCommandDecorators(baseHandler, modifyingDecorator)

		// Handle the command
		err := decoratedHandler(context.Background(), &TestCommand{ID: "1", Payload: "test"})

		// Verify the command was modified
		require.NoError(t, err)
		assert.Equal(t, "test-modified", receivedPayload)
	})
}

func TestWithDecorators(t *testing.T) {
	t.Run("create command handler with decorators", func(t *testing.T) {
		// Create call tracking variables
		handlerCalled := false
		decoratorCalled := false

		// Create a base handler function
		baseHandler := func(ctx context.Context, cmd *TestCommand) error {
			handlerCalled = true
			return nil
		}

		// Create a decorator
		decorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
			return func(ctx context.Context, cmd *TestCommand) error {
				decoratorCalled = true
				return next(ctx, cmd)
			}
		}

		// Create a command handler with the decorator
		handler := cqrs.WithDecorators(baseHandler, decorator)

		// Use the command handler
		err := handler.Handle(context.Background(), &TestCommand{ID: "1", Payload: "test"})

		// Verify everything was called correctly
		require.NoError(t, err)
		assert.True(t, handlerCalled, "Base handler should be called")
		assert.True(t, decoratorCalled, "Decorator should be called")
		assert.NotEmpty(t, handler.HandlerName(), "Handler should have a name")
	})
}

// Example of common decorators that would be useful in practice
func ExampleWithDecorators() {
	// Create a logging decorator
	loggingDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
		return func(ctx context.Context, cmd *TestCommand) error {
			// Log before handling
			println("Handling command:", cmd.ID)

			// Execute handler
			err := next(ctx, cmd)

			// Log after handling
			if err != nil {
				println("Command handling failed:", err.Error())
			} else {
				println("Command handled successfully")
			}

			return err
		}
	}

	// Create a validation decorator
	validationDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
		return func(ctx context.Context, cmd *TestCommand) error {
			// Validate command
			if cmd.ID == "" {
				return errors.New("command ID is required")
			}

			// Proceed with execution if valid
			return next(ctx, cmd)
		}
	}

	// Create a timeout decorator
	timeoutDecorator := func(next cqrs.CommandHandlerFunc[TestCommand]) cqrs.CommandHandlerFunc[TestCommand] {
		return func(ctx context.Context, cmd *TestCommand) error {
			// Create a timeout context
			timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			// Execute with timeout context
			return next(timeoutCtx, cmd)
		}
	}

	// Base handler implementation
	baseHandler := func(ctx context.Context, cmd *TestCommand) error {
		// Actual command handling logic
		return nil
	}

	// Create a command handler with decorators
	handler := cqrs.WithDecorators(
		baseHandler,
		loggingDecorator,
		validationDecorator,
		timeoutDecorator,
	)

	// Use it with the CQRS framework
	_ = handler
}
