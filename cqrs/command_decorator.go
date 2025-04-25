package cqrs

import (
	"context"
)

// CommandHandlerFunc is a function that handles a command of type Command
// It's a generic type to provide type safety when working with decorators
type CommandHandlerFunc[Command any] func(ctx context.Context, cmd *Command) error

// CommandDecorator is a function that takes a CommandHandlerFunc and returns a new CommandHandlerFunc
// with additional functionality. It's generic to maintain type safety throughout the decoration chain.
type CommandDecorator[Command any] func(CommandHandlerFunc[Command]) CommandHandlerFunc[Command]

// ApplyCommandDecorators applies multiple decorators to a command handler function
// The decorators are applied in reverse order so that the first decorator in the list
// is the outermost wrapper around the handler.
func ApplyCommandDecorators[Command any](
	handler CommandHandlerFunc[Command],
	decorators ...CommandDecorator[Command],
) CommandHandlerFunc[Command] {
	result := handler
	// Apply decorators in reverse order
	for i := len(decorators) - 1; i >= 0; i-- {
		result = decorators[i](result)
	}
	return result
}

// WithDecorators is a helper function that takes a command handler function and decorators,
// applies the decorators to the handler function, and then wraps it with NewCommandHandler
// to create a CommandHandler that can be used with the CQRS framework.
func WithDecorators[Command any](
	handler CommandHandlerFunc[Command],
	decorators ...CommandDecorator[Command],
) CommandHandler {
	// Apply decorators to handler
	decoratedHandler := ApplyCommandDecorators(handler, decorators...)

	// Create a CommandHandler using the decorated handler function
	return NewCommandHandler(decoratedHandler)
}
