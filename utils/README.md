# Utils Package

A collection of common utility functions for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/utils
```

## Overview

The `utils` package provides a set of lightweight utility functions for common tasks in Go applications, including pointer handling, string manipulation, display name normalization, and reflection helpers.

## Features

- Generic pointer creation for any type
- URL-friendly slug generation
- Email display name normalization
- Function and struct name reflection
- Pretty printing of complex data structures
- Zero external dependencies

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
customPtr := utils.Ptr(MyStruct{Field: "value"})
```

### String Manipulation

```go
// Generate URL-friendly slugs
slug := utils.GenerateSlug("Hello World & Café!")
fmt.Println(slug) // "hello-world-cafe"

// Works with special characters, spaces, and diacritics
titleSlug := utils.GenerateSlug("Über die Brücke")
fmt.Println(titleSlug) // "uber-die-brucke"

// Use it for file names, URL paths, etc.
articleSlug := utils.GenerateSlug("10 Best Ways to Learn Go!")
fmt.Println(articleSlug) // "10-best-ways-to-learn-go"
```

### Display Name Utilities

```go
// Extract and normalize display names from email addresses
name := utils.NormalizeDisplayName("john.doe123@example.com")
fmt.Println(name) // "John Doe"

name = utils.NormalizeDisplayName("jane-smith@company.org")
fmt.Println(name) // "Jane Smith"

name = utils.NormalizeDisplayName("dev.team_lead@tech-corp.io")
fmt.Println(name) // "Dev Team Lead"
```

### Reflection Utilities

```go
// Get the fully qualified name of a function
funcName := utils.QualifiedFuncName(myFunction)
fmt.Println(funcName) // "github.com/myorg/mypackage.myFunction"

// Works with method references too
methodName := utils.QualifiedFuncName(myStruct.Method)
fmt.Println(methodName) // "github.com/myorg/mypackage.MyStruct.Method"

// Get struct names for logging and reflection
type User struct {
    Name string
    Age  int
}
structName := utils.GetStructName(User{})
fmt.Println(structName) // "User"

// Works with pointers too
structName = utils.GetStructName(&User{})
fmt.Println(structName) // "User"
```

### Pretty Printing

```go
// Pretty print complex data structures
data := map[string]interface{}{
    "users": []User{
        {Name: "Alice", Age: 30},
        {Name: "Bob", Age: 25},
    },
    "settings": map[string]bool{
        "enabled": true,
        "debug":   false,
    },
}

// Print with formatting and colors (to console)
utils.PrettyPrint(data)

// Get pretty-printed string
prettyJSON := utils.PrettyString(data)
fmt.Println(prettyJSON)
```

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

// Deprecated alias for GenerateSlug
func ToSlug(s string) string
```

### Email/Name Functions

```go
// Extract and normalize a display name from an email address
func NormalizeDisplayName(email string) string

// Deprecated alias for NormalizeDisplayName
func GetNormalizedDisplayName(email string) string
```

### Reflection Functions

```go
// Get the fully qualified name of a function
func QualifiedFuncName(v any) string

// Deprecated alias for QualifiedFuncName
func FullyQualifiedFuncName(v any) string

// Get the name of a struct type
func GetStructName(v any) string
```

### Print Functions

```go
// Pretty print a value to stdout
func PrettyPrint(v any)

// Get a pretty-printed string representation of a value
func PrettyString(v any) string
```

## Best Practices

1. **Using Pointer Helpers**:
   - Use `Ptr()` when initializing structs with pointer fields
   - Particularly useful for optional configuration parameters

2. **Slug Generation**:
   - Use slugs for SEO-friendly URLs and file names
   - Always validate the output length for database constraints

3. **Display Names**:
   - Apply `NormalizeDisplayName()` for user-facing displays
   - Consider additional validation for user-provided display names

4. **Reflection Functions**:
   - Use reflection utilities primarily for logging and debugging
   - Avoid relying on reflection in performance-critical code paths
