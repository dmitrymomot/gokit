# Config Package

A type-safe configuration loader for Go applications using environment variables.

## Installation

```bash
go get github.com/dmitrymomot/gokit/config
```

## Overview

The `config` package provides a simple way to load typed configurations from environment variables. It uses generics for type safety and implements a singleton pattern to ensure each configuration is loaded only once. All operations are thread-safe, allowing concurrent access from multiple goroutines.

## Features

- Type-safe configuration with Go generics
- Automatic .env file loading
- Thread-safe singleton implementation
- Comprehensive error handling
- Default values support
- Required field validation
- Environment variable expansion
- Support for various data types including slices and time.Duration

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/config"
)

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	Username string `env:"DB_USER,required"`
	Password string `env:"DB_PASS,required"`
}

func main() {
	// Load the configuration
	dbConfig, err := config.Load[DatabaseConfig]()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use the configuration
	fmt.Printf("Database connection: %s:%d\n", dbConfig.Host, dbConfig.Port)
	// Output: Database connection: localhost:5432 (assuming default values)
}
```

### Multiple Configurations

```go
// Server configuration
type ServerConfig struct {
	Port     int    `env:"SERVER_PORT" envDefault:"8080"`
	Host     string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

// Authentication configuration
type AuthConfig struct {
	JWTSecret     string `env:"JWT_SECRET,required"`
	TokenLifetime int    `env:"TOKEN_LIFETIME" envDefault:"3600"`
}

func initializeConfigs() {
	// Load different configurations independently
	serverCfg, err := config.Load[ServerConfig]()
	if err != nil {
		log.Fatalf("Server config error: %v", err)
	}
	
	authCfg, err := config.Load[AuthConfig]()
	if err != nil {
		log.Fatalf("Auth config error: %v", err)
	}
	
	fmt.Printf("Server running at %s:%d\n", serverCfg.Host, serverCfg.Port)
	// Output: Server running at 0.0.0.0:8080 (assuming default values)
	
	fmt.Printf("Token lifetime: %d seconds\n", authCfg.TokenLifetime)
	// Output: Token lifetime: 3600 seconds (assuming default value)
}
```

### Using MustLoad

For configurations required at startup:

```go
func init() {
	// Will panic if environment variables are missing or invalid
	appConfig := config.MustLoad[AppConfig]()
	fmt.Printf("Application configured with environment: %s\n", appConfig.Environment)
	// Output: Application configured with environment: development (assuming default value)
	
	// IMPORTANT: Only use MustLoad when a missing configuration should halt the application
}
```

### Complex Configuration Example

```go
type AppConfig struct {
	// Basic with default
	Port int `env:"PORT" envDefault:"8080"`
	
	// Required field
	APIKey string `env:"API_KEY,required"`
	
	// Lists with custom separator
	Hosts []string `env:"HOSTS" envSeparator:":"`
	
	// Duration parsing
	Timeout time.Duration `env:"TIMEOUT" envDefault:"30s"`
	
	// Environment variable expansion
	TempDir string `env:"TEMP_DIR,expand" envDefault:"${HOME}/tmp"`
	
	// Boolean flags
	Debug bool `env:"DEBUG" envDefault:"false"`
	
	// Nested configuration through embedding
	Database DatabaseConfig
}

func loadComplexConfig() {
	// Set environment variables for testing
	os.Setenv("API_KEY", "secret-api-key-123")
	os.Setenv("HOSTS", "host1:host2:host3")
	os.Setenv("DEBUG", "true")
	
	cfg, err := config.Load[AppConfig]()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	fmt.Printf("API Key: %s\n", cfg.APIKey)
	// Output: API Key: secret-api-key-123
	
	fmt.Printf("Hosts: %v\n", cfg.Hosts)
	// Output: Hosts: [host1 host2 host3]
	
	fmt.Printf("Timeout: %v\n", cfg.Timeout)
	// Output: Timeout: 30s
	
	fmt.Printf("Debug mode: %v\n", cfg.Debug)
	// Output: Debug mode: true
}
```

## Error Handling

The package defines specific error types that can be checked using `errors.Is()`:

```go
import (
	"errors"
	"fmt"
	"log"
	
	"github.com/dmitrymomot/gokit/config"
)

// Example 1: Handling specific error types
func handleConfigErrors() {
	// Simulate a configuration with required fields
	type RequiredConfig struct {
		APIKey string `env:"API_KEY,required"`
		Secret string `env:"SECRET,required"`
	}
	
	// Without setting the required environment variables
	cfg, err := config.Load[RequiredConfig]()
	if err != nil {
		switch {
		case errors.Is(err, config.ErrParsingConfig):
			fmt.Println("Configuration parsing error: missing or invalid fields")
			// Output: Configuration parsing error: missing or invalid fields
			
			// You might want to check specific parsing errors
			if err.Error() contains "API_KEY" {
				fmt.Println("Missing API key in environment")
				// Output: Missing API key in environment
			}
			
		case errors.Is(err, config.ErrConfigNotLoaded):
			fmt.Println("Configuration could not be loaded")
			
		default:
			fmt.Printf("Unexpected error: %v\n", err)
		}
		
		// In production, you might want to use default values or fail gracefully
		return
	}
	
	// Use the configuration...
}

// Example 2: Graceful error handling with fallback
func loadWithFallback() {
	type DatabaseConfig struct {
		Host     string `env:"DB_HOST" envDefault:"localhost"`
		Port     int    `env:"DB_PORT"`  // No default, could cause parsing error
	}
	
	// Try to load the configuration
	dbConfig, err := config.Load[DatabaseConfig]()
	if err != nil {
		// Log the error
		log.Printf("Warning: could not load database config: %v", err)
		
		// Use fallback values
		dbConfig = DatabaseConfig{
			Host: "localhost",
			Port: 5432,
		}
		
		fmt.Println("Using fallback database configuration")
		// Output: Using fallback database configuration
	}
	
	// Continue with the application using either loaded or fallback configuration
	fmt.Printf("Connecting to database at %s:%d\n", dbConfig.Host, dbConfig.Port)
	// Output: Connecting to database at localhost:5432
}
```

## Caching Behavior

The `Load` function ensures each configuration type is loaded only once, making it efficient and thread-safe:

```go
func demonstrateCaching() {
	// First call loads from environment
	cfg1, err := config.Load[ServerConfig]()
	if err != nil {
		log.Fatal(err)
	}
	
	// Second call returns the cached instance (no parsing overhead)
	cfg2, _ := config.Load[ServerConfig]()
	
	// Both variables reference the same instance
	fmt.Println("Same instance:", cfg1 == cfg2)
	// Output: Same instance: true
	
	// Different configuration types are cached independently
	dbCfg, _ := config.Load[DatabaseConfig]()
	// This is a different configuration type with its own cache
}
```

## Best Practices

1. **Configuration Structure**:
   - Define a separate struct for each logical configuration group
   - Use embedding to compose complex configurations
   - Keep configuration structs in a dedicated package
   - Document each field with comments

2. **Environment Variables**:
   - Use consistent naming conventions (e.g., `APP_DB_HOST`, `APP_REDIS_HOST`)
   - Prefer uppercase with underscores for environment variable names
   - Include service or app prefix to avoid name collisions

3. **Default Values**:
   - Always provide sensible default values for non-critical fields
   - Only make truly required fields `required`
   - Document default values in struct comments

4. **Error Handling**:
   - Use `Load` with proper error handling for libraries and components
   - Consider using `MustLoad` only for application startup when missing configuration is fatal
   - Implement fallback strategies for non-critical configurations

5. **Security**:
   - Never commit sensitive information in `.env` files to version control
   - Use separate `.env` files for different environments
   - Consider using a secure secret manager for production environments
   - Validate sensitive configuration values for proper format and strength

6. **Usage Patterns**:
   - Load configurations during initialization
   - Pass configuration as dependencies to components that need them
   - Avoid global configuration where possible
   - Consider implementing a configuration refresh mechanism for long-running services

## API Reference

### Types

```go
type Config interface{}
```
A generic interface representing a configuration structure.

### Functions

```go
func Load[T any]() (T, error)
```
Loads configuration of type T from environment variables. Returns a cached instance if already loaded, otherwise parses environment variables based on struct tags. Thread-safe.

```go
func MustLoad[T any]() T
```
Like `Load`, but panics if the configuration cannot be loaded. Use only during application initialization.

```go
func Reload[T any]() (T, error)
```
Forces reloading of configuration of type T, even if previously loaded.

```go
func MustReload[T any]() T
```
Like `Reload`, but panics if the configuration cannot be reloaded.

### Error Variables

```go
var ErrParsingConfig = errors.New("failed to parse configuration")
```
Returned when the configuration cannot be parsed due to missing required values or type conversion errors.

```go
var ErrConfigNotLoaded = errors.New("configuration not loaded")
```
Returned when attempting to access a configuration that hasn't been successfully loaded.

### Struct Tags

The package supports the following struct tags for configuration fields:

- `env:"NAME"` - Specifies the environment variable name
- `env:"NAME,required"` - Marks the field as required
- `env:"NAME,expand"` - Expands environment variables in the value
- `envDefault:"value"` - Provides a default value
- `envSeparator:"sep"` - Custom separator for slice types (default is comma)
