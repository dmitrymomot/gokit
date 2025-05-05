# GoKit

A collection of Go packages and utilities for building robust applications.

## Installation

```sh
go get github.com/dmitrymomot/gokit
```

## Packages

### API Key

The `apikey` package provides API key generation, hashing, and validation functionality for building secure API authentication systems.

Key features:
- Cryptographically secure API key generation
- Time-ordered API key generation for sortability
- Secure key hashing for storage
- Scope-based authorization (in the companion `scopes` package)

[View the API Key package documentation](./apikey/README.md)

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
- Support for primitive types and slices

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
- Multiple message transport options (Redis, Kafka, in-memory channels, PostgreSQL)
- Comprehensive error handling
- Thread-safe operation for concurrent use

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
- Proper error handling with custom error types
- Configurable logging for diagnostics

[View the I18n package documentation](./i18n/README.md)

### JWT

The `jwt` package provides a minimalist JWT implementation with HMAC-SHA256 signing.

Key features:
- Token generation and validation
- Context utilities
- HTTP middleware with flexible token extraction strategies
- Security and implementation best practices

[View the JWT package documentation](./jwt/README.md)

### Logger

The `logger` package provides context-aware structured logging built on Go's standard log/slog package.

Key features:
- Development and production logger configurations
- Context value extraction
- Environment-specific configuration
- Thread-safe operation
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

### OAuth/Google

The `oauth/google` package provides Google OAuth authentication utilities.

[View the OAuth/Google package documentation](./oauth/google/README.md)

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

The `qrcode` package provides a simplified API for generating QR codes as PNG images or base64 data URIs for HTML embedding.

[View the QR Code package documentation](./qrcode/README.md)

### Random Name

The `randomname` package provides random name generation.

[View the Random Name package documentation](./randomname/README.md)

### Redis

The `redis` package provides a Redis client wrapper for key-value storage.

Key features:
- Simple key-value storage interface compatible with caching and session systems
- Support for any redis.UniversalClient implementation
- Key expiration support
- Proper error handling
- Thread-safe operations

[View the Redis package documentation](./redis/README.md)

### Sanitizer

The `sanitizer` package provides input sanitization utilities.

[View the Sanitizer package documentation](./sanitizer/README.md)

### Scopes

The `scopes` package provides functionality for handling authentication and authorization scopes in API systems.

Key features:
- Parsing and joining scope strings
- Hierarchical scope support with wildcard matching
- Scope validation and comparison utilities
- Customizable delimiters and separators
- Scope normalization for consistent handling

[View the Scopes package documentation](./scopes/README.md)

### Semantic Versioning

The `semver` package provides utilities for parsing, validating, and comparing semantic versions.

Key features:
- Parsing and validating semantic version strings
- Version comparison following SemVer 2.0.0 rules
- Version manipulation (incrementing, modifying components)
- Range checking between versions
- Support for major.minor.patch format with prerelease identifiers

[View the Semantic Versioning package documentation](./semver/README.md)

### Server-Sent Events

The `sse` package provides Server-Sent Events implementation.

[View the SSE package documentation](./sse/README.md)

### State Machine

The `statemachine` package provides a flexible state machine implementation with a fluent builder pattern.

Key features:
- Core interfaces for State, Event, Action, and Guard
- Fluent builder pattern for configuration
- Support for conditional transitions through Guards
- Support for side effects through Actions
- Thread-safety for concurrent use
- Custom error types for better error handling

[View the State Machine package documentation](./statemachine/README.md)

### Storage

The `storage` package provides storage interfaces and implementations.

[View the Storage package documentation](./storage/README.md)

### Token

The `token` package provides a lightweight token generation and validation library with HMAC-SHA256 signatures.

Key features:
- Type-safe payload handling with Go generics
- URL-safe Base64 encoding
- Minimal API with zero external dependencies
- Thread safety and constant-time signature comparison

[View the Token package documentation](./token/README.md)

### TOTP

The `totp` package provides Time-based One-Time Password implementation.

[View the TOTP package documentation](./totp/README.md)

### User Agent

The `useragent` package provides user agent parsing utilities.

[View the User Agent package documentation](./useragent/README.md)

### Utils

The `utils` package provides lightweight utility functions including pointer creation, slug generation, display name normalization, reflection utilities, and JSON formatting.

[View the Utils package documentation](./utils/README.md)

### Validator

The `validator` package provides a flexible, tag-based validation system for Go structs.

Key features:
- 30+ built-in validators
- Customizable error messages
- Support for nested structs
- Thread-safe operation for concurrent validation

[View the Validator package documentation](./validator/README.md)

### Webhook

The `webhook` package provides webhook sending with retry, logging, and parallel execution.

Key features:
- Thread-safe parallel webhook sending
- Retry mechanism for failed requests
- Structured logging with `log/slog`
- Decorator pattern for extending functionality
- Benchmarked for high-volume, concurrent scenarios

[View the Webhook package documentation](./webhook/README.md)

## License

This project is licensed under the [Apache License 2.0](LICENSE).

## Development

### Git Hooks

This repository includes git hooks to ensure code quality and consistency:

- **pre-push**: Automatically runs `go fmt`, `goimports` (if available), and tests before pushing commits

To install the git hooks, run:

```sh
./scripts/setup-git-hooks.sh
```

The hooks will ensure that:
- Code is properly formatted according to Go standards
- All tests pass before pushing commits
- Only Go 1.24+ is used for development