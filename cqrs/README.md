# CQRS Package

The CQRS package provides a Command Query Responsibility Segregation (CQRS) implementation using Watermill and Redis. It includes functionality for command and event handling, error management, and message processing.

## Features

- Command Bus and Processor
- Event Bus and Processor
- Multiple Message Publishers (Redis, Kafka, Go Channels)
- Error Handling with predefined error types
- Redis Stream Integration
- Middleware Support (Retry, Circuit Breaker, Timeout)
- Type-safe handler implementations using generics

## Installation

```bash
go get -u github.com/dmitrymomot/gokit/cqrs
```

## Usage Examples

### Message Publishers

The package provides several message publisher implementations:

```go
// Redis publisher (default for production use)
redisPublisher, err := cqrs.NewRedisPublisher(redisClient, log)
if err != nil {
    log.Fatal(err)
}

// Kafka publisher
kafkaPublisher, err := cqrs.NewKafkaPublisher([]string{"localhost:9092"}, log)
if err != nil {
    log.Fatal(err)
}

// In-memory publisher (ideal for testing)
goChannelPublisher := cqrs.NewGoChannelPublisher(log)
```

Each publisher also has a WithConfig variant for custom configurations:

```go
// Custom Redis publisher configuration
customRedisPublisher, err := cqrs.NewRedisPublisherWithConfig(
    redisClient,
    redisstream.PublisherConfig{
        Client:     redisClient,
        Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
        // Add custom settings here
    },
    log,
)
```

### Command Bus and Handlers

```go
// Create a message publisher
publisher, err := cqrs.NewRedisPublisher(redisClient, log)
if err != nil {
    log.Fatal(err)
}

// Create a new command bus with the publisher
commandBus, err := cqrs.NewCommandBus(publisher, log)
if err != nil {
    log.Fatal(err)
}

// Define a command type
type CreateUserCommand struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Create a command handler
handler := cqrs.NewCommandHandler(func(ctx context.Context, cmd *CreateUserCommand) error {
    // Handle command
    return nil
})

// Process commands with error handling
errorHandler := func(ctx context.Context, err error) error {
    log.ErrorContext(ctx, "Command processing error", "error", err)
    return nil
}

// Start command processor in an error group
eg.Go(cqrs.CommandProcessorFunc(ctx, redisClient, errorHandler, handler))

// Send a command
err = commandBus.Send(ctx, &CreateUserCommand{
    Name:  "John Doe",
    Email: "john@example.com",
})
if err != nil {
    log.ErrorContext(ctx, "Failed to send command", "error", err)
}

// Send a command with modified message (for advanced use cases)
err = commandBus.SendWithModifiedMessage(ctx, &CreateUserCommand{
    Name:  "John Doe",
    Email: "john@example.com",
}, func(msg *message.Message) error {
    msg.Metadata.Set("priority", "high")
    return nil
})
```

### Event Bus and Handlers

```go
// Create a new event bus
bus, err := cqrs.NewEventBus(redisClient, log)
if err != nil {
    log.Fatal(err)
}

// Create an event handler
type UserCreatedEvent struct {
    UserID string `json:"user_id"`
    Name   string `json:"name"`
}

handler := cqrs.NewEventHandler(func(ctx context.Context, event *UserCreatedEvent) error {
    // Handle event
    return nil
})

// Process events with error handling
errorHandler := func(ctx context.Context, err error) error {
    log.ErrorContext(ctx, "Event processing error", "error", err)
    return nil
}

// Start event processor in an error group
eg.Go(cqrs.EventProcessorFunc(ctx, redisClient, errorHandler, handler))
```

### Running Example Applications

The package includes fully functional example applications that demonstrate CQRS principles:

#### Event Examples
- **event_producer**: Publishes events to the event bus
- **event_consumer**: Subscribes to and processes events

#### Command Examples
- **command_producer**: Sends commands to the command bus
- **command_consumer**: Processes commands and emits events

To run the examples:

```bash
# Navigate to the example directory
cd cqrs/example

# Start Redis using Docker
make docker

# Run all examples concurrently
make all
```

Or run specific examples:

```bash
# Run command producer
make run-command-producer

# Run command consumer
make run-command-consumer
```

## Complete Example

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/dmitrymomot/gokit/cqrs"
    "github.com/dmitrymomot/gokit/redis"
    "github.com/google/uuid"
    "golang.org/x/sync/errgroup"
)

// Define command and event types
type CreateUserCommand struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserCreatedEvent struct {
    UserID string `json:"user_id"`
    Name   string `json:"name"`
}

func main() {
    // Create a logger
    log := slog.With(slog.String("app", "cqrs-example"))

    // Create a root context with cancellation
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    // Connect to Redis
    redisClient, err := redis.Connect(ctx, redis.Config{
        ConnectionURL:  "redis://localhost:6379/0",
        ConnectTimeout: time.Second * 30,
        RetryAttempts:  3,
    })
    if err != nil {
        log.ErrorContext(ctx, "Failed to connect to Redis", "error", err)
        os.Exit(1)
    }

    // Create command and event buses
    commandBus, err := cqrs.NewCommandBus(redisClient, log)
    if err != nil {
        log.ErrorContext(ctx, "Failed to create command bus", "error", err)
        os.Exit(1)
    }

    eventBus, err := cqrs.NewEventBus(redisClient, log)
    if err != nil {
        log.ErrorContext(ctx, "Failed to create event bus", "error", err)
        os.Exit(1)
    }

    // Create error handler
    errorHandler := func(ctx context.Context, err error) error {
        log.ErrorContext(ctx, "Processing error", "error", err)
        return nil
    }

    // Create an error group
    eg, ctx := errgroup.WithContext(ctx)

    // Run command processor
    eg.Go(cqrs.CommandProcessorFunc(
        ctx,
        redisClient,
        errorHandler,
        cqrs.NewCommandHandler(func(ctx context.Context, cmd *CreateUserCommand) error {
            userID := uuid.New().String()
            log.InfoContext(ctx, "Processing command", "user", cmd.Name)
            
            // Emit event after processing command
            return eventBus.Publish(ctx, &UserCreatedEvent{
                UserID: userID,
                Name:   cmd.Name,
            })
        }),
    ))

    // Run event processor
    eg.Go(cqrs.EventProcessorFunc(
        ctx,
        redisClient,
        errorHandler,
        cqrs.NewEventHandler(func(ctx context.Context, event *UserCreatedEvent) error {
            log.InfoContext(ctx, "Event processed", "user_id", event.UserID, "name", event.Name)
            return nil
        }),
    ))

    // Send a command
    eg.Go(func() error {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return nil
            case <-ticker.C:
                cmd := &CreateUserCommand{
                    Name:  "John Doe",
                    Email: "john@example.com",
                }
                if err := commandBus.Send(ctx, cmd); err != nil {
                    log.ErrorContext(ctx, "Failed to send command", "error", err)
                }
            }
        }
    })

    // Wait for all goroutines to complete
    if err := eg.Wait(); err != nil {
        log.ErrorContext(ctx, "Application stopped with an error", "error", err)
        os.Exit(1)
    }
}
```

## Notes

- All processors include built-in middleware for retry, circuit breaking, and timeout handling
- Multiple message publisher options (Redis, Kafka, Go Channels) for different use cases
- Command bus for sending commands to handlers
- Event bus for publishing events after state changes
- Thread-safe implementation for concurrent use
- Context-based cancellation for graceful shutdowns
- Example applications demonstrate working implementations

For more detailed information about specific components, please refer to the example applications and source code documentation.
