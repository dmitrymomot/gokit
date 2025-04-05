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

// WorkspaceCreate is a command to create a new workspace
type WorkspaceCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// WorkspaceCreatedEvent is an event that is emitted when a workspace is created
type WorkspaceCreatedEvent struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	CreatedAt     string `json:"created_at"`
}

func main() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		uuidStr := uuid.New().String()
		appName = "command-consumer-" + uuidStr[len(uuidStr)-6:]
	}

	log := slog.With(slog.String("app", appName))

	// Create a root context that listens for OS interrupt signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	redisClient, err := redis.Connect(ctx, redis.Config{
		ConnectionURL:  "redis://localhost:6379/0",
		ConnectTimeout: time.Second * 30,
		RetryAttempts:  3,
	})
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	// Create an event bus to publish events
	eventBus, err := cqrs.NewEventBus(redisClient, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create event bus", "error", err)
		os.Exit(1)
	}

	// Create command handler for WorkspaceCreate
	workspaceCreateHandler := func(ctx context.Context, cmd *WorkspaceCreate) error {
		// Log received command
		log.InfoContext(ctx, "Received WorkspaceCreate command", "workspace_name", cmd.Name)

		// Create a new workspace ID
		workspaceID := uuid.New().String()

		// Emit WorkspaceCreatedEvent
		event := WorkspaceCreatedEvent{
			WorkspaceID:   workspaceID,
			WorkspaceName: cmd.Name,
			CreatedAt:     time.Now().Format(time.RFC3339),
		}

		// Publish the event
		if err := eventBus.Publish(ctx, event); err != nil {
			log.ErrorContext(ctx, "Failed to publish WorkspaceCreatedEvent", "error", err)
			return err
		}

		log.InfoContext(ctx, "Published WorkspaceCreatedEvent",
			"workspace_id", workspaceID,
			"workspace_name", cmd.Name)

		return nil
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(cqrs.CommandProcessorFunc(ctx, redisClient,
		func(ctx context.Context, err error) error {
			log.ErrorContext(ctx, "Command processing error", "error", err)
			return nil
		},
		cqrs.NewCommandHandler(workspaceCreateHandler),
	))

	// Wait for the application to stop
	// This will block until the context is done
	if err := eg.Wait(); err != nil {
		log.ErrorContext(ctx, "Application stopped with an error", "error", err)
		os.Exit(1)
	}
}
