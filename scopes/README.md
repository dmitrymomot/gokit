# Scopes Package

A lightweight utility for handling authentication and authorization scopes in Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/scopes
```

## Overview

The `scopes` package provides tools for working with authentication and authorization scopes in API systems. It supports hierarchical scope patterns, wildcard matching, and various operations for scope validation and manipulation.

## Features

- Parse and join space-separated scope strings
- Support for hierarchical scopes (e.g., `admin.read`, `user.write`)
- Wildcard matching for scope patterns (`admin.*`)
- Scope validation against allowed collections
- Scope containment and equality checking
- Customizable delimiters and separators
- Common scope constant definitions
- Thread-safe implementation

## Usage

### Parsing and Joining Scopes

```go
import "github.com/dmitrymomot/gokit/scopes"

// Parse a space-separated string of scopes into a slice
scopeList := scopes.ParseScopes("read write admin.users.*")
// Returns: []string{"read", "write", "admin.users.*"}

// Join a slice of scopes back into a space-separated string
scopeStr := scopes.JoinScopes([]string{"read", "write", "admin.*"})
// Returns: "read write admin.*"
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

### Normalizing and Validating Scopes

```go
// Remove duplicates and sort scope collections for consistency
normalized := scopes.NormalizeScopes([]string{"write", "read", "read", "admin.*"})
// Returns: []string{"admin.*", "read", "write"}

// Check if all scopes are valid according to a list of allowed scopes
valid := scopes.ValidateScopes(
    []string{"admin.read", "user.write"},
    []string{"admin.*", "user.*", "system.*"}
)
// Returns: true
```

### Using Custom Separators and Delimiters

```go
// Save original values
origSeparator := scopes.ScopeSeparator
origDelimiter := scopes.ScopeDelimiter
origWildcard := scopes.ScopeWildcard

// Use custom separator (e.g., commas)
scopes.ScopeSeparator = ","
commaSeparatedScopes := scopes.ParseScopes("read,write,admin.users")
// Returns: []string{"read", "write", "admin.users"}

// Use custom delimiter for hierarchical scopes
scopes.ScopeDelimiter = ":"
colonScopes := scopes.ParseScopes("admin:read user:write")
// Returns: []string{"admin:read", "user:write"}

// Use custom wildcard
scopes.ScopeWildcard = "?"
wildcardScopes := scopes.ContainsScope([]string{"admin:?"}, "admin:read")
// Returns: true

// Restore original values when done
defer func() {
    scopes.ScopeSeparator = origSeparator
    scopes.ScopeDelimiter = origDelimiter
    scopes.ScopeWildcard = origWildcard
}()
```

## API Reference

### Constants

```go
// Common scope constants for convenience
const (
    ScopeRead   = "read"
    ScopeWrite  = "write"
    ScopeDelete = "delete"
    ScopeAdmin  = "admin"
)
```

### Configuration Variables

```go
// Customizable configuration (package variables)
var ScopeSeparator = " "  // Separates scopes in a string
var ScopeDelimiter = "."  // Separates parts in hierarchical scopes
var ScopeWildcard  = "*"  // Represents a wildcard that matches anything
```

### Functions

```go
// Parsing and joining
func ParseScopes(scopes string) []string
func JoinScopes(scopes []string) string

// Scope matching
func ContainsScope(scopeList []string, scope string) bool
func HasAllScopes(scopeList, requiredScopes []string) bool
func HasAnyScopes(scopeList, requiredScopes []string) bool
func EqualScopes(scopes1, scopes2 []string) bool

// Normalization and validation
func NormalizeScopes(scopes []string) []string
func NormalizeScopesWithDelimiter(scopes []string, delimiter string) []string
func ValidateScopes(scopes, allowedScopes []string) bool
```

### Error Types

```go
// Error types for handling scope-related errors
var ErrInvalidScope = errors.New("invalid scope")
var ErrScopeNotAllowed = errors.New("scope not allowed")
```

## Error Handling

```go
// Validate user-provided scopes
userScopes := scopes.ParseScopes(userInput)
if valid, err := scopes.ValidateScopes(userScopes, allowedScopes); err != nil {
    switch {
    case errors.Is(err, scopes.ErrInvalidScope):
        // Handle invalid scope format
        return fmt.Errorf("invalid scope format: %w", err)
    case errors.Is(err, scopes.ErrScopeNotAllowed):
        // Handle disallowed scope
        return fmt.Errorf("scope not allowed: %w", err)
    default:
        // Handle other errors
        return err
    }
}
```

## Best Practices

1. **Scope Design**:
   - Design hierarchical scopes that map to your resource structure
   - Use consistent naming patterns (e.g., `resource.action`)
   - Consider using standard OAuth scope conventions when applicable

2. **Validation and Security**:
   - Always validate user-provided scopes against a whitelist
   - For security-critical operations, check both `HasAllScopes()` and `ValidateScopes()`
   - Use specific scopes rather than overly broad ones for better security

3. **Data Consistency**:
   - Normalize scopes before storage or comparison using `NormalizeScopes()`
   - Use wildcard patterns to simplify scope management (e.g., `admin.*` instead of individual scopes)
   - When storing scopes, store the normalized form to ensure consistent comparison

4. **Performance**:
   - Cache validation results for frequently used scope combinations
   - For large scope collections, consider optimizing scope matching algorithms
   - Use appropriate data structures for scope storage and lookup
