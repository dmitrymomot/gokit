package main

import (
	"context"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmitrymomot/gokit/sse"
	"github.com/dmitrymomot/gokit/sse/bus"
	"github.com/dmitrymomot/templatex"
	"github.com/google/uuid"
)

func main() {
	// Create a root context that listens for OS interrupt signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize with options
	tmpl, err := templatex.New("templates/",
		templatex.WithExtensions(".gohtml"),
	)
	if err != nil {
		panic(err)
	}

	// Create a channel-based message bus
	msgBus := bus.NewChannelBus()

	// Create an SSE server with custom heartbeat
	sseServer := sse.NewServer(msgBus, sse.WithHeartbeat(15*time.Second))

	// Create an HTTP server
	http.HandleFunc("/events", sseServer.Handler(func(r *http.Request) string {
		// Extract topic from request, e.g., from query parameter
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			topic = "notification"
		}
		return topic
	}))

	// Handler to serve the HTML template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		if topic == "" {
			topic = "notification"
		}
		if err := tmpl.Render(r.Context(), w, "index", map[string]any{
			"Messages": renderNotification(r.Context(), tmpl, "init"),
			"Topic":    topic,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
			return
		}
	})

	// Start the HTTP server in a separate goroutine
	httpServer := &http.Server{
		Addr: "localhost:8080",
	}

	go func() {
		log.Println("Starting HTTP server on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start the notification publisher in a separate goroutine
	go func() {
		log.Println("Starting notification publisher...")
		publishNotifications(ctx, msgBus, tmpl)
	}()

	// Wait for the context to be canceled (e.g., via Ctrl+C)
	<-ctx.Done()
	stop()
	log.Println("Shutting down gracefully...")

	// Create a context with timeout for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server Shutdown Failed:%+v", err)
	}

	log.Println("Server exited properly")
}

func renderNotification(ctx context.Context, tmpl *templatex.Engine, templateName ...string) template.HTML {
	// Available templates
	templates := []string{"info", "warn", "danger", "success"}

	// Use provided template name if available, otherwise select randomly
	notificationTemplate := ""
	if len(templateName) > 0 && templateName[0] != "" {
		notificationTemplate = templateName[0]
	} else {
		// Select a random template
		randomIndex := rand.Intn(len(templates))
		notificationTemplate = templates[randomIndex]
	}

	// Render a notification message using the selected template
	notification, err := tmpl.RenderHTML(ctx, notificationTemplate, nil)
	if err != nil {
		log.Printf("Error rendering notification: %v", err)
		return "Error rendering notification"
	}
	return notification
}

func publishNotifications(ctx context.Context, msgBus sse.MessageBus, tmpl *templatex.Engine) {
	// Publish notifications every 5 seconds
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Render the notification HTML
			notificationHTML := renderNotification(ctx, tmpl)

			// Convert template.HTML to string to prevent JSON escaping
			htmlString := string(notificationHTML)

			// Prepare the event
			event := sse.Event{
				ID:    uuid.New().String(),
				Event: "notification",
				Data:  htmlString,
			}

			// Publish a notification message to all clients
			if err := msgBus.Publish(ctx, "notification", event); err != nil {
				log.Printf("Error publishing notification: %v", err)
			} else {
				log.Printf("Successfully published notification with ID: %s", event.ID)
			}
		case <-ctx.Done():
			// Context was canceled, perform graceful shutdown
			log.Println("Stopping notification publisher...")
			return
		}
	}
}
