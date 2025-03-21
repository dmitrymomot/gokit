# PGX Package

A Go package that provides a convenient wrapper around the [pgx](https://github.com/jackc/pgx) PostgreSQL driver with additional functionality for connection management, configuration, migrations, and health checking.

## Features

- Connection pooling with automatic retries
- Database health checking
- Migration support using [goose](https://github.com/pressly/goose)
- Error handling utilities
- Structured logging

## Installation

```bash
go get github.com/dmitrymomot/gokit/pg
```

## Configuration

The package uses environment variables for configuration. You can set these either in your environment or in a `.env` file:

```env
PG_CONN_URL=postgres://user:password@localhost:5432/dbname
PG_MAX_OPEN_CONNS=10
PG_MAX_IDLE_CONNS=5
PG_HEALTHCHECK_PERIOD=1m
PG_MAX_CONN_IDLE_TIME=10m
PG_MAX_CONN_LIFETIME=30m
PG_RETRY_ATTEMPTS=3
PG_RETRY_INTERVAL=5s
PG_MIGRATIONS_PATH=db/migrations
PG_MIGRATIONS_TABLE=schema_migrations
```

## Usage

### Basic Connection

```go
import "github.com/dmitrymomot/gokit/pg"

func main() {
    ctx := context.Background()

    // Create a new database connection
    db, err := pg.Connect(ctx, pg.Config{
        ConnectionString: "postgres://user:password@localhost:5432/dbname",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}
```

### Custom Configuration

```go
func main() {
    ctx := context.Background()

    cfg := pg.Config{
        ConnectionString: "postgres://user:password@localhost:5432/dbname",
        MaxOpenConns: 10,
        MaxIdleConns: 5,
        HealthCheckPeriod: time.Minute,
        MaxConnIdleTime: 10 * time.Minute,
        MaxConnLifetime: 30 * time.Minute,
        RetryAttempts: 3,
        RetryInterval: 5 * time.Second,
        MigrationsPath: "db/migrations",
        MigrationsTable: "schema_migrations",
    }

    db, err := pg.Connect(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
}
```

### Database Migrations

```go
import (
    "log/slog"

    "github.com/dmitrymomot/gokit/pg"
)

func main() {
    ctx := context.Background()
    
    // Create a logger
    logger := slog.Default()
    
    // Initialize database configuration
    cfg := pg.Config{
        ConnectionString: "postgres://user:password@localhost:5432/dbname",
        MigrationsPath: "./migrations",
    }
    
    // Initialize database connection
    pool, err := pg.Connect(ctx, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()

    // Run migrations
    err = pg.Migrate(ctx, pool, cfg, logger)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Health Checking

```go
func main() {
    ctx := context.Background()
    db, err := pg.Connect(ctx, pg.Config{
        ConnectionString: "postgres://user:password@localhost:5432/dbname",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Perform health check
    healthCheck := pg.Healthcheck(db)
    if err := healthCheck(ctx); err != nil {
        log.Printf("Database health check failed: %v", err)
    }
}
```

### Error Handling

```go
func handleDatabaseError(err error) {
    switch {
    case pg.IsNotFoundError(err):
        // Handle not found error
    case pg.IsDuplicateKeyError(err):
        // Handle duplicate key error
    case pg.IsForeignKeyViolationError(err):
        // Handle foreign key violation
    case pg.IsTxClosedError(err):
        // Handle closed transaction error
    default:
        // Handle other errors
    }
}
```

## Error Types

The package provides several error types for common database operations:

- `ErrFailedToOpenDBConnection`: Failed to establish database connection
- `ErrEmptyConnectionString`: No connection string provided
- `ErrHealthcheckFailed`: Database health check failed
- `ErrFailedToParseDBConfig`: Configuration parsing failed
- `ErrFailedToApplyMigrations`: Database migration failed
- `ErrMigrationsDirNotFound`: Migrations directory not found
- `ErrMigrationPathNotProvided`: Migration path not provided

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This package is part of the gokit project and is released under the MIT License. See the LICENSE file in the root directory for details.
