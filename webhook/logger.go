package webhook

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"time"
)

// LoggerDecorator wraps a WebhookSender and adds logging functionality
type LoggerDecorator struct {
	sender       WebhookSender
	logger       *slog.Logger
	hideParams   bool
	maskedFields map[string]bool // fields to mask with asterisks
}

// LoggerOption configures the logger decorator
type LoggerOption func(*LoggerDecorator)

// WithHideParams prevents request parameters from being logged
// This is useful for security/privacy when parameters contain sensitive information
func WithHideParams() LoggerOption {
	return func(l *LoggerDecorator) {
		l.hideParams = true
	}
}

// WithMaskedFields specifies field names whose values should be masked with asterisks
// This is useful for logging that fields exist while hiding their sensitive values
func WithMaskedFields(fields ...string) LoggerOption {
	return func(l *LoggerDecorator) {
		if l.maskedFields == nil {
			l.maskedFields = make(map[string]bool)
		}
		for _, field := range fields {
			l.maskedFields[field] = true
		}
	}
}

// NewLoggerDecorator creates a new webhook sender with logging capabilities
func NewLoggerDecorator(sender WebhookSender, logger *slog.Logger, opts ...LoggerOption) WebhookSender {
	if logger == nil {
		logger = slog.Default()
	}

	ld := &LoggerDecorator{
		sender: sender,
		logger: logger,
	}

	// Apply options
	for _, opt := range opts {
		opt(ld)
	}

	return ld
}

// extractMethodFromOptions extracts the HTTP method from request options if present
func extractMethodFromOptions(opts []RequestOption) string {
	// Create a temporary requestOptions to apply the options
	o := &requestOptions{
		Method: "POST", // Default method
	}

	// Apply all options to determine the method
	for _, opt := range opts {
		opt(o)
	}

	return o.Method
}

// maskSensitiveParams creates a copy of params with sensitive fields masked
func (l *LoggerDecorator) maskSensitiveParams(params any) any {
	if params == nil || len(l.maskedFields) == 0 {
		return params
	}

	v := reflect.ValueOf(params)

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return params
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		// Create a copy of the map with masked values
		result := reflect.MakeMap(v.Type())
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			keyStr := key.String()
			value := iter.Value()

			// Check if this field should be masked
			if l.maskedFields[keyStr] {
				// Create a masked value of the appropriate type
				maskedValue := createMaskedValue(value)
				result.SetMapIndex(key, maskedValue)
			} else {
				result.SetMapIndex(key, value)
			}
		}
		return result.Interface()

	case reflect.Struct:
		// Create a copy of the struct with masked fields
		result := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			fieldName := field.Name
			// Check for json tag
			if tag := field.Tag.Get("json"); tag != "" && tag != "-" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
			}

			// Check if this field should be masked
			if l.maskedFields[fieldName] {
				// Create a masked value of the appropriate type
				maskedValue := createMaskedValue(v.Field(i))
				result.Field(i).Set(maskedValue)
			} else {
				result.Field(i).Set(v.Field(i))
			}
		}
		return result.Interface()
	}

	return params
}

// createMaskedValue creates a value of the same type as the input, but with masked content
func createMaskedValue(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.String:
		if v.Len() == 0 {
			return reflect.ValueOf("")
		}
		// Mask with asterisks but keep length indicators (first and last chars)
		length := v.Len()
		if length <= 2 {
			return reflect.ValueOf("**")
		}
		// For longer strings, keep first and last character
		return reflect.ValueOf(string(v.String()[0]) + strings.Repeat("*", length-2) + string(v.String()[length-1]))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Mask number but indicate it's a number
		return reflect.ValueOf("*****")
	case reflect.Float32, reflect.Float64:
		// Mask number but indicate it's a number
		return reflect.ValueOf("*****")
	case reflect.Bool:
		// Don't mask booleans (they're typically not sensitive)
		return v
	default:
		// For other types, create a new zero value
		return reflect.Zero(v.Type())
	}
}

// Send sends a webhook with logging before and after the request
func (l *LoggerDecorator) Send(ctx context.Context, url string, params any, opts ...RequestOption) (*Response, error) {
	// Clone the request options to avoid modifying the original
	requestOpts := make([]RequestOption, len(opts))
	copy(requestOpts, opts)

	// Extract the method from options
	method := extractMethodFromOptions(opts)

	startTime := time.Now()

	// Create log attributes
	logAttrs := []any{
		slog.String("url", url),
		slog.String("method", method),
	}

	// Handle parameters in logs
	if params != nil {
		if l.hideParams {
			// Don't log params at all
		} else if len(l.maskedFields) > 0 {
			// Mask sensitive fields
			maskedParams := l.maskSensitiveParams(params)
			logAttrs = append(logAttrs, slog.Any("params", maskedParams))
		} else {
			// Log all params
			logAttrs = append(logAttrs, slog.Any("params", params))
		}
	}

	// Log before sending
	l.logger.InfoContext(ctx, "Sending webhook request", logAttrs...)

	// Send the webhook
	resp, err := l.sender.Send(ctx, url, params, requestOpts...)

	duration := time.Since(startTime)

	if err != nil {
		// Log error
		l.logger.ErrorContext(ctx, "Webhook request failed",
			slog.String("url", url),
			slog.String("method", method),
			slog.Duration("duration", duration),
			slog.Any("error", err),
		)
		return nil, err
	}

	// Log success
	l.logger.InfoContext(ctx, "Webhook request completed",
		slog.String("url", url),
		slog.String("method", method),
		slog.Int("status_code", resp.StatusCode),
		slog.Bool("success", resp.IsSuccessful()),
		slog.Duration("duration", duration),
		slog.Int("body_size", len(resp.Body)),
	)

	return resp, nil
}
