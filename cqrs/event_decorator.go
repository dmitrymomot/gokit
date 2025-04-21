package cqrs

import (
	"context"
)

// EventHandlerFunc is a function that handles an event of type Event
// It's a generic type to provide type safety when working with decorators
type EventHandlerFunc[Event any] func(ctx context.Context, event *Event) error

// EventDecorator is a function that takes an EventHandlerFunc and returns a new EventHandlerFunc
// with additional functionality. It's generic to maintain type safety throughout the decoration chain.
type EventDecorator[Event any] func(EventHandlerFunc[Event]) EventHandlerFunc[Event]

// ApplyEventDecorators applies multiple decorators to an event handler function
// The decorators are applied in reverse order so that the first decorator in the list
// is the outermost wrapper around the handler.
func ApplyEventDecorators[Event any](
	handler EventHandlerFunc[Event],
	decorators ...EventDecorator[Event],
) EventHandlerFunc[Event] {
	result := handler
	// Apply decorators in reverse order
	for i := len(decorators) - 1; i >= 0; i-- {
		result = decorators[i](result)
	}
	return result
}

// WithEventDecorators is a helper function that takes an event handler function and decorators,
// applies the decorators to the handler function, and then wraps it with NewEventHandler
// to create an EventHandler that can be used with the CQRS framework.
func WithEventDecorators[Event any](
	handler EventHandlerFunc[Event],
	decorators ...EventDecorator[Event],
) EventHandler {
	// Apply decorators to handler
	decoratedHandler := ApplyEventDecorators(handler, decorators...)
	
	// Create an EventHandler using the decorated handler function
	return NewEventHandler(decoratedHandler)
}
