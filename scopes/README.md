# Scopes Package

A flexible authorization scopes management package for granular permission control in Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/scopes
```

## Overview

The `scopes` package provides a comprehensive solution for working with authorization scopes in authentication systems. It offers parsing, validation, and comparison capabilities with support for hierarchical scopes and wildcard patterns for flexible permission management.

## Features

- Parses and joins scope strings with customizable separators
- Supports wildcard patterns for flexible permissions (`*` and namespace wildcards like `admin.*`)
- Validates scopes against allowed scope collections
- Handles hierarchical scopes with customizable delimiters (e.g., `admin.read`)
- Provides optimized implementations for both small and large scope collections
- Normalizes scopes by removing duplicates and ensuring consistent ordering

## Usage

### Basic Scope Parsing and Validation

```go
package main

import (
	"fmt"

	"github.com/dmitrymomot/gokit/scopes"
)

func main() {
	// Parse scopes from a string
	userScopes := scopes.ParseScopes("read write admin.users")
	fmt.Println("User scopes:", userScopes)
	// Output: User scopes: [read write admin.users]

	// Join scopes back to a string
	scopesStr := scopes.JoinScopes(userScopes)
	fmt.Println("Scopes string:", scopesStr)
	// Output: Scopes string: read write admin.users

	// Validate against allowed scopes
	allowedScopes := []string{"read", "write", "admin.*"}
	isValid := scopes.ValidateScopes(userScopes, allowedScopes)
	fmt.Println("Scopes valid:", isValid) 
	// Output: Scopes valid: true
}
```

### Checking for Specific Permissions

```go
package main

import (
	"fmt"

	"github.com/dmitrymomot/gokit/scopes"
)

func main() {
	// User has these scopes
	userScopes := []string{"read", "admin.users"}

	// Check if user has a specific scope
	hasRead := scopes.HasScope(userScopes, "read")
	fmt.Println("Has read permission:", hasRead) 
	// Output: Has read permission: true

	// Check using wildcard matching
	hasAdminUsers := scopes.HasScope(userScopes, "admin.users")
	fmt.Println("Has admin.users permission:", hasAdminUsers)
	// Output: Has admin.users permission: true

	// Check multiple required scopes
	hasAllRequired := scopes.HasAllScopes(userScopes, []string{"read", "admin.users"})
	fmt.Println("Has all required scopes:", hasAllRequired)
	// Output: Has all required scopes: true

	// Check if user has any of the required scopes
	hasAnyRequired := scopes.HasAnyScopes(userScopes, []string{"write", "admin.users"})
	fmt.Println("Has any required scopes:", hasAnyRequired)
	// Output: Has any required scopes: true
}
```

### Working with Wildcards

```go
package main

import (
	"fmt"

	"github.com/dmitrymomot/gokit/scopes"
)

func main() {
	// Admin with wildcard permissions
	adminScopes := []string{"admin.*"}

	// Check various admin permissions
	hasAdminRead := scopes.HasScope(adminScopes, "admin.read")
	hasAdminWrite := scopes.HasScope(adminScopes, "admin.write")
	hasAdminUsers := scopes.HasScope(adminScopes, "admin.users")

	fmt.Println("Has admin.read:", hasAdminRead)
	// Output: Has admin.read: true
	fmt.Println("Has admin.write:", hasAdminWrite)
	// Output: Has admin.write: true
	fmt.Println("Has admin.users:", hasAdminUsers)
	// Output: Has admin.users: true

	// But non-admin scopes won't match
	hasUserRead := scopes.HasScope(adminScopes, "user.read")
	fmt.Println("Has user.read:", hasUserRead)
	// Output: Has user.read: false

	// Global wildcard matches everything
	superAdmin := []string{"*"}
	fmt.Println("Super admin has admin.anything:", 
		scopes.HasScope(superAdmin, "admin.anything"))
	// Output: Super admin has admin.anything: true
}
```

### Normalizing Scopes

```go
package main

import (
	"fmt"

	"github.com/dmitrymomot/gokit/scopes"
)

func main() {
	// Normalize scopes (removes duplicates and sorts alphabetically)
	messyScopes := []string{"write", "read", "read", "admin.*", "write"}
	normalized := scopes.NormalizeScopes(messyScopes)

	fmt.Println("Normalized scopes:", normalized)
	// Output: Normalized scopes: [admin.* read write]

	// Compare two scope collections (order-independent)
	scopes1 := []string{"read", "write", "admin.users"}
	scopes2 := []string{"write", "admin.users", "read"}

	equal := scopes.EqualScopes(scopes1, scopes2)
	fmt.Println("Scopes are equal:", equal)
	// Output: Scopes are equal: true
}
```

### Customizing Scope Delimiters

```go
package main

import (
	"fmt"

	"github.com/dmitrymomot/gokit/scopes"
)

func main() {
	// Set custom separators for a different format
	scopes.ScopeSeparator = ","
	scopes.ScopeWildcard = "?"
	scopes.ScopeDelimiter = ":"

	// Now parse using comma separator
	userScopes := scopes.ParseScopes("read,write,admin:users")
	fmt.Println("Parsed with custom separator:", userScopes)
	// Output: Parsed with custom separator: [read write admin:users]

	// Check wildcard with custom format
	adminScopes := []string{"admin:?"}
	hasAdminUsers := scopes.HasScope(adminScopes, "admin:users")
	fmt.Println("Has admin:users with custom wildcard:", hasAdminUsers)
	// Output: Has admin:users with custom wildcard: true
}
```

## Best Practices

1. **Normalize scopes** before storing to ensure consistent format
2. **Use hierarchical scopes** (e.g., `resource.action`) for more granular permissions
3. **Leverage wildcards** for granting broader permissions to trusted users
4. **Validate all user scopes** against allowed scopes before processing requests

## API Reference

### Variables

#### `ScopeSeparator string`
Defines the character used to separate multiple scopes in a string (default: space " ").

#### `ScopeWildcard string`
Defines the wildcard character used for pattern matching (default: "*").

#### `ScopeDelimiter string`
Defines the character used to separate hierarchical scope parts (default: ".").

### Constants

#### `ScopeRead`, `ScopeWrite`, `ScopeDelete`, `ScopeAdmin`
Common scope constants for typical operations.

### Functions

#### `ParseScopes(scopesStr string) []string`
Converts a space-separated string of scopes into a string slice.

#### `JoinScopes(scopes []string) string`
Converts a slice of scopes back to a space-separated string.

#### `ScopeMatches(scope, pattern string) bool`
Checks if a single scope matches a pattern, supporting wildcards.

#### `HasScope(scopes []string, scope string) bool`
Checks if a collection of scopes contains a specific scope.

#### `HasAllScopes(scopes, required []string) bool`
Checks if scopes contain all of the required scopes.

#### `HasAnyScopes(scopes, required []string) bool`
Checks if scopes contain any of the required scopes.

#### `EqualScopes(scopes1, scopes2 []string) bool`
Checks if two scope collections are identical (regardless of order).

#### `ValidateScopes(scopes, validScopes []string) bool`
Checks if all scopes are valid according to the provided valid scopes.

#### `NormalizeScopes(scopes []string) []string`
Removes duplicate scopes and sorts them alphabetically.

### Errors

#### `ErrInvalidScope`
Returned when a scope is not valid.

#### `ErrScopeNotAllowed`
Returned when a scope is not in the list of allowed scopes.