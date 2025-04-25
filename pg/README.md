# PostgreSQL Package

Modern PostgreSQL database wrapper with connection management, migrations, and health checks.

## Installation

```bash
go get github.com/dmitrymomot/gokit/pg
```

## Overview

The `pg` package provides a robust interface to PostgreSQL databases using the pgx driver. It handles connection management, migrations with goose, health checking, and offers comprehensive error handling.

## Features

- Type-safe configuration with environment variable support
- Connection pooling with retry capabilities
- Database migrations via [goose](https://github.com/pressly/goose)
- Built-in health check functionality
- Comprehensive error handling with specialized error types
- Context-aware operations for proper cancellation
- Structured logging integration

## Usage

### Basic Connection

```go
import (
    "context"
    "github.com/dmitrymomot/gokit/pg"
)

// Create a PostgreSQL connection
db, err := pg.Connect(context.Background(), pg.Config{
    ConnectionString: "postgres://user:password@localhost:5432/dbname",
})
if err != nil {
    // Handle connection error
}
defer db.Close()

// Use the connection
// ...
```

### Environment-Based Configuration

```go
import (
    "github.com/dmitrymomot/gokit/config"
    "github.com/dmitrymomot/gokit/pg"
)

// Load PostgreSQL config from environment variables
cfg, err := config.Load[pg.Config]()
if err != nil {
    // Handle configuration error
}

// Connect with loaded config
db, err := pg.Connect(context.Background(), cfg)
```

### Database Migrations

```go
import (
    "context"
    "log/slog"
    "github.com/dmitrymomot/gokit/pg"
)

// Create logger (required for migrations)
logger := slog.Default()

// Connect to database
db, err := pg.Connect(context.Background(), pg.Config{
    ConnectionString: "postgres://user:password@localhost:5432/dbname",
    MigrationsPath:   "./migrations", // Path to migration files
})
if err != nil {
    // Handle error
}
defer db.Close()

// Run migrations to latest version
err = pg.Migrate(context.Background(), db, cfg, logger)
if err != nil {
    // Handle migration error
}
```

### Health Checking

```go
// Create a health check function
healthCheck := pg.Healthcheck(db)

// Use in HTTP handler
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    if err := healthCheck(r.Context()); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        w.Write([]byte("Database unavailable"))
        return
    }
    w.Write([]byte("Database healthy"))
})
```

### Error Handling

```go
// Check for specific error types
switch {
case pg.IsNotFoundError(err):
    // Handle record not found
case pg.IsDuplicateKeyError(err):
    // Handle duplicate key constraint violation
case pg.IsForeignKeyViolationError(err):
    // Handle foreign key constraint violation
case pg.IsTxClosedError(err):
    // Handle closed transaction error
default:
    // Handle other errors
}

// Check for specific connection errors
if errors.Is(err, pg.ErrFailedToOpenDBConnection) {
    // Handle connection failure
}
```

## Configuration

The `Config` struct provides comprehensive options for PostgreSQL connections:

```go
type Config struct {
    ConnectionString  string        `env:"PG_CONN_URL,required"`
    MaxOpenConns      int32         `env:"PG_MAX_OPEN_CONNS" envDefault:"10"`
    MaxIdleConns      int32         `env:"PG_MAX_IDLE_CONNS" envDefault:"5"`
    HealthCheckPeriod time.Duration `env:"PG_HEALTHCHECK_PERIOD" envDefault:"1m"`
    MaxConnIdleTime   time.Duration `env:"PG_MAX_CONN_IDLE_TIME" envDefault:"10m"`
    MaxConnLifetime   time.Duration `env:"PG_MAX_CONN_LIFETIME" envDefault:"30m"`
    RetryAttempts     int           `env:"PG_RETRY_ATTEMPTS" envDefault:"3"`
    RetryInterval     time.Duration `env:"PG_RETRY_INTERVAL" envDefault:"5s"`
    MigrationsPath    string        `env:"PG_MIGRATIONS_PATH" envDefault:"db/migrations"`
    MigrationsTable   string        `env:"PG_MIGRATIONS_TABLE" envDefault:"schema_migrations"`
}
```

## API Reference

### Connection Management

- `Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error)`: Create a new database connection pool

### Migrations

- `Migrate(ctx context.Context, pool *pgxpool.Pool, cfg Config, log logger) error`: Run database migrations

### Health Monitoring

- `Healthcheck(db *pgxpool.Pool) func(context.Context) error`: Create a health check function

### Error Helpers

- `IsNotFoundError(err error) bool`: Check if error is a "record not found" error
- `IsDuplicateKeyError(err error) bool`: Check if error is a duplicate key violation
- `IsForeignKeyViolationError(err error) bool`: Check if error is a foreign key violation
- `IsTxClosedError(err error) bool`: Check if error is a closed transaction error

## Best Practices

1. **Always close connections**: Use `defer db.Close()` to ensure proper connection cleanup
2. **Use context for timeouts**: Pass context with timeouts for operations that might take long
3. **Configure pool sizes appropriately**: Set MaxOpenConns and MaxIdleConns based on your workload
4. **Run migrations at startup**: Apply migrations when your application starts
5. **Check specific errors**: Use the provided error helper functions for better error handling
6. **Monitor connection health**: Implement the health check in your monitoring system
