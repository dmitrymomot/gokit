package logger

import (
	"context"
	"log/slog"
)

// ContextExtractor is a function that extracts a value from context and returns an attribute.
type ContextExtractor func(ctx context.Context) (slog.Attr, bool)

// LogHandlerDecorator decorates a slog.Handler with additional functionality 
// like adding context values as log attributes.
type LogHandlerDecorator struct {
	next       slog.Handler
	extractors []ContextExtractor
}

// NewLogHandlerDecorator creates a new LogHandlerDecorator with the specified options.
func NewLogHandlerDecorator(next slog.Handler, opts ...LogHandlerOption) *LogHandlerDecorator {
	h := &LogHandlerDecorator{
		next:       next,
		extractors: []ContextExtractor{},
	}
	
	for _, opt := range opts {
		opt(h)
	}
	
	return h
}

// LogHandlerOption is a function that configures a LogHandlerDecorator.
type LogHandlerOption func(*LogHandlerDecorator)

// WithContextValue adds a context extractor that gets a value from context with the specified key
// and adds it as an attribute with the given name.
func WithContextValue(name string, key any) LogHandlerOption {
	return func(h *LogHandlerDecorator) {
		h.extractors = append(h.extractors, func(ctx context.Context) (slog.Attr, bool) {
			if val := ctx.Value(key); val != nil {
				return slog.Any(name, val), true
			}
			return slog.Attr{}, false
		})
	}
}

// WithContextExtractor adds a custom context extractor function.
func WithContextExtractor(extractor ContextExtractor) LogHandlerOption {
	return func(h *LogHandlerDecorator) {
		h.extractors = append(h.extractors, extractor)
	}
}

// Enabled implements slog.Handler.
func (h *LogHandlerDecorator) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

// Handle implements slog.Handler.
func (h *LogHandlerDecorator) Handle(ctx context.Context, rec slog.Record) error {
	// Add context attributes to the log record
	for _, extractor := range h.extractors {
		if attr, ok := extractor(ctx); ok {
			rec.AddAttrs(attr)
		}
	}

	return h.next.Handle(ctx, rec)
}

// WithAttrs implements slog.Handler.
func (h *LogHandlerDecorator) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogHandlerDecorator{
		next:       h.next.WithAttrs(attrs),
		extractors: h.extractors,
	}
}

// WithGroup implements slog.Handler.
func (h *LogHandlerDecorator) WithGroup(name string) slog.Handler {
	return &LogHandlerDecorator{
		next:       h.next.WithGroup(name),
		extractors: h.extractors,
	}
}
