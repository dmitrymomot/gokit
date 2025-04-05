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

type WorkspaceCreatedEvent struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	CreatedAt     string `json:"created_at"`
}

func main() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		uuidStr := uuid.New().String()
		appName = "event-consumer-" + uuidStr[len(uuidStr)-6:]
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

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(cqrs.EventProcessorFunc(ctx, cqrs.NewRedisSubscriber(redisClient, log),
		func(ctx context.Context, err error) error {
			log.ErrorContext(ctx, "Event processing error", "error", err)
			return nil
		},
		cqrs.NewEventHandler(
			func(ctx context.Context, event *WorkspaceCreatedEvent) error {
				log.InfoContext(ctx, "Received event by handler 1", "event", event.WorkspaceName)
				return nil
			},
		),
		cqrs.NewEventHandler(
			func(ctx context.Context, event *WorkspaceCreatedEvent) error {
				log.InfoContext(ctx, "Received event by handler 2", "event", event.WorkspaceName)
				return nil
			},
		),
	))

	// Wait for the application to stop
	// This will block until the context is done
	if err := eg.Wait(); err != nil {
		slog.ErrorContext(ctx, "Application stopped with an error", "error", err)
		os.Exit(1)
	}
}
