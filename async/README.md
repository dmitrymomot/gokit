# Async Package

A type-safe asynchronous execution package using Go generics.

## Installation

```bash
go get github.com/dmitrymomot/gokit/async
```

## Overview

The `async` package provides a clean, type-safe way to execute functions asynchronously in Go. It uses generics to ensure type safety throughout the asynchronous operation lifecycle. This package is thread-safe and optimized for concurrent operations in production systems.

## Features

- Type-safe asynchronous function execution with generic types
- Future pattern implementation with await functionality
- Timeout support for asynchronous operations
- Utility functions for working with multiple futures
- Context propagation for cancellation
- Thread-safe implementation for concurrent use

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
// Create an async task with a timeout
future := async.Async(ctx, 5, func(ctx context.Context, n int) (string, error) {
    // Simulate long-running work
    time.Sleep(2 * time.Second)
    return fmt.Sprintf("Result: %d", n*2), nil
})

// Wait for the result with a shorter timeout
result, err := future.AwaitWithTimeout(500 * time.Millisecond)
if err != nil {
    // This will happen because our function takes 2 seconds
    fmt.Println("Operation timed out or failed:", err)
    // err will contain context.DeadlineExceeded
    return
}
```

### Checking Completion Status

```go
future := async.Async(ctx, 5, func(ctx context.Context, n int) (string, error) {
    time.Sleep(time.Second)
    return fmt.Sprintf("Result: %d", n*2), nil
})

// Check if complete without blocking
if future.IsComplete() {
    fmt.Println("The operation has completed!")
    result, err := future.Await() // This won't block since it's complete
    // Process result...
} else {
    fmt.Println("Still working...")
}

// Wait a bit and check again
time.Sleep(1500 * time.Millisecond)
if future.IsComplete() {
    // This will now be true
    fmt.Println("The operation has completed!")
}
```

### Working with Multiple Futures

```go
// Function to process a number asynchronously
processNumber := func(ctx context.Context, n int) (string, error) {
    time.Sleep(time.Duration(n) * 200 * time.Millisecond)
    return fmt.Sprintf("Processed %d", n), nil
}

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
// Output:
// Result 1: Processed 1
// Result 2: Processed 2
// Result 3: Processed 3

// Or wait for any future to complete
index, result, err := async.WaitAny(future1, future2, future3)
if err != nil {
    fmt.Println("The completed operation failed:", err)
    return
}

fmt.Printf("Future %d completed first with result: %s\n", index+1, result)
// Output will be: "Future 1 completed first with result: Processed 1"
// Because it has the shortest sleep time
```

### Error Handling

```go
import (
    "context"
    "errors"
    "fmt"
    "time"
    
    "github.com/dmitrymomot/gokit/async"
)

// Function that may fail
riskyOperation := func(ctx context.Context, shouldFail bool) (string, error) {
    if shouldFail {
        return "", errors.New("operation failed")
    }
    return "success", nil
}

// Successful case
successFuture := async.Async(ctx, false, riskyOperation)
result, err := successFuture.Await()
if err != nil {
    // This won't execute
    fmt.Println("Error:", err)
} else {
    fmt.Println("Success:", result)
    // Output: Success: success
}

// Failure case
failureFuture := async.Async(ctx, true, riskyOperation)
result, err := failureFuture.Await()
if err != nil {
    fmt.Println("Error:", err)
    // Output: Error: operation failed
} else {
    // This won't execute
    fmt.Println("Success:", result)
}

// Handling cancellation
cancelCtx, cancel := context.WithCancel(ctx)
future := async.Async(cancelCtx, false, func(ctx context.Context, _ bool) (string, error) {
    time.Sleep(2 * time.Second)
    // Check if context was cancelled
    if ctx.Err() != nil {
        return "", ctx.Err()
    }
    return "completed", nil
})

// Cancel the operation
cancel()

// Wait for the result
result, err := future.Await()
if err != nil {
    fmt.Println("Operation was cancelled:", err)
    // Output: Operation was cancelled: context canceled
}
```

## Best Practices

1. **Context Usage**:
   - Always pass a context.Context to control cancellation
   - Use context timeouts for long-running operations
   - Check context cancellation in long-running async functions

2. **Error Handling**:
   - Always check errors returned from Await() or AwaitWithTimeout()
   - Handle timeouts gracefully
   - Consider retry logic for transient failures

3. **Performance**:
   - Use WaitAll for concurrent operations that all need to complete
   - Use WaitAny when you only need the first result
   - Avoid creating too many futures that could overwhelm system resources

4. **Resource Management**:
   - Cancel long-running operations that are no longer needed
   - Be aware of memory usage with large result values
   - Consider resource cleanup in your async functions

## API Reference

### Types

```go
type Future[U any] struct {
    // Contains unexported fields
}
```
Represents the result of an asynchronous computation.

### Functions

```go
func Async[T any, U any](ctx context.Context, param T, fn func(context.Context, T) (U, error)) *Future[U]
```
Executes a function asynchronously and returns a Future.

```go
func WaitAll[U any](futures ...*Future[U]) ([]U, error)
```
Waits for all futures to complete and returns a slice of their results.

```go
func WaitAny[U any](futures ...*Future[U]) (int, U, error)
```
Waits for any of the futures to complete and returns the index of the completed future, its result, and any error.

### Methods

```go
func (f *Future[U]) Await() (U, error)
```
Waits for the asynchronous function to complete and returns its result and error.

```go
func (f *Future[U]) AwaitWithTimeout(timeout time.Duration) (U, error)
```
Waits for the asynchronous function to complete with a timeout.

```go
func (f *Future[U]) IsComplete() bool
```
Checks if the asynchronous function is complete without blocking.

### Error Types

The async package primarily uses standard Go error types:
- `context.Canceled`: Returned when the context is canceled
- `context.DeadlineExceeded`: Returned when an operation times out
- Custom errors from the provided async function will be returned directly
