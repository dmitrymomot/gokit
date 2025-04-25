# Config Package

A type-safe configuration loader for Go applications using environment variables.

## Installation

```bash
go get github.com/dmitrymomot/gokit/config
```

## Overview

The `config` package provides a simple way to load typed configurations from environment variables. It uses generics for type safety and implements a singleton pattern to ensure each configuration is loaded only once.

## Features

- Type-safe configuration with Go generics
- Automatic .env file loading
- Thread-safe singleton implementation
- Comprehensive error handling
- Default values support
- Required field validation

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

For configurations required at startup:

```go
// Will panic if environment variables are missing or invalid
appConfig := config.MustLoad[AppConfig]()
```

## Environment Variable Tags

The package supports various tag options:

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

## Error Handling

The package defines specific error types:

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

## Caching Behavior

The `Load` function ensures each configuration type is loaded only once:

1. First call parses environment variables
2. Subsequent calls return the cached instance
3. Different configuration types are cached independently
