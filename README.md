# GoKit

A collection of Go packages and utilities for building robust applications.

## Installation

```sh
go get github.com/dmitrymomot/gokit
```

## Packages

### Async

The `async` package provides a simple and type-safe way to run functions concurrently and wait for their results. It allows you to execute functions in parallel and retrieve their results once they are completed.

Key features:
- Type-safe async function execution with generics
- Context cancellation support
- Waiting with timeout
- Checking completion status without blocking
- Waiting for multiple futures to complete
- Waiting for any future to complete

[View the Async package documentation](./async/README.md)

### Binder

The `binder` package provides HTTP request data binding to Go structs with tag-based field mapping.

Key features:
- Automatic binding based on content type
- JSON request body binding
- Form data binding
- Query parameters binding
- Tag-based field mapping

[View the Binder package documentation](./binder/README.md)

### Cache

The `cache` package provides caching functionality for Go applications.

[View the Cache package documentation](./cache/README.md)

### Config

The `config` package provides type-safe configuration loading using environment variables.

[View the Config package documentation](./config/README.md)

### CQRS

The `cqrs` package implements Command Query Responsibility Segregation pattern for Go applications.

Key features:
- Command handling and dispatching
- Event handling and publishing
- Redis-based messaging infrastructure

[View the CQRS package documentation](./cqrs/README.md)

### Feature

The `feature` package provides feature flag functionality for Go applications.

[View the Feature package documentation](./feature/README.md)

### I18n

The `i18n` package provides internationalization support with middleware for HTTP servers.

Key features:
- Automatic language detection from HTTP requests
- Accept-Language header parsing
- Custom language extractors
- Default language fallback mechanism

[View the I18n package documentation](./i18n/README.md)

### JWT

The `jwt` package provides JSON Web Token handling utilities.

[View the JWT package documentation](./jwt/README.md)

### Logger

The `logger` package provides structured logging with context extractors.

Key features:
- Development and production logger configurations
- Context value extraction
- Integration with Go's `log/slog` package
- Custom attribute support

[View the Logger package documentation](./logger/README.md)

### Mailer

The `mailer` package provides email sending utilities.

[View the Mailer package documentation](./mailer/README.md)

### MongoDB

The `mongo` package provides a MongoDB client wrapper with configuration and health check support.

Key features:
- Client management
- Environment-based configuration
- Health check functionality
- Proper error handling

[View the MongoDB package documentation](./mongo/README.md)

### OAuth

The `oauth` package provides OAuth authentication utilities.

[View the OAuth package documentation](./oauth/README.md)

### OpenSearch

The `opensearch` package provides an OpenSearch client wrapper.

[View the OpenSearch package documentation](./opensearch/README.md)

### Pagination

The `pagination` package provides tools for implementing pagination in web applications.

[View the Pagination package documentation](./pagination/README.md)

### PostgreSQL

The `pg` package provides PostgreSQL database utilities.

[View the PostgreSQL package documentation](./pg/README.md)

### Privacy

The `privacy` package provides utilities for masking sensitive data like emails, phone numbers, and credit cards.

[View the Privacy package documentation](./privacy/README.md)

### QR Code

The `qrcode` package provides QR code generation utilities.

[View the QR Code package documentation](./qrcode/README.md)

### Random Name

The `randomname` package provides random name generation.

[View the Random Name package documentation](./randomname/README.md)

### RBAC

The `rbac` package provides Role-Based Access Control functionality.

[View the RBAC package documentation](./rbac/README.md)

### Redis

The `redis` package provides a Redis client wrapper for key-value storage.

Key features:
- Simple key-value storage interface
- Support for various Redis client implementations
- Key expiration support
- Proper error handling

[View the Redis package documentation](./redis/README.md)

### Router

The `router` package provides an HTTP router with middleware support.

[View the Router package documentation](./router/README.md)

### Sanitizer

The `sanitizer` package provides input sanitization utilities.

[View the Sanitizer package documentation](./sanitizer/README.md)

### Semantic Versioning

The `semver` package provides utilities for parsing, validating, and comparing semantic versions.

Key features:
- Parsing and validating semantic version strings
- Version comparison following SemVer 2.0.0 rules
- Version manipulation (incrementing, modifying components)
- Range checking between versions

[View the Semantic Versioning package documentation](./semver/README.md)

### Server-Sent Events

The `sse` package provides Server-Sent Events implementation.

[View the SSE package documentation](./sse/README.md)

### State Machine

The `statemachine` package provides a flexible state machine implementation with a fluent builder pattern.

Key features:
- Core interfaces for State, Event, Action, and Guard
- Fluent builder pattern for configuration
- Support for conditional transitions
- Support for side effects through Actions
- Thread-safety for concurrent use

[View the State Machine package documentation](./statemachine/README.md)

### Storage

The `storage` package provides storage interfaces and implementations.

[View the Storage package documentation](./storage/README.md)

### Token

The `token` package provides token generation and validation utilities.

[View the Token package documentation](./token/README.md)

### TOTP

The `totp` package provides Time-based One-Time Password implementation.

[View the TOTP package documentation](./totp/README.md)

### User Agent

The `useragent` package provides user agent parsing utilities.

[View the User Agent package documentation](./useragent/README.md)

### Utils

The `utils` package provides various utility functions.

[View the Utils package documentation](./utils/README.md)

### Validator

The `validator` package provides struct validation with custom rules support.

[View the Validator package documentation](./validator/README.md)

### Webhook

The `webhook` package provides webhook sending with retry, logging, and parallel execution.

Key features:
- Thread-safe parallel webhook sending
- Retry mechanism for failed requests
- Structured logging with `log/slog`
- Decorator pattern for extending functionality

[View the Webhook package documentation](./webhook/README.md)

## License

This project is licensed under the [Apache License 2.0](LICENSE).