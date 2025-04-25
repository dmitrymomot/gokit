# Async Package

A type-safe asynchronous execution package using Go generics.

## Installation

```bash
go get github.com/dmitrymomot/gokit/async
```

## Overview

The `async` package provides a clean, type-safe way to execute functions asynchronously in Go. It uses generics to ensure type safety throughout the asynchronous operation lifecycle.

## Features

- Type-safe asynchronous function execution with generic types
- Future pattern implementation with await functionality
- Timeout support for asynchronous operations
- Utility functions for working with multiple futures
- Context propagation for cancellation

## Usage

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/dmitrymomot/gokit/async"
)

func main() {
	ctx := context.Background()
	
	// Define an async function that will be executed
	future := async.Async(ctx, 5, func(ctx context.Context, n int) (string, error) {
		// Simulate work
		time.Sleep(time.Second)
		return fmt.Sprintf("Result: %d", n*2), nil
	})
	
	// Do other work while the async function is running...
	
	// Wait for the result
	result, err := future.Await()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	fmt.Println(result) // Output: Result: 10
}
```

### Using Timeouts

```go
result, err := future.AwaitWithTimeout(500 * time.Millisecond)
if err != nil {
	// Handle timeout or other errors
	fmt.Println("Operation timed out or failed:", err)
	return
}
```

### Checking Completion Status

```go
if future.IsComplete() {
	fmt.Println("The operation has completed!")
	result, err := future.Await() // This won't block since it's complete
	// Process result...
} else {
	fmt.Println("Still working...")
}
```

### Working with Multiple Futures

```go
// Create multiple futures
future1 := async.Async(ctx, 1, processNumber)
future2 := async.Async(ctx, 2, processNumber)
future3 := async.Async(ctx, 3, processNumber)

// Wait for all futures to complete
results, err := async.WaitAll(future1, future2, future3)
if err != nil {
	fmt.Println("One of the operations failed:", err)
	return
}

// Process all results
for i, result := range results {
	fmt.Printf("Result %d: %s\n", i+1, result)
}

// Or wait for any future to complete
index, result, err := async.WaitAny(future1, future2, future3)
if err != nil {
	fmt.Println("The completed operation failed:", err)
	return
}

fmt.Printf("Future %d completed first with result: %s\n", index+1, result)
```

## API Reference

### Types

#### `Future[U any]`

Represents the result of an asynchronous computation.

### Functions

#### `Async[T any, U any](ctx context.Context, param T, fn func(context.Context, T) (U, error)) *Future[U]`

Executes a function asynchronously and returns a Future.

#### `WaitAll[U any](futures ...*Future[U]) ([]U, error)`

Waits for all futures to complete and returns a slice of their results.

#### `WaitAny[U any](futures ...*Future[U]) (int, U, error)`

Waits for any of the futures to complete and returns the index of the completed future, its result, and any error.

### Methods

#### `(f *Future[U]) Await() (U, error)`

Waits for the asynchronous function to complete and returns its result and error.

#### `(f *Future[U]) AwaitWithTimeout(timeout time.Duration) (U, error)`

Waits for the asynchronous function to complete with a timeout.

#### `(f *Future[U]) IsComplete() bool`

Checks if the asynchronous function is complete without blocking.
