# Utils Package

A collection of common utility functions for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/utils
```

## Overview

The `utils` package provides a set of lightweight utility functions for common tasks in Go applications, including pointer creation, string manipulation, display name normalization, and reflection utilities. All functions are thread-safe and designed to be simple and efficient with zero external dependencies beyond the Go standard library.

## Features

- Generic pointer creation for any type with zero allocation overhead
- URL-friendly slug generation with proper Unicode handling
- Email address display name normalization and formatting
- Function and struct name reflection utilities for debugging and logging
- Pretty JSON formatting for complex data structures
- Thread-safe implementation suitable for concurrent use
- Zero external dependencies beyond Go standard library

## Usage

### Pointer Utilities

```go
import "github.com/dmitrymomot/gokit/utils"

// Create pointers to values without separate variables
type Config struct {
    Name     string
    Enabled  *bool
    Count    *int
    Timeout  *time.Duration
}

// Create a struct with pointer fields directly
config := Config{
    Name:     "Example",
    Enabled:  utils.Ptr(true),
    Count:    utils.Ptr(42),
    Timeout:  utils.Ptr(5 * time.Second),
}

// Works with any type
stringPtr := utils.Ptr("hello")
floatPtr := utils.Ptr(3.14)
timePtr := utils.Ptr(time.Now())
```

### String Manipulation

```go
// Generate URL-friendly slugs
slug := utils.GenerateSlug("Hello World & Café!")
// Returns: "hello-world-cafe"

// Works with special characters, spaces, and diacritics
titleSlug := utils.GenerateSlug("Über die Brücke")
// Returns: "uber-die-brucke"

// Use it for file names, URL paths, etc.
articleSlug := utils.GenerateSlug("10 Best Ways to Learn Go!")
// Returns: "10-best-ways-to-learn-go"
```

### Display Name Utilities

```go
// Extract and normalize display names from email addresses
name := utils.NormalizeDisplayName("john.doe123@example.com")
// Returns: "John Doe"

name = utils.NormalizeDisplayName("jane-smith@company.org")
// Returns: "Jane Smith"

name = utils.NormalizeDisplayName("dev.team_lead@tech-corp.io")
// Returns: "Dev Team Lead"
```

### Reflection Utilities

```go
// Get the fully qualified name of a function
func myFunction() {}
funcName := utils.QualifiedFuncName(myFunction)
// Returns: "github.com/myorg/mypackage.myFunction"

// Get struct name for debugging and logging
type User struct {
    Name string
    Age  int
}
structName := utils.StructName(User{})
// Returns: "User"

// Works with pointers too
structName = utils.StructName(&User{})
// Returns: "User"

// Get fully qualified struct name with package
qualifiedName := utils.QualifiedStructName(User{})
// Returns: "github.com/myorg/mypackage.User"

// Use GetNameFromStruct with types that implement Name() method
type Person struct{}
func (p Person) Name() string { return "CustomPerson" }

name := utils.GetNameFromStruct(Person{}, utils.StructName)
// Returns: "CustomPerson" (from Name() method)
```

### JSON Formatting

```go
// Format complex data structures as pretty JSON
data := map[string]interface{}{
    "users": []map[string]interface{}{
        {"name": "Alice", "age": 30},
        {"name": "Bob", "age": 25},
    },
    "settings": map[string]bool{
        "enabled": true,
        "debug":   false,
    },
}

// Get formatted JSON string
jsonStr := utils.FormatJSON(data)
// Returns pretty-printed JSON with proper indentation
```

## Best Practices

1. **Pointer Usage**:
   - Use `Ptr()` when initializing structs with pointer fields for cleaner code
   - Particularly useful for optional configuration parameters
   - Remember that pointers introduce indirection; use them appropriately

2. **Slug Generation**:
   - Use slugs for SEO-friendly URLs, file names, and database identifiers
   - Always validate the output length for database column constraints
   - Consider adding a uniqueness check if slugs must be unique in your application

3. **Display Names**:
   - Apply `NormalizeDisplayName()` for user-facing displays, not for authentication
   - Consider additional validation for user-provided display names
   - Be aware that the function removes numbers at the beginning and end of names

4. **Reflection Functions**:
   - Use reflection utilities primarily for logging, debugging, and diagnostics
   - Avoid using reflection in performance-critical code paths
   - Cache reflection results when they're used repeatedly

5. **JSON Formatting**:
   - Use `FormatJSON()` for debugging and logging purposes, not for production APIs
   - Be cautious with large data structures as JSON formatting can be memory-intensive
   - Handle potential marshaling errors when working with custom types

## API Reference

### Pointer Functions

```go
// Create a pointer to any value
func Ptr[T any](v T) *T
```

### String Functions

```go
// Generate a URL-friendly slug from a string
func GenerateSlug(s string) string

// Deprecated: Use GenerateSlug instead
func ToSlug(s string) string
```

### Email/Name Functions

```go
// Extract and normalize a display name from an email address
func NormalizeDisplayName(email string) string

// Deprecated: Use NormalizeDisplayName instead
func GetNormalizedDisplayName(email string) string
```

### Reflection Functions

```go
// Get the fully qualified name of a function
func QualifiedFuncName(v any) string

// Deprecated: Use QualifiedFuncName instead
func FullyQualifiedFuncName(v any) string

// Get the name of a struct type without package
func StructName(v any) string

// Get the fully qualified struct name with package
func QualifiedStructName(v any) string

// Deprecated: Use QualifiedStructName instead
func FullyQualifiedStructName(v any) string

// Interface for types that can provide their own name
type NamedEntity interface {
    Name() string
}

// Get the name from a struct that implements NamedEntity or use fallback
func GetNameFromStruct(v any, fallback func(v any) string) string
```

### Print Functions

```go
// Format any value as a pretty-printed JSON string
func FormatJSON(v ...any) string

// Deprecated: Use FormatJSON instead
func PrettyPrint(v ...any) string
```
