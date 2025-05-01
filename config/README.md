# Config Package

A type-safe configuration loader for Go applications using environment variables.

## Installation

```bash
go get github.com/dmitrymomot/gokit/config
```

## Overview

The `config` package provides a simple way to load typed configurations from environment variables. It uses generics for type safety and implements a thread-safe singleton pattern to ensure each configuration is loaded only once. The package automatically loads variables from `.env` files and provides support for defaults, required fields, and custom formats.

## Features

- Type-safe configuration loading with Go generics
- Automatic `.env` file loading without explicit initialization
- Thread-safe singleton implementation for each config type
- Comprehensive error handling with specific error types
- Default values and required field validation
- Support for various data types including slices and time.Duration
- Environment variable expansion support

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

// Load different configurations independently
serverCfg, err := config.Load[ServerConfig]()
authCfg, err := config.Load[AuthConfig]()
```

### Using MustLoad

```go
// Will panic if environment variables are missing or invalid
appConfig := config.MustLoad[AppConfig]()

// Useful for configurations that are required at startup
```

### Error Handling

```go
config, err := config.Load[MyConfig]()
if err != nil {
	switch {
	case errors.Is(err, config.ErrParsingConfig):
		// Handle parsing error (missing required field, invalid format)
	case errors.Is(err, config.ErrConfigNotLoaded):
		// Handle not loaded error
	default:
		// Handle other errors
	}
}
```

## Best Practices

1. **Configuration Structure**:
   - Define separate configuration structs for different components
   - Group related settings within logical configuration types
   - Use clear, descriptive field and environment variable names

2. **Error Handling**:
   - Use `Load` for configurations that might fail at runtime
   - Use `MustLoad` only for configurations that are essential for startup
   - Check for specific error types when handling configuration errors

3. **Environment Variables**:
   - Use a consistent naming convention for environment variables
   - Prefix variables with component names to avoid collisions
   - Store sensitive information only in environment variables, not in code

4. **Default Values**:
   - Provide sensible defaults for non-critical configuration
   - Mark truly required fields with the `required` tag option

## API Reference

### Functions

```go
func Load[T any]() (T, error)
```
Loads environment variables into the configuration struct of type T. Ensures each configuration type is only loaded once and subsequent calls return the cached instance. Returns an error if parsing fails.

```go
func MustLoad[T any]() T
```
Like Load but panics if configuration loading fails. Useful for configurations that are required for the application to start.

### Environment Variable Tags

The package supports the following field tags:

```go
type Config struct {
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
}
```

### Error Types

```go
var ErrParsingConfig = errors.New("failed to parse environment variables into config")
var ErrInvalidConfigType = errors.New("invalid config type")
var ErrConfigNotLoaded = errors.New("configuration has not been loaded")
```
