# Scopes Package

A flexible scope management system for API authorization and OAuth-style permission handling.

## Installation

```bash
go get github.com/dmitrymomot/gokit/scopes
```

## Overview

The `scopes` package provides a comprehensive toolkit for working with OAuth-style scopes in authorization systems. It offers efficient parsing, validation, and comparison of scope strings with support for hierarchical scopes and wildcards. This package is thread-safe and optimized for both small and large collections of scopes.

## Features

- Hierarchical scope support with customizable delimiters (`admin.users.read`)
- Wildcard matching for flexible permission patterns (`admin.*`)
- Efficient scope parsing and string conversion
- Validation against defined permission sets
- Scope normalization (deduplication and sorting)
- Thread-safe implementation with optimized algorithms
- Customizable separators, delimiters, and wildcards

## Usage

### Basic Example

```go
import "github.com/dmitrymomot/gokit/scopes"

// Parse a scope string into individual scopes
scopes := scopes.ParseScopes("read write admin.users")
// Returns: []string{"read", "write", "admin.users"}

// Convert scopes back to a string
scopesStr := scopes.JoinScopes(scopes)
// Returns: "read write admin.users"

// Check if user has a specific scope (with wildcard support)
hasAccess := scopes.HasScope(scopes, "admin.users.read")
// Returns: false (because "admin.users" doesn't match "admin.users.read")

// However, with wildcards:
scopes = []string{"admin.*", "read"}
hasAccess = scopes.HasScope(scopes, "admin.users.read")
// Returns: true (because "admin.*" matches "admin.users.read")
```

### Scope Validation

```go
// Define valid scopes for your API
validScopes := []string{"read", "write", "admin.*", "user.*"}

// Validate user scopes against valid scopes
userScopes := []string{"read", "admin.users", "admin.settings"}
valid := scopes.ValidateScopes(userScopes, validScopes)
// Returns: true (all user scopes match valid scopes patterns)

// Example with invalid scope
userScopes = []string{"read", "delete"} // "delete" is not in validScopes
valid = scopes.ValidateScopes(userScopes, validScopes)
// Returns: false
```

### Checking Required Scopes

```go
// Check if user has all required scopes
userScopes := []string{"admin.*", "read"}
requiredScopes := []string{"admin.users", "read"}

hasAll := scopes.HasAllScopes(userScopes, requiredScopes)
// Returns: true (user has all required scopes)

// Check if user has any of the required scopes
requiredForPartialAccess := []string{"admin.settings", "write"}
hasAny := scopes.HasAnyScopes(userScopes, requiredForPartialAccess)
// Returns: true (user has at least "admin.settings" via "admin.*")
```

### Error Handling

```go
// Handling scope validation in your application
userScopes := []string{"read", "delete"}
validScopes := []string{"read", "write", "admin.*"}

if !scopes.ValidateScopes(userScopes, validScopes) {
    // Handle invalid scopes scenario
    // You can create or wrap errors based on the result
    return fmt.Errorf("invalid scope provided: %w", scopes.ErrScopeNotAllowed)
}

// When implementing your own scope validation with custom errors
func validateUserScopes(userScopes, validScopes []string) error {
    for _, scope := range userScopes {
        if !scopes.HasScope(validScopes, scope) {
            if !isValidScopeFormat(scope) {
                return fmt.Errorf("%s: %w", scope, scopes.ErrInvalidScope)
            }
            return fmt.Errorf("%s: %w", scope, scopes.ErrScopeNotAllowed)
        }
    }
    return nil
}

// Using the custom validator
if err := validateUserScopes(userScopes, validScopes); err != nil {
    switch {
    case errors.Is(err, scopes.ErrInvalidScope):
        // Handle invalid scope format
        return fmt.Errorf("invalid scope format: %w", err)
    case errors.Is(err, scopes.ErrScopeNotAllowed):
        // Handle unauthorized scope
        return fmt.Errorf("scope not allowed: %w", err)
    default:
        // Handle unexpected errors
        return fmt.Errorf("scope validation failed: %w", err)
    }
}
```

### Customizing Separators and Delimiters

```go
// Customize the separator (default is space " ")
scopes.ScopeSeparator = ","

// Parse comma-separated scopes
scopes := scopes.ParseScopes("read,write,admin.users")
// Returns: []string{"read", "write", "admin.users"}

// Customize the delimiter (default is ".")
scopes.ScopeDelimiter = ":"

// Customize the wildcard character (default is "*")
scopes.ScopeWildcard = "?"

// Now wildcards work with the new format
hasScope := scopes.HasScope([]string{"admin:?"}, "admin:users")
// Returns: true
```

## Best Practices

1. **Scope Design**:
   - Design hierarchical scopes for intuitive permission grouping
   - Limit scope depth to 2-3 levels for readability
   - Use consistent naming conventions (e.g., resource.action)

2. **Wildcard Usage**:
   - Use wildcards sparingly and with caution
   - Prefer specific scopes over wildcards for security-critical operations
   - Consider implementing additional checks for sensitive operations

3. **Performance**:
   - Normalize scopes after parsing to improve later comparisons
   - Cache scope validation results when possible
   - For high-traffic applications, pre-compute common scope validations

4. **Security**:
   - Validate all user-provided scopes against your valid scopes
   - Never trust client-provided scope strings without validation
   - Implement scope-based authorization at all levels of your application

## API Reference

### Configuration Variables

```go
var ScopeSeparator = " " // Separator used between scopes in a string
var ScopeWildcard = "*"   // Wildcard character for scope matching
var ScopeDelimiter = "." // Delimiter for hierarchical scope parts
```

### Functions

```go
func ParseScopes(scopesStr string) []string
```
Converts a space-separated string of scopes into a string slice.

```go
func JoinScopes(scopes []string) string
```
Converts a slice of scopes back to a space-separated string.

```go
func ScopeMatches(scope, pattern string) bool
```
Checks if a single scope matches a pattern (with wildcard support).

```go
func HasScope(scopes []string, scope string) bool
```
Checks if a collection of scopes contains a specific scope.

```go
func HasAllScopes(scopes, required []string) bool
```
Checks if scopes contain all of the required scopes.

```go
func HasAnyScopes(scopes, required []string) bool
```
Checks if scopes contain any of the required scopes.

```go
func EqualScopes(scopes1, scopes2 []string) bool
```
Checks if two scope collections are identical (regardless of order).

```go
func ValidateScopes(scopes, validScopes []string) bool
```
Checks if all scopes are valid according to the provided valid scopes.

```go
func NormalizeScopes(scopes []string) []string
```
Removes duplicate scopes and sorts them alphabetically.

### Error Types

```go
var ErrInvalidScope = errors.New("invalid scope")
var ErrScopeNotAllowed = errors.New("scope not allowed")