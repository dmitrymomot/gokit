# Logger Package

Extends Go's standard `log/slog` with utilities for structured, context-aware logging.

## Installation

```bash
go get github.com/dmitrymomot/gokit/logger
```

## Overview

The `logger` package provides tools for enhancing Go's built-in `log/slog` with type-safe context value extraction, consistent configuration across environments, and simplified logger setup. The implementation is thread-safe and suitable for concurrent use in production environments.

## Features

- Context-aware logging with automatic extraction from request context
- Predefined configurations for development and production environments
- JSON and text output formats with sensible defaults
- Type-safe context value extraction
- Simplified logger factory with flexible configuration
- Zero external dependencies (uses only standard library)
- Thread-safe implementation for concurrent usage
- Backward-compatible factory functions supporting context extractors

## Usage

### Basic Example

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
// Output (JSON): {"time":"2023-05-22T15:04:05Z","level":"INFO","msg":"Application started","port":8080}
```

### Environment-Based Loggers

```go
// Development logger (text format, debug level)
devLog := logger.NewDevelopmentLogger("my-service")
devLog.Debug("Server starting", "port", 8080)
// Output (text): DEBUG my-service Server starting port=8080

// Production logger (JSON format, info level)
prodLog := logger.NewProductionLogger("my-service")
prodLog.Info("Server started", "port", 8080)
// Output (JSON): {"time":"2023-05-22T15:04:05Z","level":"INFO","service":"my-service","msg":"Server started","port":8080}

// Or choose based on environment
isProduction := os.Getenv("ENV") == "production"
env := logger.EnvDevelopment
if isProduction {
	env = logger.EnvProduction
}
log := logger.NewEnvironmentLogger("my-service", env)
log.Info("Server ready")
// Output: depends on environment setting
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
	// Add userID to context
	ctx := context.WithValue(r.Context(), userIDKey, "user-123")
	
	// This log will automatically include request_id and user_id from context
	log.InfoContext(ctx, "Processing request")
	// Output (JSON): {"time":"2023-05-22T15:04:05Z","level":"INFO","msg":"Processing request","request_id":"abc-123","user_id":"user-123"}
}
```

### Custom Context Extractors

```go
// Define a session type
type Session struct {
	ID        string
	CreatedAt time.Time
	UserID    string
}

var sessionKey = contextKey("session")

// Extract complex objects from context
sessionExtractor := logger.WithContextExtractor(func(ctx context.Context) (slog.Attr, bool) {
	if sess, ok := ctx.Value(sessionKey).(*Session); ok && sess != nil {
		return slog.Group("session", 
			slog.String("id", sess.ID),
			slog.Time("created_at", sess.CreatedAt),
			slog.String("user_id", sess.UserID),
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

// Create context with session
ctx := context.Background()
session := &Session{
	ID:        "sess-456",
	CreatedAt: time.Now(),
	UserID:    "user-123",
}
ctx = context.WithValue(ctx, sessionKey, session)

// Log with session information
log.InfoContext(ctx, "User authenticated")
// Output (JSON): {"time":"2023-05-22T15:04:05Z","level":"INFO","msg":"User authenticated","session":{"id":"sess-456","created_at":"2023-05-22T15:04:05Z","user_id":"user-123"}}
```

### Using Context Extractor Factory Functions

```go
import (
	"context"
	"github.com/dmitrymomot/gokit/logger"
)

// Define your extractors
requestIDExtractor := logger.WithContextValue("request_id", requestIDKey)
userIDExtractor := logger.WithContextValue("user_id", userIDKey)

// Create logger with the new context-aware factory functions
log := logger.NewDevelopmentLoggerWithExtractors(
	"my-service",
	requestIDExtractor,
	userIDExtractor,
)

// Or for production
prodLog := logger.NewProductionLoggerWithExtractors(
	"my-service",
	requestIDExtractor,
	userIDExtractor,
)

// Or environment-based with extractors
envLog := logger.NewEnvironmentLoggerWithExtractors(
	"my-service",
	logger.EnvProduction,
	requestIDExtractor,
	userIDExtractor,
)

// Use the loggers with context
ctx := context.WithValue(context.Background(), requestIDKey, "req-123")
log.InfoContext(ctx, "Operation completed")
// Output includes the request_id from context automatically
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

// Use with context
ctx := context.Background()
ctx = context.WithValue(ctx, requestIDKey, "req-abc-123")
ctx = context.WithValue(ctx, userIDKey, "user-456")

log.InfoContext(ctx, "Request processed")
// Output (JSON): {"time":"2023-05-22T15:04:05Z","level":"INFO","msg":"Request processed","request_id":"req-abc-123","user_id":"user-456"}
```

### Error Handling

```go
import (
	"errors"
	"os"
	"log/slog"
	"github.com/dmitrymomot/gokit/logger"
)

// Example 1: Handling configuration errors
invalidConfig := logger.Config{
	// Invalid output destination
	Output: nil,
}

log, err := logger.NewLoggerWithError(invalidConfig)
if err != nil {
	// Handle configuration error
	fmt.Printf("Logger configuration error: %v\n", err)
	// Use a default fallback logger
	log = logger.NewDevelopmentLogger("fallback")
}

// Example 2: Logging errors properly
func processItem(id string) error {
	// Some operation that might fail
	if id == "" {
		return errors.New("invalid item ID")
	}
	return nil
}

// Log error with structured information
err = processItem("")
if err != nil {
	log.Error("Failed to process item", 
		"error", err, 
		"component", "processor",
		"severity", "high")
	// Output (text): ERROR Failed to process item error="invalid item ID" component=processor severity=high
}

// Example 3: Error grouping
errors := []error{
	errors.New("database connection failed"),
	errors.New("cache update failed"),
}

log.Error("Multiple errors occurred", 
	slog.Group("errors", 
		slog.Any("db", errors[0]),
		slog.Any("cache", errors[1]),
	),
	"operation", "system_startup")
// Output: detailed error log with grouped errors
```

## Best Practices

1. **Configuration**:
   - Use environment-appropriate logger (development/production)
   - Set appropriate log levels for each environment
   - Consider using JSON format in production for machine parsing
   - Include service name in all logs for multi-service environments

2. **Context Usage**:
   - Define clear context key names and types
   - Use consistent context keys across your application
   - Extract critical information that helps with request tracing
   - Consider using UUID request IDs for tracking requests across services

3. **Performance**:
   - Avoid expensive computations in log statements
   - Use appropriate log levels to minimize overhead
   - Consider sampling high-volume logs in production
   - Use level-checking before complex log preparation

4. **Security**:
   - Never log sensitive information (passwords, tokens, personal data)
   - Implement proper log rotation and retention policies
   - Consider log sanitization for user-supplied data
   - Be careful with error messages that might expose internal details

## API Reference

### Types

```go
type Config struct {
    Level             slog.Level
    Format            Format
    Output            io.Writer
    DefaultAttrs      []slog.Attr
    ContextExtractors []ContextExtractor
}
```
Configuration for logger creation.

```go
type Format string
```
Output format type (FormatJSON or FormatText).

```go
type Environment string
```
Environment type (EnvDevelopment or EnvProduction).

```go
type ContextExtractor func(context.Context) (slog.Attr, bool)
```
Function type for extracting values from context.

```go
type LogHandlerDecorator struct {
    // Contains unexported fields
}
```
Decorator for slog.Handler that adds context extraction.

```go
type LogHandlerOption func(*LogHandlerDecorator)
```
Option function for configuring a LogHandlerDecorator.

### Functions

```go
func NewLogger(cfg Config) *slog.Logger
```
Create a new logger with the given configuration.

```go
func NewLoggerWithError(cfg Config) (*slog.Logger, error)
```
Create a new logger with the given configuration, returning any validation errors.

```go
func NewDevelopmentLogger(serviceName string, attrs ...slog.Attr) *slog.Logger
```
Create a development-optimized logger (text format, debug level).

```go
func NewProductionLogger(serviceName string, attrs ...slog.Attr) *slog.Logger
```
Create a production-optimized logger (JSON format, info level).

```go
func NewEnvironmentLogger(serviceName string, env Environment, attrs ...slog.Attr) *slog.Logger
```
Create a logger based on the specified environment.

```go
func NewDevelopmentLoggerWithExtractors(serviceName string, extractors ...ContextExtractor) *slog.Logger
```
Create a development logger with context extractors.

```go
func NewProductionLoggerWithExtractors(serviceName string, extractors ...ContextExtractor) *slog.Logger
```
Create a production logger with context extractors.

```go
func NewEnvironmentLoggerWithExtractors(serviceName string, env Environment, extractors ...ContextExtractor) *slog.Logger
```
Create an environment-specific logger with context extractors.

```go
func SetAsDefault(log *slog.Logger)
```
Set the given logger as the default slog logger.

```go
func WithContextValue(name string, key any) ContextExtractor
```
Create a context extractor for a specific value by key.

```go
func WithContextExtractor(extractor ContextExtractor) LogHandlerOption
```
Add a custom context extractor to a handler decorator.

```go
func NewLogHandlerDecorator(next slog.Handler, opts ...LogHandlerOption) *LogHandlerDecorator
```
Create a new log handler decorator that wraps an existing handler.

### Constants

```go
const (
    FormatJSON Format = "json"
    FormatText Format = "text"
)
```
Output format constants.

```go
const (
    EnvDevelopment Environment = "development"
    EnvProduction  Environment = "production"
)
```
Environment constants.
