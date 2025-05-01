# CQRS Package

A Command Query Responsibility Segregation pattern implementation with event sourcing support.

## Installation

```bash
go get github.com/dmitrymomot/gokit/cqrs
```

## Overview

The CQRS package provides a robust implementation of the Command Query Responsibility Segregation pattern for building message-driven applications. It offers type-safe command and event handling using Go generics, built on top of Watermill for message processing. The package is thread-safe, supports multiple message transports, and is designed for both synchronous and asynchronous communication patterns.

## Features

- Type-safe command and event handling with Go generics
- Multiple message transports (Redis, Kafka, Go channels, PostgreSQL)
- Middleware support for retry, timeout, and circuit breaker patterns
- Comprehensive error handling with customizable error filters
- Delayed message processing capabilities
- Poison queue support for unprocessable messages
- Thread-safe implementations suitable for concurrent processing

## Usage

### Basic Command Handling

```go
import (
    "context"
    "log/slog"
    
    "github.com/dmitrymomot/gokit/cqrs"
)

// 1. Define your command
type CreateUserCommand struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// 2. Create a command handler
handler := cqrs.NewCommandHandler(func(ctx context.Context, cmd *CreateUserCommand) error {
    // Handle the command - create user, validate, etc.
    return nil
})

// 3. Set up the command bus
redisClient := getRedisClient() // Your Redis client
logger := slog.Default()

publisher, err := cqrs.NewRedisPublisher(redisClient, logger)
if err != nil {
    // Handle error
}

commandBus, err := cqrs.NewCommandBus(publisher, logger)
if err != nil {
    // Handle error
}

// 4. Set up command processor
errorHandler := func(ctx context.Context, err error) error {
    logger.ErrorContext(ctx, "Command processing error", "error", err)
    return nil
}

// Start in a goroutine
go cqrs.CommandProcessorFunc(
    ctx,
    cqrs.CommandProcessorConfig{
        Logger:                logger,
        SubscriberConstructor: cqrs.NewRedisSubscriber(redisClient, logger),
        ErrorHandler:          errorHandler,
    },
    handler,
)()

// 5. Send commands
err = commandBus.Send(ctx, &CreateUserCommand{
    Name:  "John Doe",
    Email: "john@example.com",
})
```

### Basic Event Handling

```go
// 1. Define your event
type UserCreatedEvent struct {
    UserID string `json:"user_id"`
    Name   string `json:"name"`
}

// 2. Create an event handler
eventHandler := cqrs.NewEventHandler(func(ctx context.Context, event *UserCreatedEvent) error {
    // React to the event - send email, update projections, etc.
    return nil
})

// 3. Set up the event bus
eventBus, err := cqrs.NewEventBus(publisher, logger)
if err != nil {
    // Handle error
}

// 4. Set up event processor
go cqrs.EventProcessorFunc(
    ctx,
    cqrs.EventProcessorConfig{
        Logger:                logger,
        SubscriberConstructor: cqrs.NewRedisSubscriber(redisClient, logger),
        ErrorHandler:          errorHandler,
    },
    eventHandler,
)()

// 5. Publish events
err = eventBus.Publish(ctx, &UserCreatedEvent{
    UserID: "123",
    Name:   "John Doe",
})
```

### Using Different Message Publishers

```go
// Redis publisher (recommended for production)
redisPublisher, err := cqrs.NewRedisPublisher(redisClient, logger)

// Kafka publisher
kafkaPublisher, err := cqrs.NewKafkaPublisher([]string{"localhost:9092"}, logger)

// In-memory publisher (for testing)
goChannelPublisher := cqrs.NewGoChannelPublisher(logger)

// PostgreSQL publisher (for delayed processing)
pgPublisher, err := cqrs.NewPostgresPublisher(db, logger)
```

### Error Handling

```go
// Define custom error handler
config := cqrs.CommandProcessorConfig{
    // Other config...
    ErrorHandler: func(ctx context.Context, err error) error {
        // Log the error
        logger.ErrorContext(ctx, "Failed to process command", "error", err)
        
        // Determine whether to retry based on error type
        if errors.Is(err, SomeTransientError) {
            return err // Return the error to trigger retry
        }
        
        // Don't retry for permanent errors
        return nil
    },
    // Ignore specific errors
    ErrorsIgnore: []error{cqrs.ErrDuplicateMessage},
}
```

## Best Practices

1. **Message Design**:
   - Keep commands and events simple and focused
   - Use immutable data structures
   - Include correlation IDs for tracing

2. **Error Handling**:
   - Configure appropriate error handlers for commands and events
   - Use error filters to distinguish between transient and permanent failures
   - Set up poison queues for unprocessable messages

3. **Performance**:
   - Choose the appropriate message transport for your needs
   - Configure appropriate timeout values
   - Use middleware carefully as it adds overhead

4. **Testing**:
   - Use in-memory transport for unit tests
   - Test command and event handlers in isolation
   - Use the decorator pattern for cross-cutting concerns

## API Reference

### Configuration Types

```go
type CommandProcessorConfig struct {
    SubscriberConstructor            SubscriberConstructor
    Logger                          *slog.Logger
    Publisher                        message.Publisher
    ErrorHandler                     func(context.Context, error) error
    ErrorsIgnore                     []error
    UnprocessableMessageErrorFilter  func(error) bool
    UnprocessableMessageTopic        string
    PoisonQueueEnabled               bool
    RetryConfig                     *RetryConfig
    CircuitBreakerConfig            *CircuitBreakerConfig
}

type EventProcessorConfig struct {
    SubscriberConstructor            SubscriberConstructor
    Logger                          *slog.Logger
    Publisher                        message.Publisher
    ErrorHandler                     func(context.Context, error) error
    ErrorsIgnore                     []error
    UnprocessableMessageErrorFilter  func(error) bool
    UnprocessableMessageTopic        string
    PoisonQueueEnabled               bool
    RetryConfig                     *RetryConfig
    CircuitBreakerConfig            *CircuitBreakerConfig
}
```

### Publisher Functions

```go
func NewRedisPublisher(client redis.UniversalClient, logger *slog.Logger) (message.Publisher, error)
func NewKafkaPublisher(brokers []string, logger *slog.Logger) (message.Publisher, error)
func NewGoChannelPublisher(logger *slog.Logger) message.Publisher
func NewPostgresPublisher(db *sql.DB, logger *slog.Logger) (message.Publisher, error)
```

### Subscriber Functions

```go
func NewRedisSubscriber(client redis.UniversalClient, logger *slog.Logger) SubscriberConstructor
func NewKafkaSubscriber(brokers []string, logger *slog.Logger) SubscriberConstructor
func NewGoChannelSubscriber(logger *slog.Logger, pubSub *message.GoChannel) SubscriberConstructor
func NewPostgresSubscriber(db *sql.DB, logger *slog.Logger) SubscriberConstructor
```

### Command and Event Functions

```go
func NewCommandBus(publisher message.Publisher, logger *slog.Logger) (*CommandBus, error)
func NewEventBus(publisher message.Publisher, logger *slog.Logger) (*EventBus, error)
func NewCommandHandler[C any](handler func(context.Context, *C) error) *CommandHandler
func NewEventHandler[E any](handler func(context.Context, *E) error) *EventHandler
func CommandProcessorFunc(ctx context.Context, config CommandProcessorConfig, handlers ...*CommandHandler) func() error
func EventProcessorFunc(ctx context.Context, config EventProcessorConfig, handlers ...*EventHandler) func() error
```

### Decorator Functions

```go
func CommandHandlerMiddleware[C any](h CommandHandlerFunc[C], middlewares ...CommandHandlerMiddlewareFunc[C]) CommandHandlerFunc[C]
func EventHandlerMiddleware[E any](h EventHandlerFunc[E], middlewares ...EventHandlerMiddlewareFunc[E]) EventHandlerFunc[E]
```

### Error Types

```go
var ErrDuplicateMessage = errors.New("duplicate message")
var ErrMessagingInfrastructure = errors.New("messaging infrastructure error")
var ErrHandlerNotFound = errors.New("handler not found")
var ErrPublisherNotFound = errors.New("publisher not found")
var ErrSubscriberNotFound = errors.New("subscriber not found")
var ErrInvalidMessagePayload = errors.New("invalid message payload")
```

## Examples

The package includes example applications in the `examples` directory:

1. **Distributed examples**: Event/command producers and consumers using Redis
2. **PostgreSQL examples**: Delayed message processing and pub/sub patterns

To run the examples:

```bash
cd cqrs/examples/distributed
make docker  # Start Redis
make all     # Run all examples
