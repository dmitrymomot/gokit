package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/dmitrymomot/gokit/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewLogger(t *testing.T) {
	t.Run("creates JSON logger with default config", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cfg := logger.Config{
			Output: buf,
			Level:  slog.LevelInfo,
		}
		
		log := logger.NewLogger(cfg)
		require.NotNil(t, log)
		
		log.Info("test message")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "INFO", logEntry["level"])
		assert.Equal(t, "test message", logEntry["msg"])
	})
	
	t.Run("creates Text logger", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cfg := logger.Config{
			Output: buf,
			Format: logger.FormatText,
			Level:  slog.LevelInfo,
		}
		
		log := logger.NewLogger(cfg)
		require.NotNil(t, log)
		
		log.Info("test message")
		
		logOutput := buf.String()
		assert.Contains(t, logOutput, "INFO")
		assert.Contains(t, logOutput, "test message")
	})
	
	t.Run("includes default attributes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cfg := logger.Config{
			Output: buf,
			DefaultAttrs: []slog.Attr{
				slog.String("service", "test-service"),
				slog.Int("version", 1),
			},
		}
		
		log := logger.NewLogger(cfg)
		require.NotNil(t, log)
		
		log.Info("test message")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "test-service", logEntry["service"])
		assert.Equal(t, float64(1), logEntry["version"]) // JSON unmarshals as float64
	})
	
	t.Run("extracts values from context", func(t *testing.T) {
		buf := &bytes.Buffer{}
		
		type contextKey string
		testKey := contextKey("test-key")
		
		cfg := logger.Config{
			Output: buf,
			ContextExtractors: []logger.ContextExtractor{
				func(ctx context.Context) (slog.Attr, bool) {
					if val := ctx.Value(testKey); val != nil {
						return slog.String("test_id", val.(string)), true
					}
					return slog.Attr{}, false
				},
			},
		}
		
		log := logger.NewLogger(cfg)
		require.NotNil(t, log)
		
		ctx := context.WithValue(context.Background(), testKey, "test-value")
		log.InfoContext(ctx, "test message")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "test-value", logEntry["test_id"])
	})
}

func TestSetAsDefault(t *testing.T) {
	t.Run("sets logger as default", func(t *testing.T) {
		buf := &bytes.Buffer{}
		cfg := logger.Config{
			Output: buf,
		}
		
		log := logger.NewLogger(cfg)
		logger.SetAsDefault(log)
		
		slog.Info("default logger test")
		
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "INFO", logEntry["level"])
		assert.Equal(t, "default logger test", logEntry["msg"])
	})
}
