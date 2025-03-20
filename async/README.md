# Async

The `async` package provides a simple way to run functions concurrently and wait for all of them to finish. It is designed to handle asynchronous computations with ease, allowing you to execute functions in parallel and retrieve their results once they are completed.

## Usage

### Basic Usage

The `Async` function allows you to run a function asynchronously and returns a `Future` object. You can use the `Await` method on the `Future` to wait for the function to complete and retrieve its result.

Here's a basic example:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"saas/pkg/async"
)

func main() {
	ctx := context.Background()

	// Function that takes an int parameter and returns a string
	future := async.Async[int, string](ctx, 42, func(ctx context.Context, num int) (string, error) {
		time.Sleep(100 * time.Millisecond) // Simulate work
		return fmt.Sprintf("Number: %d", num), nil
	})

	// Wait for the result
	result, err := future.Await()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result)
	}
}
```

### Handling Context Cancellation

The `Async` function respects the provided `context.Context`. If the context is canceled or times out, the asynchronous function will stop execution and return an error.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"saas/pkg/async"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	future := async.Async[int, string](ctx, 42, func(ctx context.Context, num int) (string, error) {
		select {
		case <-time.After(100 * time.Millisecond): // Simulate work
			return fmt.Sprintf("Number: %d", num), nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})

	result, err := future.Await()
	if err != nil {
		fmt.Println("Error:", err) // Expected: context deadline exceeded
	} else {
		fmt.Println("Result:", result)
	}
}
```

### Error Propagation

Errors from the asynchronous function are propagated correctly and can be checked after calling `Await`.

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"saas/pkg/async"
)

func main() {
	ctx := context.Background()

	expectedErr := errors.New("an error occurred in the async function")

	future := async.Async[int, int](ctx, 42, func(ctx context.Context, num int) (int, error) {
		time.Sleep(50 * time.Millisecond) // Simulate work
		return 0, expectedErr
	})

	result, err := future.Await()
	if err != nil {
		fmt.Println("Error:", err) // Expected: an error occurred in the async function
	} else {
		fmt.Println("Result:", result)
	}
}
```

### Using Custom Structures

You can use custom structures as parameters and return types for the asynchronous function.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"saas/pkg/async"
)

type Input struct {
	X int
	Y int
}

type Output struct {
	Sum int
}

func main() {
	ctx := context.Background()

	future := async.Async[Input, Output](ctx, Input{X: 10, Y: 15}, func(ctx context.Context, in Input) (Output, error) {
		time.Sleep(50 * time.Millisecond) // Simulate work
		return Output{Sum: in.X + in.Y}, nil
	})

	result, err := future.Await()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Result:", result.Sum) // Expected: 25
	}
}
```

### Concurrency Example

The `async` package allows you to run multiple asynchronous functions concurrently and wait for all of them to complete.

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"saas/pkg/async"
)

func main() {
	ctx := context.Background()
	startTime := time.Now()

	var mu sync.Mutex
	order := []string{}

	future1 := async.Async[int, int](ctx, 1, func(ctx context.Context, num int) (int, error) {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		order = append(order, "first")
		mu.Unlock()
		return num, nil
	})

	future2 := async.Async[int, int](ctx, 2, func(ctx context.Context, num int) (int, error) {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		order = append(order, "second")
		mu.Unlock()
		return num, nil
	})

	future3 := async.Async[int, int](ctx, 3, func(ctx context.Context, num int) (int, error) {
		time.Sleep(70 * time.Millisecond)
		mu.Lock()
		order = append(order, "third")
		mu.Unlock()
		return num, nil
	})

	// Await the results
	_, _ = future1.Await()
	_, _ = future2.Await()
	_, _ = future3.Await()

	duration := time.Since(startTime)
	fmt.Println("Duration:", duration)

	// Check the order of completion
	fmt.Println("Order:", order)
}
```

## Benchmarks

The package includes benchmark tests to measure the performance of the `Async` helper under different conditions. You can run the benchmarks using the following command:

```sh
go test -bench=. saas/pkg/async
```
