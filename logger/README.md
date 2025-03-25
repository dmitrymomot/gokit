# Logger

The logger package provides extensions and utilities for Go's standard `log/slog` package.

## Features

- `LogHandlerDecorator`: A slog.Handler wrapper that allows adding context values to log records
- Options pattern for flexible configuration
- Type-safe context value extraction

## Usage

### Basic Usage

```go
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dmitrymomot/gokit/logger"
)

func main() {
	// Create a base handler
	baseHandler := slog.NewJSONHandler(os.Stdout, nil)
	
	// Create a decorated handler with context extraction
	handler := logger.NewLogHandlerDecorator(
		baseHandler,
		logger.WithContextValue("request_id", requestIDKey),
		logger.WithContextValue("user_id", userIDKey),
	)
	
	// Create a logger with the decorated handler
	log := slog.New(handler)
	
	// Set as default logger
	slog.SetDefault(log)
}
```

### Using Context Values in Logs

```go
type contextKey string

var (
	requestIDKey = contextKey("request_id")
	userIDKey    = contextKey("user_id")
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Create a new context with values
	ctx := context.WithValue(r.Context(), requestIDKey, "req-123")
	ctx = context.WithValue(ctx, userIDKey, "user-456")
	
	// Use the context in logs
	slog.InfoContext(ctx, "Processing request")
	// Output: {"level":"INFO","msg":"Processing request","request_id":"req-123","user_id":"user-456"}
}
```

### Custom Context Extraction

```go
// Create a custom extractor
clientIPExtractor := func(ctx context.Context) (slog.Attr, bool) {
	if r, ok := ctx.Value(requestKey).(*http.Request); ok && r != nil {
		return slog.String("client_ip", r.RemoteAddr), true
	}
	return slog.Attr{}, false
}

// Use the custom extractor
handler := logger.NewLogHandlerDecorator(
	baseHandler,
	logger.WithContextExtractor(clientIPExtractor),
)
```

## Advanced Configuration

The `LogHandlerDecorator` can be configured with multiple options:

```go
handler := logger.NewLogHandlerDecorator(
	baseHandler,
	// Extract simple context values
	logger.WithContextValue("request_id", requestIDKey),
	logger.WithContextValue("user_id", userIDKey),
	
	// Use custom extractors for complex values
	logger.WithContextExtractor(func(ctx context.Context) (slog.Attr, bool) {
		if sess, ok := ctx.Value(sessionKey).(*Session); ok && sess != nil {
			return slog.Group("session", 
				slog.String("id", sess.ID),
				slog.Time("created_at", sess.CreatedAt),
			), true
		}
		return slog.Attr{}, false
	}),
)
```

## Logger Factory

The package provides a flexible logger factory to create and configure `slog.Logger` instances:

```go
// Create a logger with JSON output
logger := logger.NewLogger(logger.Config{
    Level:  slog.LevelDebug,
    Format: logger.FormatJSON,
    DefaultAttrs: []slog.Attr{
        slog.String("service", "api"),
        slog.String("version", "1.0.0"),
    },
})

// Create a logger with text output
textLogger := logger.NewLogger(logger.Config{
    Level:  slog.LevelInfo,
    Format: logger.FormatText,
    Output: os.Stderr, // send logs to stderr instead of stdout
})

// Set as default logger
logger.SetAsDefault(logger)
```

You can also configure context extraction in the factory:

```go
// Define context keys
type contextKey string
var requestIDKey = contextKey("request_id")

// Configure logger with context extraction
logger := logger.NewLogger(logger.Config{
    Level: slog.LevelInfo,
    ContextExtractors: []logger.ContextExtractor{
        func(ctx context.Context) (slog.Attr, bool) {
            if val := ctx.Value(requestIDKey); val != nil {
                return slog.String("request_id", val.(string)), true
            }
            return slog.Attr{}, false
        },
    },
})
```

## Environment-Specific Loggers

The package provides helper functions to create pre-configured loggers for different environments:

### Development Logger

```go
// Create a development logger with text format and debug level
logger := logger.NewDevelopmentLogger("my-service")

// Add additional attributes
logger := logger.NewDevelopmentLogger("my-service",
    slog.String("version", "1.0.0"),
    slog.Int("server_id", 42),
)

// Log using the development logger
logger.Debug("Starting server", "port", 8080)
```

### Production Logger

```go
// Create a production logger with JSON format and info level
logger := logger.NewProductionLogger("my-service")

// Add additional attributes
logger := logger.NewProductionLogger("my-service",
    slog.String("version", "1.0.0"),
    slog.String("region", "eu-west"),
)

// Log using the production logger
logger.Info("Server started", "port", 8080)
```

### Environment-Based Logger

```go
// Create a logger based on the environment
env := logger.EnvProduction
if os.Getenv("APP_ENV") == "development" {
    env = logger.EnvDevelopment
}

logger := logger.NewEnvironmentLogger("my-service", env)

// The environment determines the logger configuration automatically
// - Development: text format, debug level
// - Production: JSON format, info level
