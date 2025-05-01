# Logger Package

A structured, context-aware logging package built on top of Go's standard `log/slog`.

## Installation

```bash
go get github.com/dmitrymomot/gokit/logger
```

## Overview

The `logger` package enhances Go's built-in `log/slog` with type-safe context value extraction, consistent configuration across environments, and simplified logger setup. It is designed to be thread-safe for concurrent usage in production environments while maintaining zero external dependencies.

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
// Returns: {"time":"2023-05-22T15:04:05Z","level":"INFO","msg":"Application started","port":8080}
```

### Environment-Based Loggers

```go
// Development logger (text format, debug level)
devLog := logger.NewDevelopmentLogger("my-service")
devLog.Debug("Server starting", "port", 8080)
// Returns: DEBUG my-service Server starting port=8080

// Production logger (JSON format, info level)
prodLog := logger.NewProductionLogger("my-service")
prodLog.Info("Server started", "port", 8080)
// Returns: {"time":"2023-05-22T15:04:05Z","level":"INFO","service":"my-service","msg":"Server started","port":8080}

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

// In your HTTP handler
func handler(w http.ResponseWriter, r *http.Request) {
	// Add values to context
	ctx := context.WithValue(r.Context(), userIDKey, "user-123")
	
	// This log will automatically include request_id and user_id from context
	log.InfoContext(ctx, "Processing request")
	// Returns: {"time":"2023-05-22T15:04:05Z","level":"INFO","msg":"Processing request","user_id":"user-123"}
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

// Create a custom extractor for complex objects
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
```

### Error Handling

```go
// Handle errors properly in logs
if err != nil {
	log.Error("Operation failed", 
		"error", err,
		"component", "database",
		"operation", "query")
}

// Log multiple errors together
errors := []error{dbErr, cacheErr}
if len(errors) > 0 {
	log.Error("Multiple failures occurred",
		slog.Group("errors",
			slog.Any("db", errors[0]),
			slog.Any("cache", errors[1]),
		),
		"operation", "system_startup")
}
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
    HandlerOptions    *slog.HandlerOptions
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
func SetAsDefault(logger *slog.Logger)
```
Set the given logger as the default slog logger.

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
func NewDevelopmentLoggerWithExtractors(serviceName string, extractors []ContextExtractor, attrs ...slog.Attr) *slog.Logger
```
Create a development logger with context extractors.

```go
func NewProductionLoggerWithExtractors(serviceName string, extractors []ContextExtractor, attrs ...slog.Attr) *slog.Logger
```
Create a production logger with context extractors.

```go
func NewEnvironmentLoggerWithExtractors(serviceName string, env Environment, extractors []ContextExtractor, attrs ...slog.Attr) *slog.Logger
```
Create an environment-specific logger with context extractors.

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
