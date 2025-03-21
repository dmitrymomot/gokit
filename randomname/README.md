# Random Name Generator

A Go package for generating random, human-readable workspace names in the format "adjective-noun" or "adjective-noun-xxxxxx".

## Features

- Thread-safe implementation
- Session-based uniqueness tracking
- Support for external validation (e.g., database checks)
- Two generation formats:
    - With suffix: `brave-tiger-1a2b3c`
    - Without suffix: `brave-tiger`
- Large namespace:
    - 42 adjectives × 44 nouns = 1,848 base combinations
    - With 24-bit suffix: 1,848 × 16,777,215 ≈ 31 billion possible combinations

## Installation

```bash
go get github.com/dmitrymomot/gokit/randomname
```

## Usage

### Basic Usage

```go
import "github.com/dmitrymomot/gokit/randomname"

// Generate a name with suffix (e.g., "brave-tiger-1a2b3c")
name := randomname.Generate(nil)

// Generate a simple name without suffix (e.g., "brave-tiger")
simpleName := randomname.GenerateSimple(nil)
```

### With Database Validation

```go
// Generate a name that doesn't exist in the database
name := randomname.Generate(func(name string) bool {
    exists, _ := db.WorkspaceExists(name)
    return !exists
})
```

### Managing Used Names

```go
// Clear the internal cache of used names
randomname.Reset()
```

## Thread Safety

The package is thread-safe and optimized for concurrent use:

- Uses mutex for synchronization
- Efficiently handles slow validation callbacks
- Prevents race conditions with immediate name reservation
- Cleans up rejected names automatically

## Implementation Details

### Name Generation Process

1. Generate and reserve a candidate name under lock
2. Release the lock and execute any validation callback
3. If the callback rejects the candidate:
    - Remove it from reservations
    - Try again with a new candidate
4. Return the successful candidate

### Word Lists

The package includes carefully selected word lists:

- 42 descriptive adjectives (e.g., brave, swift, gentle)
- 44 animal nouns (e.g., tiger, eagle, panda)

These words are chosen to create memorable, friendly, and professional workspace names.

## Best Practices

1. **Database Validation**

    ```go
    // Always handle database errors appropriately
    name := randomname.Generate(func(name string) bool {
        exists, err := db.WorkspaceExists(name)
        if err != nil {
            // Handle error appropriately
            return false
        }
        return !exists
    })
    ```

2. **Memory Management**

    ```go
    // Periodically reset the used names cache if not needed
    // for the entire session
    randomname.Reset()
    ```

3. **Suffix Usage**
    - Use `Generate()` when uniqueness is critical
    - Use `GenerateSimple()` when readability is more important than uniqueness

## License

This package is part of the github.com/dmitrymomot/gokit project. See the LICENSE file in the root directory for details.
