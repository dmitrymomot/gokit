package cqrs

import (
	"context"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/dmitrymomot/gokit/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

// CommandHandler is an alias for cqrs.CommandHandler.
type CommandHandler = cqrs.CommandHandler

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
	redisClient redis.UniversalClient,
	errorHandler func(context.Context, error) error,
	cmds ...CommandHandler,
) error {
	logger := watermill.NewSlogLogger(slog.With(slog.String("component", "command-processor")))
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return err
	}
	defer router.Close()

	// Check if context is already cancelled before proceeding
	if err := ctx.Err(); err != nil {
		return err
	}

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
			MaxRetries:      10,
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
		middleware.Timeout(time.Second*30),

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

	// If error handler is not provided, use a default one that does nothing.
	if errorHandler == nil {
		errorHandler = func(ctx context.Context, err error) error { return nil }
	}

	processor, err := cqrs.NewCommandProcessorWithConfig(router, cqrs.CommandProcessorConfig{
		GenerateSubscribeTopic:   genCommandSubscriberTopic,
		SubscriberConstructor:    commandSubscriberConstructor(redisClient),
		Marshaler:                marshaler,
		Logger:                   logger,
		OnHandle:                 commandProcessorOnHandle(errorHandler),
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
	redisClient redis.UniversalClient,
	errorHandler func(context.Context, error) error,
	cmds ...CommandHandler,
) func() error {
	return func() error {
		defer slog.InfoContext(ctx, "Command processor stopped", "component", "command-processor")
		return CommandProcessor(ctx, redisClient, errorHandler, cmds...)
	}
}

// genCommandSubscriberTopic generates a subscribe topic for the command processor.
// It uses the command name as the topic.
func genCommandSubscriberTopic(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
	return params.CommandName, nil
}

// subscriberConstructor is a function that creates a new subscriber for the command processor.
func commandSubscriberConstructor(redisClient redis.UniversalClient) cqrs.CommandProcessorSubscriberConstructorFn {
	return func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
		return redisstream.NewSubscriber(
			redisstream.SubscriberConfig{
				Client:        redisClient,
				Unmarshaller:  redisstream.DefaultMarshallerUnmarshaller{},
				ConsumerGroup: params.HandlerName,
			},
			watermill.NewSlogLogger(slog.With(slog.String("component", "command_subscriber"))),
		)
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
