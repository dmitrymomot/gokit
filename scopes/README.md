# Scopes Package

This package provides functionality for handling authentication and authorization scopes in API systems.

## Overview

The `scopes` package offers tools to:

1. Parse and join scope strings
2. Match scopes against patterns
3. Validate scopes against allowed collections
4. Check scope containment and equality

This package is designed to work with hierarchical scopes (e.g., `admin.read`, `user.write`) and wildcard matching.

## Installation

```go
import "github.com/dmitrymomot/gokit/scopes"
```

## Usage

### Parsing and Joining Scopes

```go
// Parse a space-separated string of scopes into a slice
scopeList := scopes.ParseScopes("read write admin.users.*")
// Returns: []string{"read", "write", "admin.users.*"}

// Join a slice of scopes back into a space-separated string
scopeStr := scopes.JoinScopes([]string{"read", "write", "admin.*"})
// Returns: "read write admin.*"
```

### Custom Delimiters and Separators

The package uses public variables that can be modified to customize scope handling:

```go
// Use custom separator (e.g., commas)
scopes.ScopeSeparator = "," // Default is space (" ")
commaSeparatedScopes := scopes.ParseScopes("read,write,admin.users")
// Returns: []string{"read", "write", "admin.users"}

// Use custom delimiter for hierarchical scopes
scopes.ScopeDelimiter = ":" // Default is dot (".")
// With this change, "admin:read" would be a hierarchical scope

// Use custom wildcard
scopes.ScopeWildcard = "?" // Default is asterisk ("*")
// With this change, "admin:?" would be a wildcard scope

// Remember to restore original values if needed
scopes.ScopeSeparator = " "
scopes.ScopeDelimiter = "."
scopes.ScopeWildcard = "*"

// Or save original values at the beginning
originalSeparator := scopes.ScopeSeparator
defer func() {
    scopes.ScopeSeparator = originalSeparator
}()
```

### Checking Scope Matching

```go
// Check if a collection of scopes contains a specific scope
// Supports wildcards and hierarchical scopes
hasScope := scopes.ContainsScope([]string{"admin.*", "read"}, "admin.users")
// Returns: true (because "admin.*" matches "admin.users")

// Check if scopes contain all required scopes
hasAll := scopes.HasAllScopes(
    []string{"admin.*", "read", "write"},
    []string{"admin.users", "read"}
)
// Returns: true

// Check if scopes contain any of the required scopes
hasAny := scopes.HasAnyScopes(
    []string{"read", "write"}, 
    []string{"delete", "read"}
)
// Returns: true

// Check if two scope collections are equal (regardless of order)
equal := scopes.EqualScopes(
    []string{"read", "write"}, 
    []string{"write", "read"}
)
// Returns: true
```

### Normalizing Scopes

```go
// Remove duplicates and sort scope collections for consistency
normalized := scopes.NormalizeScopes([]string{"write", "read", "read", "admin.*"})
// Returns: []string{"admin.*", "read", "write"}

// Use custom delimiters for normalization
normalized := scopes.NormalizeScopesWithDelimiter(
    []string{"write", "read", "admin:*"},
    ":" // Custom delimiter
)
// Returns: []string{"admin:*", "read", "write"}
```

### Validating Scopes

```go
// Check if all scopes are valid according to a list of allowed scopes
valid := scopes.ValidateScopes(
    []string{"admin.read", "user.write"},
    []string{"admin.*", "user.*", "system.*"}
)
// Returns: true
```

## Configuration

The package provides public variables that can be customized:

```go
// Default values
scopes.ScopeSeparator // " " (space) - separates scopes in a string
scopes.ScopeWildcard  // "*" - represents a wildcard that matches anything
scopes.ScopeDelimiter // "." - separates parts in hierarchical scopes

// You can modify these values to use custom delimiters and separators
```

## Common Scope Constants

The package defines common scope constants for convenience:

```go
scopes.ScopeRead    // "read"
scopes.ScopeWrite   // "write"
scopes.ScopeDelete  // "delete"
scopes.ScopeAdmin   // "admin"
```

## Error Handling

The package provides the following error types:

- `ErrInvalidScope` - Returned when a scope format is invalid
- `ErrScopeNotAllowed` - Returned when a scope is not in the list of allowed scopes

These are defined in the `errors.go` file and should be used with the standard Go error handling patterns like `errors.Is()` or `errors.As()`.

## Best Practices

1. Always normalize scopes before storage or comparison using `NormalizeScopes()`
2. Use wildcard patterns to simplify scope management (e.g., `admin.*` instead of individual scopes)
3. Design hierarchical scopes that map to your resource structure
4. Validate user-provided scopes against a whitelist using `ValidateScopes()`
5. For security-critical code, check both `HasAllScopes()` and `ValidateScopes()`
