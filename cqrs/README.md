# CQRS Package

The CQRS package provides a Command Query Responsibility Segregation (CQRS) implementation using Watermill and Redis. It includes functionality for command and event handling, error management, and message processing.

## Features

- Command Bus and Processor
- Event Bus and Processor
- Error Handling
- Redis Stream Integration
- Middleware Support (Retry, Circuit Breaker, Timeout)

## Types

### PublicError

Represents a public error response that can be sent to the client.

```go
type PublicError[Details any] struct {
    Code    int     `json:"code"`
    Error   string  `json:"error"`
    Reason  string  `json:"reason"`
    Details Details `json:"details"`
}
```

Example:

```go
err := NewPublicError(400, errors.New("invalid input"), "VALIDATION_ERROR", map[string]string{
    "field": "email",
    "message": "invalid email format",
})
```

### CommandHandler and EventHandler

Aliases for Watermill's CQRS handlers.

## Functions

### Command Bus and Handlers

```go
// Create a new command bus
bus, err := NewCommandBus(redisClient)
if err != nil {
    log.Fatal(err)
}

// Create a command handler
type CreateUserCommand struct {
    Name  string
    Email string
}

handler := NewCommandHandler(func(ctx context.Context, cmd *CreateUserCommand) error {
    // Handle command
    return nil
})

// Process commands
err = CommandProcessor(ctx, redisClient, errorHandler, handler)
if err != nil {
    log.Fatal(err)
}
```

### Event Bus and Handlers

```go
// Create a new event bus
bus, err := NewEventBus(redisClient)
if err != nil {
    log.Fatal(err)
}

// Create an event handler
type UserCreatedEvent struct {
    UserID string
    Name   string
}

handler := NewEventHandler(func(ctx context.Context, event *UserCreatedEvent) error {
    // Handle event
    return nil
})

// Process events
err = EventProcessor(ctx, redisClient, errorHandler, handler)
if err != nil {
    log.Fatal(err)
}
```

### Error Handling

```go
// Create an error message
errMsg := NewErrorMessage(
    errors.New("validation failed"),
    "VALIDATION_ERROR",
    "field", "email",
    "message", "invalid format",
)

// Create a processor error handler
errorHandler := ProcessorErrorsHandler(
    eventBus,
    func(err error) string {
        // Map errors to reasons
        switch err.(type) {
        case *ValidationError:
            return "VALIDATION_ERROR"
        default:
            return "UNKNOWN"
        }
    },
)
```

## Full Example

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/redis/go-redis/v9"
    "golang.org/x/sync/errgroup"
    "saas/pkg/cqrs"
)

func main() {
    // Create a root context that listens for OS interrupt signals
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Initialize Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // Create command and event buses
    commandBus, err := cqrs.NewCommandBus(redisClient)
    if err != nil {
        slog.Error("Failed to create command bus", "error", err)
        os.Exit(1)
    }

    eventBus, err := cqrs.NewEventBus(redisClient)
    if err != nil {
        slog.Error("Failed to create event bus", "error", err)
        os.Exit(1)
    }

    // Define command and event types
    type CreateUserCommand struct {
        Name  string
        Email string
    }

    type UserCreatedEvent struct {
        UserID string
        Name   string
    }

    // Create handlers
    commandHandler := cqrs.NewCommandHandler(func(ctx context.Context, cmd *CreateUserCommand) error {
        // Handle command logic
        return eventBus.Publish(ctx, &UserCreatedEvent{
            UserID: "123",
            Name:   cmd.Name,
        })
    })

    eventHandler := cqrs.NewEventHandler(func(ctx context.Context, event *UserCreatedEvent) error {
        slog.Info("User created", "userID", event.UserID, "name", event.Name)
        return nil
    })

    // Create error handler
    errorHandler := cqrs.ProcessorErrorsHandler(
        eventBus,
        func(err error) string {
            return "UNKNOWN" // Add your error mapping logic
        },
    )

    // Create an error group with the root context
    g, ctx := errgroup.WithContext(ctx)

    // Run command processor
    g.Go(func() error {
        return cqrs.CommandProcessor(
            ctx,
            redisClient,
            errorHandler,
            commandHandler,
        )
    })

    // Run event processor
    g.Go(func() error {
        return cqrs.EventProcessor(
            ctx,
            redisClient,
            errorHandler,
            eventHandler,
        )
    })

    // Example: Send a command in a separate goroutine
    g.Go(func() error {
        cmd := &CreateUserCommand{
            Name:  "John Doe",
            Email: "john@example.com",
        }
        return commandBus.Send(ctx, cmd)
    })

    // Wait for all goroutines to complete or context to be cancelled
    if err := g.Wait(); err != nil {
        slog.Error("Application stopped with error", "error", err)
        os.Exit(1)
    }
}
```

## Notes

- All processors include built-in middleware for retry, circuit breaking, and timeout handling
- Redis is used as the message broker
- Error handling includes support for public error responses and error event publishing

For more detailed information about specific components, please refer to the source code documentation.
