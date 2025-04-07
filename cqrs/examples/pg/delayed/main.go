package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitrymomot/gokit/config"
	"github.com/dmitrymomot/gokit/cqrs"
	"github.com/dmitrymomot/gokit/pg"
	"github.com/dmitrymomot/gokit/randomname"
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
		appName = "pg-example-" + uuidStr[len(uuidStr)-6:]
	}

	log := slog.With(slog.String("app", appName))

	// Create a root context that listens for OS interrupt signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Init PostgreSQL connection
	pgCfg, err := config.Load[pg.Config]()
	if err != nil {
		log.ErrorContext(ctx, "Failed to load config", "error", err)
		os.Exit(1)
	}
	db, err := pg.Connect(ctx, pgCfg)
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Init Redis publisher
	publisher, err := cqrs.NewDelayedPostgresPublisher(db, log)
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

	// Start the event consumer
	eg.Go(cqrs.EventProcessorFunc(
		ctx,
		cqrs.EventProcessorConfig{
			Logger:                log.With(slog.String("component", "event-processor")),
			SubscriberConstructor: cqrs.NewDelayedPostgresSubscriber(db, log),
			ErrorHandler: func(ctx context.Context, err error) error {
				log.ErrorContext(ctx, "Event processing error", "error", err)
				return nil
			},
		},
		cqrs.NewEventHandler(
			func(ctx context.Context, event *WorkspaceCreatedEvent) error {
				log.InfoContext(ctx, "Received event ", "workspace_name", event.WorkspaceName)
				return nil
			},
		),
	))

	// Start the event producer
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
				if err := eventBus.PublishWithDelay(ctx, event, time.Second*10); err != nil {
					log.ErrorContext(ctx, "Failed to publish event", "error", err)
				}
				log.InfoContext(ctx, "Published event", "workspace_name", event.WorkspaceName)
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
