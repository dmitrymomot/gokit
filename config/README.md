# Config Package

A lightweight, type-safe configuration loader for Go applications using environment variables.

## Overview

The `config` package provides a simple, generic interface for loading type-safe configurations from environment variables. It implements a singleton pattern to ensure each configuration type is loaded only once during the application lifecycle, improving performance and consistency.

## Features

- Type-safe configuration loading with Go generics
- Automatic .env file loading via godotenv
- Thread-safe singleton implementation
- Comprehensive error handling

## Installation

```bash
go get github.com/dmitrymomot/gokit/config
```

## Usage

### Basic Usage

Define your configuration struct with environment variable tags:

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

### Multiple Configuration Types

You can define and load multiple configuration types:

```go
type ServerConfig struct {
	Port      int    `env:"SERVER_PORT" envDefault:"8080"`
	Host      string `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Debug     bool   `env:"DEBUG" envDefault:"false"`
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
}

type AuthConfig struct {
	JWTSecret     string `env:"JWT_SECRET,required"`
	TokenLifetime int    `env:"TOKEN_LIFETIME" envDefault:"3600"` // in seconds
}

func main() {
	// Load server config
	serverConfig, err := config.Load[ServerConfig]()
	if err != nil {
		log.Fatalf("Failed to load server config: %v", err)
	}

	// Load auth config
	authConfig, err := config.Load[AuthConfig]()
	if err != nil {
		log.Fatalf("Failed to load auth config: %v", err)
	}

	// ... use configs
}
```

## Error Handling

The package provides specific error types for different failure scenarios:

- `ErrParsingConfig`: Returned when environment variables cannot be parsed into the config struct
- `ErrInvalidConfigType`: Returned when trying to access a config with an invalid type
- `ErrConfigNotLoaded`: Returned when attempting to access a config that hasn't been loaded

Proper error handling:

```go
config, err := config.Load[MyConfig]()
if err != nil {
	if errors.Is(err, config.ErrParsingConfig) {
		// Handle parsing error
	} else if errors.Is(err, config.ErrConfigNotLoaded) {
		// Handle not loaded error
	} else {
		// Handle other errors
	}
}
```

## Advanced Usage

### Custom Environment Tags

The package uses [github.com/caarlos0/env](https://github.com/caarlos0/env) under the hood, which provides extensive tagging options:

```go
type Config struct {
	Home         string        `env:"HOME"`
	Port         int           `env:"PORT" envDefault:"3000"`
	Password     string        `env:"PASSWORD,required"`
	IsProduction bool          `env:"PRODUCTION"`
	Hosts        []string      `env:"HOSTS" envSeparator:":"`
	Duration     time.Duration `env:"DURATION"`
	TempFolder   string        `env:"TEMP_FOLDER,expand" envDefault:"${HOME}/tmp"`
}
```

## Dependencies

- github.com/caarlos0/env/v11
- github.com/joho/godotenv/autoload
