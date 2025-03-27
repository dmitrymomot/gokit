# Queue Package Example

This example demonstrates how to use the `gokit/queue` package with a Redis storage backend, using `miniredis` for testing purposes.

## Overview

The example shows several key features of the queue package:

1. Setting up a queue with Redis storage
2. Adding middleware to the queue (logging and timing)
3. Registering type-safe handlers for different task types
4. Enqueueing immediate and delayed jobs
5. Handling job failures and automatic retries
6. Graceful queue shutdown

## Components

- **EmailPayload** and **NotificationPayload**: Example payload types for different task types
- **LoggingMiddleware**: Records job processing details and execution status
- **TimingMiddleware**: Measures execution time for each job
- **Error handling**: Demonstrates the retry mechanism with a handler that always fails

## Usage

To run the example:

```sh
go run main.go
```

The example will:
1. Start a miniredis server (in-memory Redis for testing)
2. Create a queue with Redis storage
3. Register handlers for different task types
4. Process immediate, delayed, and error-producing jobs
5. Output detailed logs of the job processing

## Expected Output

The output will show:
- Jobs being enqueued with their IDs
- Processing logs with timing information
- Retries for the failed job
- Graceful shutdown of the queue

## Dependencies

- `github.com/alicebob/miniredis/v2`: In-memory Redis server for testing
- `github.com/redis/go-redis/v9`: Redis client
- `github.com/dmitrymomot/gokit/queue`: The gokit queue package
- `github.com/dmitrymomot/gokit/queue/redis`: Redis storage implementation
