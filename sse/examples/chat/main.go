package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
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

	// Initialize templatex
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

	// Create an HTTP server and handle SSE connections
	http.HandleFunc("/events", sseServer.Handler(func(r *http.Request) string {
		// Use room query parameter as topic
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "general"
		}
		return room
	}))

	// Handler to serve the HTML template
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get room and user from query parameters
		room := r.URL.Query().Get("room")
		if room == "" {
			room = "general"
		}

		user := r.URL.Query().Get("user")
		if user == "" {
			user = "Anonymous"
		}

		if err := tmpl.Render(r.Context(), w, "index", map[string]any{
			"InitMessage": renderMessage(r.Context(), tmpl, "init", "System", "Welcome to the chat!"),
			"Room":        room,
			"User":        user,
		}); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
			return
		}
	})

	// Handler for sending messages
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		room := r.FormValue("room")
		if room == "" {
			room = "general"
		}

		user := r.FormValue("user")
		if user == "" {
			user = "Anonymous"
		}

		message := r.FormValue("message")
		if message == "" {
			http.Error(w, "Message cannot be empty", http.StatusBadRequest)
			return
		}

		fmt.Printf("Received message from %s in room %s: %s\n", user, room, message)

		// Render the message HTML
		messageHTML := renderMessage(r.Context(), tmpl, "message", user, message)

		// Convert template.HTML to string
		htmlString := string(messageHTML)

		// Prepare the event
		event := sse.Event{
			ID:    uuid.New().String(),
			Event: "chat-message",
			Data:  htmlString,
		}

		// Publish the message to the specified room
		if err := msgBus.Publish(r.Context(), room, event); err != nil {
			log.Printf("Error publishing message: %v", err)
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}

		// Render the form template with the same room and user
		if err := tmpl.Render(r.Context(), w, "form", map[string]any{
			"Room": room,
			"User": user,
		}); err != nil {
			http.Error(w, "Failed to render form template", http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
			return
		}
	})

	// Start the HTTP server in a separate goroutine
	httpServer := &http.Server{
		Addr: "localhost:8080",
	}

	go func() {
		log.Println("Starting chat server on http://localhost:8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
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

// renderMessage renders a chat message using the specified template
func renderMessage(ctx context.Context, tmpl *templatex.Engine, templateName, user, message string) template.HTML {
	// Render a message using the selected template
	messageHTML, err := tmpl.RenderHTML(ctx, templateName, map[string]any{
		"User":    user,
		"Message": message,
		"Time":    time.Now(),
	})
	if err != nil {
		log.Printf("Error rendering message: %v", err)
		return "Error rendering message"
	}
	return messageHTML
}
