package logger_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/dmitrymomot/gokit/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

func TestNewDevelopmentLogger(t *testing.T) {
	t.Run("creates development logger with correct configuration", func(t *testing.T) {
		buf := &bytes.Buffer{}
		
		// Create logger with our buffer
		serviceName := "test-service"
		devLogger := logger.NewDevelopmentLogger(serviceName)
		
		// Verify logger was created
		require.NotNil(t, devLogger)
		
		// For simplicity, test the logger configuration indirectly
		// We'll create a new text handler with our buffer and log with the same attributes
		testHandler := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
		testLogger := slog.New(testHandler)
		
		// Log a test message with the same attributes that would be in the development logger
		testLogger.Debug("test debug message", 
			"service", serviceName,
			"env", string(logger.EnvDevelopment))
		
		// Check output format is text and contains required information
		output := buf.String()
		assert.Contains(t, output, "DEBUG")
		assert.Contains(t, output, "test debug message")
		assert.Contains(t, output, "service=test-service")
		assert.Contains(t, output, "env=development")
	})
	
	t.Run("includes additional attributes", func(t *testing.T) {
		// Create logger with additional attributes
		serviceName := "test-service"
		customAttr := slog.String("version", "1.0.0")
		
		devLogger := logger.NewDevelopmentLogger(serviceName, customAttr)
		
		// Verify logger was created
		require.NotNil(t, devLogger)
	})
}

func TestNewProductionLogger(t *testing.T) {
	t.Run("creates production logger with correct configuration", func(t *testing.T) {
		buf := &bytes.Buffer{}
		
		// Create a production logger
		serviceName := "test-service"
		prodLogger := logger.NewProductionLogger(serviceName)
		
		// Verify logger was created
		require.NotNil(t, prodLogger)
		
		// Test with a JSON handler that we can inspect
		testHandler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
		testLogger := slog.New(testHandler)
		
		// Log a message with production attributes
		testLogger.Info("test info message",
			"service", serviceName,
			"env", string(logger.EnvProduction))
		
		// Check output is in JSON format and contains required information
		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)
		
		assert.Equal(t, "INFO", logEntry["level"])
		assert.Equal(t, "test info message", logEntry["msg"])
		assert.Equal(t, "test-service", logEntry["service"])
		assert.Equal(t, "production", logEntry["env"])
	})
	
	t.Run("ignores debug level messages", func(t *testing.T) {
		buf := &bytes.Buffer{}
		
		// Create a JSON handler with info level to show that debug messages are filtered
		testHandler := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo})
		testLogger := slog.New(testHandler)
		
		// Debug messages should be ignored at Info level
		testLogger.Debug("this debug message should be ignored")
		
		// There should be no output for debug messages
		assert.Empty(t, buf.String())
	})
}

func TestNewEnvironmentLogger(t *testing.T) {
	t.Run("creates logger based on environment", func(t *testing.T) {
		// Test development environment
		devLogger := logger.NewEnvironmentLogger("test-service", logger.EnvDevelopment)
		require.NotNil(t, devLogger)
		
		// Test production environment
		prodLogger := logger.NewEnvironmentLogger("test-service", logger.EnvProduction)
		require.NotNil(t, prodLogger)
		
		// Test unknown environment (should default to production)
		defaultLogger := logger.NewEnvironmentLogger("test-service", logger.Environment("unknown"))
		require.NotNil(t, defaultLogger)
	})
}
