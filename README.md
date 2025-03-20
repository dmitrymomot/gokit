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

## License

This project is licensed under the [Apache License 2.0](LICENSE).