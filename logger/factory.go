package logger

import (
	"io"
	"log/slog"
	"os"
)

// Format represents the output format of the logger.
type Format string

const (
	// FormatJSON outputs logs in JSON format.
	FormatJSON Format = "json"
	// FormatText outputs logs in human-readable text format.
	FormatText Format = "text"
)

// Environment represents the application environment.
type Environment string

const (
	// EnvDevelopment represents a development environment.
	EnvDevelopment Environment = "development"
	// EnvProduction represents a production environment.
	EnvProduction Environment = "production"
)

// Config contains configuration for the logger.
type Config struct {
	// Level sets the minimum log level that will be logged.
	// Default is slog.LevelInfo if not specified.
	Level slog.Level
	
	// Format specifies the output format (json or text).
	// Default is FormatJSON if not specified.
	Format Format
	
	// Output is where the logs will be written to.
	// Default is os.Stdout if not specified.
	Output io.Writer
	
	// DefaultAttrs are attributes that will be included with every log message.
	DefaultAttrs []slog.Attr
	
	// HandlerOptions provides additional options for the slog handler.
	// If nil, default options with the specified Level will be used.
	HandlerOptions *slog.HandlerOptions
	
	// ContextExtractors specifies functions to extract values from context
	// and add them as log attributes.
	ContextExtractors []ContextExtractor
}

// NewLogger creates a new logger instance with the specified configuration.
// It returns a slog.Logger instance configured according to the provided options.
func NewLogger(cfg Config) *slog.Logger {
	// Set defaults for nil values
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	
	if cfg.Format == "" {
		cfg.Format = FormatJSON
	}
	
	// Create handler options, using provided options or creating new ones with the specified level
	handlerOpts := cfg.HandlerOptions
	if handlerOpts == nil {
		handlerOpts = &slog.HandlerOptions{Level: cfg.Level}
	}
	
	// Create the appropriate handler based on the format
	var handler slog.Handler
	switch cfg.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	case FormatText:
		handler = slog.NewTextHandler(cfg.Output, handlerOpts)
	default:
		// Default to JSON if an invalid format is specified
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	}
	
	// Add default attributes if provided
	if len(cfg.DefaultAttrs) > 0 {
		handler = handler.WithAttrs(cfg.DefaultAttrs)
	}
	
	// Create options for the log handler decorator
	decoratorOpts := make([]LogHandlerOption, 0, len(cfg.ContextExtractors))
	for _, extractor := range cfg.ContextExtractors {
		decoratorOpts = append(decoratorOpts, WithContextExtractor(extractor))
	}
	
	// Create the decorated handler
	decoratedHandler := NewLogHandlerDecorator(handler, decoratorOpts...)
	
	// Return the new logger
	return slog.New(decoratedHandler)
}

// SetAsDefault sets the provided logger as the default slog logger.
// This is a convenience function for slog.SetDefault(logger).
func SetAsDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// NewDevelopmentLogger creates a new logger with predefined configuration 
// suitable for development environments.
//
// Development configuration uses:
// - Text format for better human readability
// - Debug log level for more verbose output
// - Standard output (os.Stdout)
// - Service name and environment attributes
// - Optional context extractors for adding context values to logs
//
// Example:
//
//	// Basic usage
//	logger := logger.NewDevelopmentLogger("my-service")
//	
//	// With additional attributes
//	logger := logger.NewDevelopmentLogger("my-service", 
//	    slog.String("version", "1.0.0"),
//	    slog.Int("server_id", 42))
//	
//	// With context extractors
//	logger := logger.NewDevelopmentLoggerWithExtractors("my-service", 
//	    []logger.ContextExtractor{
//	        logger.WithContextValue("request_id", requestIDKey),
//	    })
func NewDevelopmentLogger(serviceName string, attrs ...slog.Attr) *slog.Logger {
	defaultAttrs := []slog.Attr{
		slog.String("service", serviceName),
		slog.String("env", string(EnvDevelopment)),
	}
	
	if len(attrs) > 0 {
		defaultAttrs = append(defaultAttrs, attrs...)
	}
	
	return NewLogger(Config{
		Level:        slog.LevelDebug,
		Format:       FormatText,
		Output:       os.Stdout,
		DefaultAttrs: defaultAttrs,
	})
}

// NewDevelopmentLoggerWithExtractors creates a new development logger with support for
// context extractors to add request context values to log entries.
func NewDevelopmentLoggerWithExtractors(serviceName string, extractors []ContextExtractor, attrs ...slog.Attr) *slog.Logger {
	defaultAttrs := []slog.Attr{
		slog.String("service", serviceName),
		slog.String("env", string(EnvDevelopment)),
	}
	
	if len(attrs) > 0 {
		defaultAttrs = append(defaultAttrs, attrs...)
	}
	
	return NewLogger(Config{
		Level:            slog.LevelDebug,
		Format:           FormatText,
		Output:           os.Stdout,
		DefaultAttrs:     defaultAttrs,
		ContextExtractors: extractors,
	})
}

// NewProductionLogger creates a new logger with predefined configuration 
// suitable for production environments.
//
// Production configuration uses:
// - JSON format for easier machine parsing
// - Info log level to reduce noise
// - Standard output (os.Stdout)
// - Service name and environment attributes
// - Optional context extractors for adding context values to logs
//
// Example:
//
//	// Basic usage
//	logger := logger.NewProductionLogger("my-service")
//	
//	// With additional attributes
//	logger := logger.NewProductionLogger("my-service", 
//	    slog.String("version", "1.0.0"),
//	    slog.String("region", "eu-west"))
//	
//	// With context extractors
//	logger := logger.NewProductionLoggerWithExtractors("my-service", 
//	    []logger.ContextExtractor{
//	        logger.WithContextValue("request_id", requestIDKey),
//	    })
func NewProductionLogger(serviceName string, attrs ...slog.Attr) *slog.Logger {
	defaultAttrs := []slog.Attr{
		slog.String("service", serviceName),
		slog.String("env", string(EnvProduction)),
	}
	
	if len(attrs) > 0 {
		defaultAttrs = append(defaultAttrs, attrs...)
	}
	
	return NewLogger(Config{
		Level:        slog.LevelInfo,
		Format:       FormatJSON,
		Output:       os.Stdout,
		DefaultAttrs: defaultAttrs,
	})
}

// NewProductionLoggerWithExtractors creates a new production logger with support for
// context extractors to add request context values to log entries.
func NewProductionLoggerWithExtractors(serviceName string, extractors []ContextExtractor, attrs ...slog.Attr) *slog.Logger {
	defaultAttrs := []slog.Attr{
		slog.String("service", serviceName),
		slog.String("env", string(EnvProduction)),
	}
	
	if len(attrs) > 0 {
		defaultAttrs = append(defaultAttrs, attrs...)
	}
	
	return NewLogger(Config{
		Level:            slog.LevelInfo,
		Format:           FormatJSON,
		Output:           os.Stdout,
		DefaultAttrs:     defaultAttrs,
		ContextExtractors: extractors,
	})
}

// NewEnvironmentLogger creates a new logger with predefined configuration 
// based on the specified environment.
//
// Example:
//
//	// Basic usage
//	logger := logger.NewEnvironmentLogger("my-service", logger.EnvProduction)
//	
//	// With additional attributes
//	logger := logger.NewEnvironmentLogger("my-service", 
//	    logger.EnvProduction,
//	    slog.String("version", "1.0.0"))
func NewEnvironmentLogger(serviceName string, env Environment, attrs ...slog.Attr) *slog.Logger {
	switch env {
	case EnvDevelopment:
		return NewDevelopmentLogger(serviceName, attrs...)
	case EnvProduction:
		return NewProductionLogger(serviceName, attrs...)
	default:
		return NewProductionLogger(serviceName, attrs...)
	}
}

// NewEnvironmentLoggerWithExtractors creates a new environment-specific logger
// with support for context extractors.
//
// Example:
//
//	logger := logger.NewEnvironmentLoggerWithExtractors(
//	    "my-service", 
//	    logger.EnvProduction,
//	    []logger.ContextExtractor{
//	        logger.WithContextValue("request_id", requestIDKey),
//	    })
func NewEnvironmentLoggerWithExtractors(serviceName string, env Environment, extractors []ContextExtractor, attrs ...slog.Attr) *slog.Logger {
	switch env {
	case EnvDevelopment:
		return NewDevelopmentLoggerWithExtractors(serviceName, extractors, attrs...)
	case EnvProduction:
		return NewProductionLoggerWithExtractors(serviceName, extractors, attrs...)
	default:
		return NewProductionLoggerWithExtractors(serviceName, extractors, attrs...)
	}
}
