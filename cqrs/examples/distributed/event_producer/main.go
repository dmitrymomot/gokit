package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/randomname"
	"github.com/dmitrymomot/gokit/redis"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type WorkspaceCreatedEvent struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	CreatedAt     string `json:"created_at"`
}

func main() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		uuidStr := uuid.New().String()
		appName = "event-producer-" + uuidStr[len(uuidStr)-6:]
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

	// Init Redis publisher
	publisher, err := cqrs.NewRedisPublisher(redisClient, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create Redis publisher", "error", err)
		os.Exit(1)
	}

	eventBus, err := cqrs.NewEventBus(publisher, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create event bus", "error", err)
		os.Exit(1)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.InfoContext(ctx, "Shutting down event producer")
				return nil
			case <-ticker.C:
				// Simulate event publishing
				event := WorkspaceCreatedEvent{
					WorkspaceID:   uuid.New().String(),
					WorkspaceName: randomname.Generate(nil),
					CreatedAt:     time.Now().Format(time.RFC3339),
				}
				if err := eventBus.Publish(ctx, event); err != nil {
					log.ErrorContext(ctx, "Failed to publish event", "error", err)
				}
			}
		}
	})

	// Wait for the application to stop
	// This will block until the context is done
	if err := eg.Wait(); err != nil {
		slog.ErrorContext(ctx, "Application stopped with an error", "error", err)
		os.Exit(1)
	}
}
