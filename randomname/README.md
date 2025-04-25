# Randomname Package

Thread-safe random name generator for workspace, project, and resource naming.

## Installation

```bash
go get github.com/dmitrymomot/gokit/randomname
```

## Overview

The `randomname` package generates human-readable, memorable names in either "adjective-noun" or "adjective-noun-suffix" format. It's designed for creating workspace, project, or resource identifiers with built-in collision prevention.

## Features

- Thread-safe implementation with mutex protection
- Session-based uniqueness tracking
- Support for custom validation callbacks (e.g., database checks)
- Two name formats:
  - "adjective-noun" (brave-tiger)
  - "adjective-noun-xxxxxx" (brave-tiger-1a2b3c)
- Rich word variety: 42 adjectives × 44 nouns (1,848 base combinations)
- With 24-bit suffix: ~31 billion unique combinations

## Usage

### Basic Generation

```go
import "github.com/dmitrymomot/gokit/randomname"

// Generate name with suffix (e.g., "brave-tiger-1a2b3c")
name := randomname.Generate(nil)

// Generate a simple name without suffix (e.g., "brave-tiger")
simpleName := randomname.GenerateSimple(nil)
```

### With Custom Validation

```go
// Generate a name that doesn't exist in the database
name := randomname.Generate(func(name string) bool {
    exists, err := db.WorkspaceExists(name)
    if err != nil {
        log.Printf("Error checking workspace name: %v", err)
        return false // Reject this candidate on error
    }
    return !exists // Accept only if name doesn't exist
})
```

### Managing Used Names

```go
// Clear the internal cache of used names when no longer needed
randomname.Reset()
```

### HTTP Handler Example

```go
http.HandleFunc("/api/generate-project-name", func(w http.ResponseWriter, r *http.Request) {
    // Generate a name with an optional database check
    name := randomname.Generate(func(name string) bool {
        // Check if name exists in your datastore
        exists, _ := projectStore.Exists(name)
        return !exists
    })
    
    // Return the generated name
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"name": name})
})
```

## API Reference

### Name Generation

- `Generate(check func(name string) bool) string`: Generate a name with suffix
- `GenerateSimple(check func(name string) bool) string`: Generate a name without suffix
- `Reset()`: Clear the internal cache of used names

### Name Validation

The `check` callback function allows for custom validation:

```go
type ValidateFunc func(name string) bool
```

The callback should:
- Return `true` to accept the name
- Return `false` to reject the name and generate a new one

## Thread Safety

The package provides thread-safe operation:

- Uses sync.Mutex for synchronization
- Prevents race conditions with immediate name reservation
- Executes validation outside the lock to avoid deadlocks
- Automatically cleans up rejected names

## Customization

The package includes carefully selected word lists:
- 42 descriptive adjectives (brave, swift, gentle, etc.)
- 44 animal nouns (tiger, eagle, panda, etc.)

These create memorable, friendly, and professional resource names.

## Best Practices

1. **Reset when appropriate**: Clear the name cache when starting a new session
2. **Handle validation errors**: Always properly handle errors in validation callbacks
3. **Choose the right format**:
   - Use `Generate()` (with suffix) when uniqueness is critical
   - Use `GenerateSimple()` when readability is more important than uniqueness
4. **Consider namespace size**: With only 1,848 base combinations, `GenerateSimple()` has higher collision probability
