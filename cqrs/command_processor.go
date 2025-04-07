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

// CommandHandler is an alias for cqrs.CommandHandler.
type CommandHandler = cqrs.CommandHandler

// CommandProcessorConfig defines the configuration options for the CQRS command processor.
// It allows customization of message handling, error management, logging, and retry policies.
type CommandProcessorConfig struct {
	// SubscriberConstructor is a function that creates a new Watermill message subscriber
	// for a given command topic. This is required and used to receive incoming commands.
	SubscriberConstructor SubscriberConstructor

	// Logger is the structured logger instance used for logging command processing events.
	// If nil, logging will be disabled, but this is not recommended for production use.
	Logger *slog.Logger

	// Publisher is the message publisher used to send unprocessable messages to a dedicated topic
	// when the poison queue middleware is enabled. Required only if the poison queue feature is used.
	Publisher message.Publisher

	// ErrorHandler is a function called when a command handler returns an error.
	// It receives the context and the error, and can perform custom error handling logic.
	// If nil, a default no-op handler will be used that simply ignores errors.
	ErrorHandler func(context.Context, error) error

	// ErrorsIgnore contains a list of specific error types that should be ignored
	// by the command processor. Messages that generate these errors will still be acknowledged.
	// Useful for non-critical errors that shouldn't interrupt processing.
	ErrorsIgnore []error

	// UnprocessableMessageErrorFilter is a function that determines whether an error indicates
	// that a message is unprocessable and should be moved to the poison queue.
	// This is only used when Publisher and UnprocessableMessageTopic are configured.
	UnprocessableMessageErrorFilter func(error) bool

	// UnprocessableMessageTopic specifies the topic name where unprocessable messages
	// will be published. Required if the poison queue feature is enabled.
	UnprocessableMessageTopic string

	// HandlerTimeout defines the maximum duration a command handler can run before timing out.
	// After this duration, the context passed to the handler will be canceled.
	// This helps prevent handlers from running indefinitely.
	HandlerTimeout time.Duration

	// MaxRetries specifies the maximum number of retry attempts for a failed command handler
	// before giving up and potentially moving the message to the poison queue.
	// This provides resilience against transient failures.
	MaxRetries int
}

// NewCommandHandler creates a new CommandHandler implementation based on provided function
// and command type inferred from function argument.
func NewCommandHandler[Command any](
	handleFunc func(ctx context.Context, cmd *Command) error,
) CommandHandler {
	var cmd Command
	handlerName := utils.GetNameFromStruct(cmd, utils.StructName)
	return cqrs.NewCommandHandler(handlerName, handleFunc)
}

// CommandProcessor creates a new command processor with the specified command handlers.
// It uses the watermill router and slog logger.
// The command processor generates a subscribe topic for each command handler.
func CommandProcessor(
	ctx context.Context,
	cfg CommandProcessorConfig,
	cmds ...CommandHandler,
) error {
	// Check if context is already cancelled before proceeding
	if err := ctx.Err(); err != nil {
		return err
	}

	// Merge passed config with default
	cfg = defaultCommandProcessorConfig(cfg)

	var logger watermill.LoggerAdapter = watermill.NopLogger{}
	if cfg.Logger != nil {
		// Wrap the slog logger with a custom watermill logger
		logger = watermill.NewSlogLogger(cfg.Logger)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return err
	}
	defer router.Close()

	// Launch goroutine to handle graceful shutdown
	go func() {
		// Wait for context cancellation signal
		<-ctx.Done()
		logger.Info("Shutting down command processor router", nil)
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
			Name:        "command-processor",
			MaxRequests: 2,
			Interval:    time.Second * 5,
			Timeout:     time.Second * 10,
		}).Middleware,
	)

	// Ignore errors from the command processor
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

	// Create a new command processor with the specified configuration
	processor, err := cqrs.NewCommandProcessorWithConfig(router, cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic:   genCommandSubscriberTopic,
		SubscriberConstructor:    commandSubscriberConstructor(cfg.SubscriberConstructor),
		Marshaler:                marshaler,
		Logger:                   logger,
		OnHandle:                 commandProcessorOnHandle(cfg.ErrorHandler),
		AckCommandHandlingErrors: false,
	})
	if err != nil {
		return err
	}

	if err := processor.AddHandlers(cmds...); err != nil {
		return err
	}

	return router.Run(ctx)
}

// CommandProcessorFunc is a function that wraps the CommandProcessor function to use it in the error group.
// It returns a function that can be used in the error group.
func CommandProcessorFunc(
	ctx context.Context,
	cfg CommandProcessorConfig,
	cmds ...CommandHandler,
) func() error {
	return func() error {
		return CommandProcessor(ctx, cfg, cmds...)
	}
}

// genCommandSubscriberTopic generates a subscribe topic for the command processor.
// It uses the command name as the topic.
func genCommandSubscriberTopic(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
	return params.CommandName, nil
}

// subscriberConstructor is a function that creates a new subscriber for the command processor.
func commandSubscriberConstructor(subscriber SubscriberConstructor) cqrs.CommandProcessorSubscriberConstructorFn {
	return func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
		return subscriber(params.Handler.HandlerName())
	}
}

// commandProcessorOnHandle is a function that is called after the command is handled.
func commandProcessorOnHandle(errorHandler func(ctx context.Context, err error) error) cqrs.CommandProcessorOnHandleFn {
	return func(params cqrs.CommandProcessorOnHandleParams) error {
		ctx := params.Message.Context()
		if err := params.Handler.Handle(ctx, params.Command); err != nil {
			if err := errorHandler(ctx, err); err != nil {
				return err
			}
		}

		return nil
	}
}

// defaultCommandProcessorConfig applies reasonable default values to any unset fields
// in the CommandProcessorConfig. The SubscriberConstructor field must still be provided
// as it's required for the command processor to function.
func defaultCommandProcessorConfig(cfg CommandProcessorConfig) CommandProcessorConfig {
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
		cfg.UnprocessableMessageTopic = "unprocessable-commands"
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
