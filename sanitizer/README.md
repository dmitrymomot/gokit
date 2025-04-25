# Sanitizer Package

Tag-based field sanitization for Go structs with 15+ built-in sanitizers.

## Installation

```bash
go get github.com/dmitrymomot/gokit/sanitizer
```

## Overview

The `sanitizer` package provides a clean, zero-configuration approach to sanitize struct fields using struct tags. It supports 15+ built-in sanitizers and allows custom sanitization functions to be registered.

## Features

- Declarative tag-based field sanitization
- Thread-safe implementation with mutex protection
- Multiple sanitization rules per field
- Built-in sanitizers for common operations
- Extensible with custom sanitization functions

## Usage

### Basic Usage

```go
import (
    "fmt"
    "github.com/dmitrymomot/gokit/sanitizer"
)

type User struct {
    Username string `sanitize:"trim,lower"`
    Email    string `sanitize:"trim,lower"`
    Bio      string `sanitize:"striphtml,trim"`
    Website  string `sanitize:"trim,replace:http:https"`
}

func main() {
    user := &User{
        Username: "  JohnDoe  ",
        Email:    "  JOHN@EXAMPLE.COM  ",
        Bio:      "<p>Hello World</p>  ",
        Website:  "http://example.com",
    }

    // Apply all sanitization rules
    sanitizer.SanitizeStruct(user)

    fmt.Println(user.Username) // "johndoe"
    fmt.Println(user.Email)    // "john@example.com"
    fmt.Println(user.Bio)      // "Hello World"
    fmt.Println(user.Website)  // "https://example.com"
}
```

### Custom Sanitizers

Register your own custom sanitizers to extend functionality:

```go
import (
    "reflect"
    "strings"
    "github.com/dmitrymomot/gokit/sanitizer"
)

// Define a custom sanitizer to replace profanity
func profanityFilterSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
    if v, ok := fieldValue.(string); ok {
        profanity := []string{"bad", "words", "profanity"}
        result := v
        for _, word := range profanity {
            result = strings.ReplaceAll(result, word, "***")
        }
        return result
    }
    return fieldValue
}

func init() {
    // Register the custom sanitizer
    sanitizer.RegisterSanitizer("profanity", profanityFilterSanitizer)
}

// Use it in your structs
type Comment struct {
    Text string `sanitize:"trim,profanity"`
}
```

### HTTP Input Sanitization

Automatically sanitize user input from web forms:

```go
func handleUserRegistration(w http.ResponseWriter, r *http.Request) {
    var user User
    
    // Parse form data into struct
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Sanitize all fields based on struct tags
    sanitizer.SanitizeStruct(&user)
    
    // Continue with sanitized data
    // ...
}
```

## Built-in Sanitizers

| Sanitizer   | Description                             | Example Tag                    |
|-------------|-----------------------------------------|--------------------------------|
| trim        | Removes whitespace                      | `sanitize:"trim"`              |
| lower       | Converts to lowercase                   | `sanitize:"lower"`             |
| upper       | Converts to uppercase                   | `sanitize:"upper"`             |
| replace     | Replaces strings                        | `sanitize:"replace:old:new"`   |
| striphtml   | Removes HTML tags                       | `sanitize:"striphtml"`         |
| escape      | Escapes HTML characters                 | `sanitize:"escape"`            |
| alphanum    | Removes non-alphanumeric                | `sanitize:"alphanum"`          |
| numeric     | Keeps only digits                       | `sanitize:"numeric"`           |
| truncate    | Limits string length                    | `sanitize:"truncate:10"`       |
| normalize   | Normalizes line endings                 | `sanitize:"normalize"`         |
| capitalize  | Capitalizes first character             | `sanitize:"capitalize"`        |
| camelcase   | Converts to camelCase                   | `sanitize:"camelcase"`         |
| snakecase   | Converts to snake_case                  | `sanitize:"snakecase"`         |
| kebabcase   | Converts to kebab-case                  | `sanitize:"kebabcase"`         |
| ucfirst     | Uppercase first character               | `sanitize:"ucfirst"`           |

## API Reference

### Core Functions

```go
// Sanitize a struct using struct tags
func SanitizeStruct(s any)

// Register a custom sanitizer
func RegisterSanitizer(tag string, fn SanitizeFunc)

// Reset all sanitizers to defaults (useful for testing)
func ResetSanitizers()
```

### Sanitizer Function Type

```go
// Function signature for sanitizers
type SanitizeFunc func(fieldValue any, fieldType reflect.StructField, params []string) any
```

## Best Practices

1. **Use multiple sanitizers**: Apply multiple sanitizers to the same field using comma-separated values (e.g., `sanitize:"trim,lower"`)
2. **Order matters**: Sanitizers are applied in the order they appear in the tag
3. **Validate after sanitizing**: Use packages like `validator` after sanitization for complete input processing
4. **Custom sanitizers**: Register custom sanitizers in an `init()` function to ensure they're available globally
5. **Security**: Always sanitize user-generated content with `striphtml` or `escape` to prevent XSS attacks
6. **Performance**: Keep sanitizer functions efficient, especially for high-traffic applications
