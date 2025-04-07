package config

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload" // Load .env file automatically
)

// configCache provides a type-safe way to store and retrieve configuration
// instances using generics
type configCache struct {
	mu     sync.Mutex
	values map[string]any
	onces  map[string]*sync.Once
}

var (
	// globalCache is the singleton instance for caching configurations
	globalCache = &configCache{
		values: make(map[string]any),
		onces:  make(map[string]*sync.Once),
	}
)

// Load loads environment variables into the provided configuration struct.
// It ensures that each unique configuration type is only loaded once
// throughout the application lifecycle.
//
// The function parses environment variables into a struct based on field tags.
// If loading fails, an appropriate error will be returned (ErrParsingConfig).
// Once a configuration type is successfully loaded, subsequent calls for the same
// type will return the cached version.
//
// Example:
//
//	type DatabaseConfig struct {
//		Host     string `env:"DB_HOST" envDefault:"localhost"`
//		Port     int    `env:"DB_PORT" envDefault:"5432"`
//		Username string `env:"DB_USER,required"`
//		Password string `env:"DB_PASS,required"`
//	}
//
//	config, err := config.Load[DatabaseConfig]()
//	if err != nil {
//		// Handle error
//	}
func Load[T any]() (T, error) {
	var config T
	typeName := getTypeName[T]()

	// Single lock section to check cache and set up once
	globalCache.mu.Lock()

	// Try to retrieve from cache immediately if already loaded
	if cached, ok := globalCache.values[typeName]; ok {
		globalCache.mu.Unlock()
		return cached.(T), nil
	}

	// Get or create the once instance for this type
	once, exists := globalCache.onces[typeName]
	if !exists {
		once = new(sync.Once)
		globalCache.onces[typeName] = once
	}
	globalCache.mu.Unlock()

	// Error to be captured from the sync.Once execution
	var err error

	// Use sync.Once to ensure the config is parsed only once
	once.Do(func() {
		// Parse environment variables into a new instance of T
		if parseErr := env.Parse(&config); parseErr != nil {
			err = fmt.Errorf("%w: %v", ErrParsingConfig, parseErr)
			return
		}

		// Store the successfully parsed config in the cache
		globalCache.mu.Lock()
		globalCache.values[typeName] = config
		globalCache.mu.Unlock()
	})

	if err != nil {
		return config, err
	}

	// If there's no error, the config should be in the cache
	globalCache.mu.Lock()
	cached, ok := globalCache.values[typeName]
	globalCache.mu.Unlock()

	if ok {
		return cached.(T), nil
	}

	return config, ErrConfigNotLoaded
}

// MustLoad works like Load but panics if configuration loading fails.
// This is useful for configurations that are required for the application to start.
func MustLoad[T any]() T {
	config, err := Load[T]()
	if err != nil {
		panic(fmt.Sprintf("Failed to load required configuration: %v", err))
	}
	return config
}

// getTypeName returns a string identifier for the generic type T
func getTypeName[T any]() string {
	var zero T
	return reflect.TypeOf(zero).String()
}
