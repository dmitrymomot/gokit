package cqrs

import (
	"context"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/dmitrymomot/gokit/utils"
	"github.com/sony/gobreaker"
)

// EventHandler is an alias for cqrs.EventHandler.
type EventHandler = cqrs.EventHandler

// NewEventHandler creates a new EventHandler implementation based on provided function
// and event type inferred from function argument.
// The event handler name is inferred from the event type.
func NewEventHandler[Event any](
	handleFunc func(ctx context.Context, event *Event) error,
) EventHandler {
	handlerName := utils.QualifiedFuncName(handleFunc)
	return cqrs.NewEventHandler(handlerName, handleFunc)
}

// EventProcessor creates a new event processor with the specified event handlers.
// It uses the watermill router and slog logger.
// The event processor generates a subscribe topic for each event handler.
// The event processor also generates a publish topic for each event handler.
// The event processor uses the provided consumer group for the subscribe topics.
func EventProcessor(
	ctx context.Context,
	log *slog.Logger,
	subscriber SubscriberConstructor,
	errorHandler func(context.Context, error) error,
	events ...EventHandler,
) error {
	// Check if context is already cancelled before proceeding
	if err := ctx.Err(); err != nil {
		return err
	}

	// Wrap the slog logger with a custom watermill logger
	logger := watermill.NewSlogLogger(log)

	// Create a new message router
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return err
	}
	defer router.Close()

	// Launch goroutine to handle graceful shutdown
	go func() {
		// Wait for context cancellation signal
		<-ctx.Done()
		logger.Info("Shutting down event processor router", nil)
		// Close the router when shutdown is triggered
		router.Close()
	}()

	// Router level middleware are executed for every message sent to the router
	router.AddMiddleware(
		// CorrelationID will copy the correlation id from the incoming message's metadata to the produced messages
		middleware.CorrelationID,

		// Recoverer handles panics from handlers.
		// In this case, it passes them as errors to the Retry middleware.
		middleware.Recoverer,

		// Timeout middleware will cancel the context after the specified timeout.
		// It will also Nack the message if the context is canceled.
		middleware.Timeout(time.Second*30),

		// CircuitBreaker middleware will open the circuit if the handler returns an error.
		// It will close the circuit after the specified timeout.
		// The circuit will be half-open after the timeout, and one successful execution will close it.
		// The circuit will be open again if the handler returns an error.
		// The circuit will be closed after the specified cooldown.
		middleware.NewCircuitBreaker(gobreaker.Settings{
			Name:        "event-processor",
			MaxRequests: 2,
			Interval:    time.Second * 5,
			Timeout:     time.Second * 10,
		}).Middleware,

		// The handler function is retried if it returns an error.
		// After MaxRetries, the message is Nacked and it's up to the PubSub to resend it.
		middleware.Retry{
			MaxRetries:      1,
			InitialInterval: time.Millisecond * 100,
			MaxInterval:     time.Second * 5,
			MaxElapsedTime:  time.Minute * 5,
			Logger:          logger,
		}.Middleware,
	)

	// Add signal handler to gracefully shutdown the router
	router.AddPlugin(plugin.SignalsHandler)

	// If error handler is not provided, use a default one that does nothing.
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, err error) error { return nil }
	}

	processor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: genEventSubscribeTopic,
		SubscriberConstructor:  eventSubscriberConstructor(subscriber),
		Marshaler:              marshaler,
		Logger:                 logger,
		OnHandle:               onHandleEvent(errorHandler),
		AckOnUnknownEvent:      false,
	})
	if err != nil {
		return err
	}

	if err := processor.AddHandlers(events...); err != nil {
		return err
	}

	return router.Run(ctx)
}

// EventProcessorFunc is a function that wraps the EventProcessor function to use it in the error group.
func EventProcessorFunc(
	ctx context.Context,
	log *slog.Logger,
	subscriber SubscriberConstructor,
	errorHandler func(context.Context, error) error,
	events ...EventHandler,
) func() error {
	return func() error {
		return EventProcessor(ctx, log, subscriber, errorHandler, events...)
	}
}

// genEventSubscribeTopic generates a subscribe topic for the event processor.
func genEventSubscribeTopic(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
	return params.EventName, nil
	// return fmt.Sprintf("%s.%s", params.EventName, params.EventHandler.HandlerName()), nil
}

// eventSubscriberConstructor creates a new event subscriber.
func eventSubscriberConstructor(subscriber SubscriberConstructor) cqrs.EventProcessorSubscriberConstructorFn {
	return func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
		return subscriber(params.HandlerName)
	}
}

// onHandleEvent is a function that logs the event processing.
func onHandleEvent(errorHandler func(ctx context.Context, err error) error) cqrs.EventProcessorOnHandleFn {
	return func(params cqrs.EventProcessorOnHandleParams) error {
		ctx := params.Message.Context()
		if err := params.Handler.Handle(ctx, params.Event); err != nil {
			if err := errorHandler(ctx, err); err != nil {
				return err
			}
		}

		return nil
	}
}
