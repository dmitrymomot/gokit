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

The logger package provides helper functions to create loggers with predefined configurations suitable for different environments:

#### Development Logger

```go
// Basic usage
logger := logger.NewDevelopmentLogger("my-service")

// With additional attributes
logger := logger.NewDevelopmentLogger("my-service", 
    slog.String("version", "1.0.0"),
    slog.Int("server_id", 42))

// With context extractors
logger := logger.NewDevelopmentLoggerWithExtractors("my-service", 
    []logger.ContextExtractor{
        logger.WithContextValue("request_id", requestIDKey),
    })
```

Development configuration uses:
- Text format for better human readability
- Debug log level for more verbose output
- Standard output (os.Stdout)
- Service name and environment attributes

#### Production Logger

```go
// Basic usage
logger := logger.NewProductionLogger("my-service")

// With additional attributes
logger := logger.NewProductionLogger("my-service", 
    slog.String("version", "1.0.0"),
    slog.String("region", "eu-west"))

// With context extractors
logger := logger.NewProductionLoggerWithExtractors("my-service", 
    []logger.ContextExtractor{
        logger.WithContextValue("request_id", requestIDKey),
    })
```

Production configuration uses:
- JSON format for easier machine parsing
- Info log level to reduce noise
- Standard output (os.Stdout)
- Service name and environment attributes

#### Environment-based Logger

You can also create a logger based on a specified environment:

```go
// Basic usage
logger := logger.NewEnvironmentLogger("my-service", logger.EnvProduction)

// With additional attributes
logger := logger.NewEnvironmentLogger("my-service", 
    logger.EnvProduction,
    slog.String("version", "1.0.0"))

// With context extractors
logger := logger.NewEnvironmentLoggerWithExtractors(
    "my-service", 
    logger.EnvProduction,
    []logger.ContextExtractor{
        logger.WithContextValue("request_id", requestIDKey),
    })
```

### Context Extraction

Context extractors allow you to automatically extract values from the context in `*Context` logging methods and add them to log entries:

```go
// Define a context key
type requestIDKey struct{}

// Create a context extractor
requestIDExtractor := logger.WithContextValue("request_id", requestIDKey{})

// Create a logger with the extractor
log := logger.NewDevelopmentLoggerWithExtractors("my-service", []logger.ContextExtractor{requestIDExtractor})

// Use with context
ctx := context.WithValue(context.Background(), requestIDKey{}, "abc-123")
log.InfoContext(ctx, "Request started")
// Output: ... msg="Request started" ... request_id=abc-123
```

You can create multiple context extractors and combine them:

```go
// Define context keys
type userIDKey struct{}
type traceIDKey struct{}

// Create context extractors
extractors := []logger.ContextExtractor{
    logger.WithContextValue("user_id", userIDKey{}),
    logger.WithContextValue("trace_id", traceIDKey{}),
}

// Create a logger with multiple extractors
log := logger.NewProductionLoggerWithExtractors("my-service", extractors)
