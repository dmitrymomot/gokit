# CQRS Package

A Command Query Responsibility Segregation (CQRS) implementation with event sourcing support.

## Installation

```bash
go get github.com/dmitrymomot/gokit/cqrs
```

## Overview

The `cqrs` package provides a robust implementation of the Command Query Responsibility Segregation pattern. It leverages Watermill for message handling and supports multiple message transport mechanisms including Redis, Kafka, and in-memory channels.

## Features

- Type-safe command and event handling with generics
- Multiple message transports (Redis, Kafka, Go channels)
- Middleware support (retry, timeout, circuit breaker)
- Comprehensive error handling
- Delayed message processing
- PostgreSQL-based message transport option
- Thread-safe implementations

## Usage

### Command Handling

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

// Start in a goroutine or error group
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

### Event Handling

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

### Available Message Publishers

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

## Advanced Features

### Middleware Support

```go
// Add retry middleware
retryMiddleware := middleware.Retry{
    MaxRetries:      3,
    InitialInterval: time.Second,
    Logger:          logger,
}

// Apply to command handler
handler = handler.Use(retryMiddleware.Middleware)
```

### Custom Message Metadata

```go
// Send command with metadata
err = commandBus.SendWithModifiedMessage(ctx, &CreateUserCommand{/*...*/}, 
    func(msg *message.Message) error {
        msg.Metadata.Set("priority", "high")
        msg.Metadata.Set("source", "api")
        return nil
    })
```

### Delayed Processing (with PostgreSQL)

```go
// Process message after 1 hour
pgPublisher.PublishWithDelay(ctx, &SomeCommand{/*...*/}, time.Hour)
```

## Example Applications

The package includes example applications in the `examples` directory:

1. **Distributed examples**: Event/command producers and consumers using Redis
2. **PostgreSQL examples**: Delayed message processing and pub/sub patterns

To run the examples:

```bash
cd cqrs/examples/distributed
make docker  # Start Redis
make all     # Run all examples
