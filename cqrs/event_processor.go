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

// EventProcessorConfig defines the configuration options for the CQRS event processor.
// It allows customization of message handling, error management, logging, and retry policies.
type EventProcessorConfig struct {
	// SubscriberConstructor is a function that creates a new Watermill message subscriber
	// for a given event topic. This is required and used to receive incoming events.
	SubscriberConstructor SubscriberConstructor

	// Logger is the structured logger instance used for logging event processing events.
	// If nil, logging will be disabled, but this is not recommended for production use.
	Logger *slog.Logger

	// Publisher is the message publisher used to send unprocessable messages to a dedicated topic
	// when the poison queue middleware is enabled. Required only if the poison queue feature is used.
	Publisher message.Publisher

	// ErrorHandler is a function called when an event handler returns an error.
	// It receives the context and the error, and can perform custom error handling logic.
	// If nil, a default no-op handler will be used that simply ignores errors.
	ErrorHandler func(context.Context, error) error

	// ErrorsIgnore contains a list of specific error types that should be ignored
	// by the event processor. Messages that generate these errors will still be acknowledged.
	// Useful for non-critical errors that shouldn't interrupt processing.
	ErrorsIgnore []error

	// UnprocessableMessageErrorFilter is a function that determines whether an error indicates
	// that a message is unprocessable and should be moved to the poison queue.
	// This is only used when Publisher and UnprocessableMessageTopic are configured.
	UnprocessableMessageErrorFilter func(error) bool

	// UnprocessableMessageTopic specifies the topic name where unprocessable messages
	// will be published. Required if the poison queue feature is enabled.
	UnprocessableMessageTopic string

	// HandlerTimeout defines the maximum duration an event handler can run before timing out.
	// After this duration, the context passed to the handler will be canceled.
	// This helps prevent handlers from running indefinitely.
	HandlerTimeout time.Duration

	// MaxRetries specifies the maximum number of retry attempts for a failed event handler
	// before giving up and potentially moving the message to the poison queue.
	// This provides resilience against transient failures.
	MaxRetries int
}

// NewEventHandler creates a new EventHandler implementation based on provided function
// and event type inferred from function argument.
// The event handler name is inferred from the event type.
func NewEventHandler[Event any](
	handleFunc func(ctx context.Context, event *Event) error,
) EventHandler {
	var event Event
	handlerName := utils.GetNameFromStruct(event, utils.StructName)
	return cqrs.NewEventHandler(handlerName, handleFunc)
}

// EventProcessor creates a new event processor with the specified event handlers.
// It uses the watermill router and slog logger.
// The event processor generates a subscribe topic for each event handler.
// The event processor also generates a publish topic for each event handler.
// The event processor uses the provided consumer group for the subscribe topics.
func EventProcessor(
	ctx context.Context,
	cfg EventProcessorConfig,
	events ...EventHandler,
) error {
	// Check if context is already cancelled before proceeding
	if err := ctx.Err(); err != nil {
		return err
	}

	// Merge passed config with default
	cfg = defaultEventProcessorConfig(cfg)

	var logger watermill.LoggerAdapter = watermill.NopLogger{}
	if cfg.Logger != nil {
		// Wrap the slog logger with a custom watermill logger
		logger = watermill.NewSlogLogger(cfg.Logger)
	}

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

		// The handler function is retried if it returns an error.
		// After MaxRetries, the message is Nacked and it's up to the PubSub to resend it.
		middleware.Retry{
			MaxRetries:      cfg.MaxRetries,
			InitialInterval: time.Millisecond * 100,
			MaxInterval:     time.Second * 5,
			MaxElapsedTime:  time.Minute * 5,
			Logger:          logger,
		}.Middleware,

		// Recoverer handles panics from handlers.
		// In this case, it passes them as errors to the Retry middleware.
		middleware.Recoverer,

		// Timeout middleware will cancel the context after the specified timeout.
		// It will also Nack the message if the context is canceled.
		middleware.Timeout(cfg.HandlerTimeout),

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
	)

	// Ignore errors from the event processor
	if len(cfg.ErrorsIgnore) > 0 {
		router.AddMiddleware(middleware.NewIgnoreErrors(cfg.ErrorsIgnore).Middleware)
	}

	// PoisonQueue provides a middleware that salvages unprocessable messages and published them on a separate topic.
	// The main middleware chain then continues on, business as usual.
	if cfg.Publisher != nil {
		poisonMiddleware, err := middleware.PoisonQueueWithFilter(
			cfg.Publisher,
			cfg.UnprocessableMessageTopic,
			cfg.UnprocessableMessageErrorFilter,
		)
		if err != nil {
			return err
		}
		router.AddMiddleware(poisonMiddleware)
	}

	// Add signal handler to gracefully shutdown the router
	router.AddPlugin(plugin.SignalsHandler)

	processor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
		GenerateSubscribeTopic: genEventSubscribeTopic,
		SubscriberConstructor:  eventSubscriberConstructor(cfg.SubscriberConstructor),
		Marshaler:              marshaler,
		Logger:                 logger,
		OnHandle:               eventProcessorOnHandle(cfg.ErrorHandler),
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
// It returns a function that can be used in the error group.
func EventProcessorFunc(
	ctx context.Context,
	cfg EventProcessorConfig,
	events ...EventHandler,
) func() error {
	return func() error {
		return EventProcessor(ctx, cfg, events...)
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

// eventProcessorOnHandle is a function that is called after the event is handled.
func eventProcessorOnHandle(errorHandler func(ctx context.Context, err error) error) cqrs.EventProcessorOnHandleFn {
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

// defaultEventProcessorConfig applies reasonable default values to any unset fields
// in the EventProcessorConfig. The SubscriberConstructor field must still be provided
// as it's required for the event processor to function.
func defaultEventProcessorConfig(cfg EventProcessorConfig) EventProcessorConfig {
	// If error handler is not provided, use a default one that does nothing.
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(ctx context.Context, err error) error { return nil }
	}

	// Default empty slice for errors to ignore
	if cfg.ErrorsIgnore == nil {
		cfg.ErrorsIgnore = []error{}
	}

	// Default unprocessable message error filter
	if cfg.UnprocessableMessageErrorFilter == nil && cfg.Publisher != nil {
		cfg.UnprocessableMessageErrorFilter = func(err error) bool {
			// By default, consider all errors as unprocessable
			return err != nil
		}
	}

	// Default unprocessable message topic
	if cfg.UnprocessableMessageTopic == "" && cfg.Publisher != nil {
		cfg.UnprocessableMessageTopic = "unprocessable-events"
	}

	// Default handler timeout
	if cfg.HandlerTimeout <= 0 {
		cfg.HandlerTimeout = 30 * time.Second
	}

	// Default max retries
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}

	return cfg
}
