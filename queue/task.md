# Task: Implement a Queue System for Gokit

## Overview
Design and implement a queue server package for the Gokit library that supports different storage adapters and can run worker pools across multiple instances.

## Requirements

### API Design

Implement a simple, developer-friendly API following this pattern:
```go
q := queue.New(queue.Config{})
q.AddHandler(sendEmailTaskName, sendEmailFunc)
q.Run(ctx) // Start processing jobs, returns error or nil
```

Support type-safe handlers with this signature:
```go
func handlerFunc(ctx context.Context, params SendEmailPayload) error {
  // Job processing logic
}
```

### Core Components

#### Queue Interface:
- AddHandler - Register type-safe job handlers
- Enqueue - Add jobs to the queue
- EnqueueIn - Add delayed jobs to the queue
- Run - Start processing jobs, returns error or nil
- Stop - Gracefully shut down, returns error or nil

#### Storage Interface:
- Support pluggable storage backends
- Initial implementation for in-memory storage
- Plan for Redis and MongoDB adapters

#### Job Processing:
- Generic payload serialization/deserialization
- Concurrent job processing with worker pools
- Automatic job retries with exponential backoff
- Distributed coordination between instances

### Implementation Requirements
- Follow Gokit package standards
- Include comprehensive documentation and examples
- Provide proper error handling with defined error types
- Ensure thread safety for concurrent operations

### Deliverables
- Core queue package implementation with interfaces
- In-memory storage adapter
- README.md with documentation and examples
- Error types and handling

## Notes
- Implement step by step, do not implement everything at once. 
- Ask questions if something is not clear.
- Ask me to review the code when you are ready step by step.
- First step is to create an implementation plan and add it to the task.md file.
- Implementation plan must be as a task list, that you must follow and mark as done when you are ready.

## Implementation Plan

### Phase 1: Core Structure and Interfaces
- [x] Define error types in errors.go
- [x] Design and implement the Job struct
- [x] Design the Storage interface
- [x] Design the Queue interface
- [x] Create config.go with configuration options

### Phase 2: In-Memory Implementation
- [x] Implement in-memory storage adapter
- [x] Implement basic Queue functionality with the in-memory adapter
- [x] Add job serialization/deserialization
- [x] Implement worker pool for concurrent job processing

### Phase 3: Advanced Features
- [x] Add retry mechanism with exponential backoff
- [x] Implement delayed jobs (EnqueueIn)
- [x] Add graceful shutdown
- [x] Implement distributed locks for coordination between instances (via Redis adapter)

### Phase 4: Documentation and Testing
- [x] Write comprehensive README.md with examples
- [x] Add godoc comments to all exported functions
- [x] Write unit tests for all components
- [x] Create benchmark tests for performance analysis

### Phase 5: Extensions
- [x] Create extension points for middleware
- [x] Add events/hooks system for queue lifecycle (via middleware)
- [x] Plan Redis adapter implementation
- [ ] Plan MongoDB adapter implementation
