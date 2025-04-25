# Logger Package

Extends Go's standard `log/slog` with utilities for structured, context-aware logging.

## Installation

```bash
go get github.com/dmitrymomot/gokit/logger
```

## Overview

The `logger` package provides tools for enhancing Go's built-in `log/slog` with type-safe context value extraction, consistent configuration across environments, and simplified logger setup.

## Features

- Context-aware logging with automatic extraction from request context
- Predefined configurations for development and production environments
- JSON and text output formats with sensible defaults
- Type-safe context value extraction
- Simplified logger factory with flexible configuration
- Zero external dependencies (uses only standard library)

## Usage

### Basic Usage

```go
import (
	"log/slog"
	"os"
	"github.com/dmitrymomot/gokit/logger"
)

// Create a simple JSON logger
log := logger.NewLogger(logger.Config{
	Level:  slog.LevelInfo,
	Format: logger.FormatJSON,
})

// Set as default logger
logger.SetAsDefault(log)

// Use like standard slog
log.Info("Application started", "port", 8080)
```

### Environment-Based Loggers

```go
// Development logger (text format, debug level)
devLog := logger.NewDevelopmentLogger("my-service")

// Production logger (JSON format, info level)
prodLog := logger.NewProductionLogger("my-service")

// Or choose based on environment
isProduction := os.Getenv("ENV") == "production"
env := logger.EnvDevelopment
if isProduction {
	env = logger.EnvProduction
}
log := logger.NewEnvironmentLogger("my-service", env)
```

### Context Value Extraction

```go
import (
	"context"
	"net/http"
)

// Define context keys
type contextKey string
var (
	requestIDKey = contextKey("request_id")
	userIDKey    = contextKey("user_id")
)

// Create a logger with context extractors
log := logger.NewLogger(logger.Config{
	Level: slog.LevelInfo,
	Format: logger.FormatJSON,
	ContextExtractors: []logger.ContextExtractor{
		logger.WithContextValue("request_id", requestIDKey),
		logger.WithContextValue("user_id", userIDKey),
	},
})

// Create middleware to add values to context
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		ctx := context.WithValue(r.Context(), requestIDKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// In your handler
func handler(w http.ResponseWriter, r *http.Request) {
	// This log will automatically include request_id from context
	log.InfoContext(r.Context(), "Processing request")
}
```

### Custom Context Extractors

```go
// Extract complex objects from context
sessionExtractor := logger.WithContextExtractor(func(ctx context.Context) (slog.Attr, bool) {
	if sess, ok := ctx.Value(sessionKey).(*Session); ok && sess != nil {
		return slog.Group("session", 
			slog.String("id", sess.ID),
			slog.Time("created_at", sess.CreatedAt),
		), true
	}
	return slog.Attr{}, false
})

// Use the custom extractor
log := logger.NewLogger(logger.Config{
	Level: slog.LevelInfo,
	ContextExtractors: []logger.ContextExtractor{
		sessionExtractor,
	},
})
```

### Decorating Existing Handlers

```go
// Create a base handler
baseHandler := slog.NewJSONHandler(os.Stdout, nil)

// Decorate it with context extraction
decoratedHandler := logger.NewLogHandlerDecorator(
	baseHandler,
	logger.WithContextValue("request_id", requestIDKey),
	logger.WithContextValue("user_id", userIDKey),
)

// Create a logger with the decorated handler
log := slog.New(decoratedHandler)
```

## API Reference

### Logger Creation

- `NewLogger(cfg Config) *slog.Logger` - Create a logger with custom configuration
- `NewDevelopmentLogger(serviceName string, attrs ...slog.Attr) *slog.Logger` - Create a development-optimized logger
- `NewProductionLogger(serviceName string, attrs ...slog.Attr) *slog.Logger` - Create a production-optimized logger
- `NewEnvironmentLogger(serviceName string, env Environment, attrs ...slog.Attr) *slog.Logger` - Create a logger based on environment

### Context Extraction

- `WithContextValue(name string, key any) LogHandlerOption` - Extract value from context by key
- `WithContextExtractor(extractor ContextExtractor) LogHandlerOption` - Add custom extraction logic

### Configuration

- `Config` struct - Controls logger behavior with fields:
  - `Level` - Minimum log level (default: Info)
  - `Format` - Output format (JSON or text)
  - `Output` - Where to write logs (default: os.Stdout)
  - `DefaultAttrs` - Attributes included with every log
  - `ContextExtractors` - Functions to extract context values

### Handler Decoration

- `NewLogHandlerDecorator(next slog.Handler, opts ...LogHandlerOption) *LogHandlerDecorator` - Wrap a handler with additional functionality

## Error Handling

```go
// Config validation errors are returned when creating loggers
log, err := createLogger()
if err != nil {
    // Handle error
}

// Runtime logging errors are typically handled internally
// If needed, you can implement your own error handling in custom handlers
