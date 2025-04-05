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

// WorkspaceCreate is a command to create a new workspace
type WorkspaceCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func main() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		uuidStr := uuid.New().String()
		appName = "command-producer-" + uuidStr[len(uuidStr)-6:]
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

	// Create a command bus to send commands
	commandBus, err := cqrs.NewCommandBus(redisClient, log)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create command bus", "error", err)
		os.Exit(1)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.InfoContext(ctx, "Shutting down command producer")
				return nil
			case <-ticker.C:
				// Generate a random workspace name
				name := randomname.Generate(nil)
				
				// Create the command
				cmd := WorkspaceCreate{
					Name:        name,
					Description: "Workspace created at " + time.Now().Format(time.RFC3339),
				}
				
				// Send the command
				if err := commandBus.Send(ctx, cmd); err != nil {
					log.ErrorContext(ctx, "Failed to send command", "error", err)
				} else {
					log.InfoContext(ctx, "Sent WorkspaceCreate command", "workspace_name", name)
				}
			}
		}
	})

	// Wait for the application to stop
	// This will block until the context is done
	if err := eg.Wait(); err != nil {
		log.ErrorContext(ctx, "Application stopped with an error", "error", err)
		os.Exit(1)
	}
}
