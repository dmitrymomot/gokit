# Sanitizer Package

Tag-based field sanitization for Go structs with 15+ built-in sanitizers.

## Installation

```bash
go get github.com/dmitrymomot/gokit/sanitizer
```

## Overview

The `sanitizer` package provides a clean, tag-based approach to sanitize struct fields using struct tags. It offers a thread-safe implementation with 15+ built-in sanitizers and supports custom sanitization functions. This package is designed to simplify input sanitation in web applications and APIs.

## Features

- Declarative tag-based field sanitization
- Thread-safe implementation with mutex protection
- Multiple sanitization rules per field
- 15+ built-in sanitizers for common operations
- Extensible with custom sanitization functions
- Zero dependencies beyond standard library and strcase
- Support for parameterized sanitizers

## Usage

### Basic Example

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

## Best Practices

1. **Order Matters**:
   - Sanitizers are applied in the order they appear in the tag
   - Place `trim` before other sanitizers to handle whitespace first

2. **Security**:
   - Always sanitize user-generated content with `striphtml` or `escape` to prevent XSS attacks
   - Combine sanitization with validation for robust input processing

3. **Custom Sanitizers**:
   - Register custom sanitizers in an `init()` function to ensure they're available globally
   - Keep custom sanitizer functions simple and focused on a single task

4. **Performance**:
   - Keep sanitizer functions efficient, especially for high-traffic applications
   - Consider applying sanitization only on the fields that need it

5. **Usage Patterns**:
   - Use multiple sanitizers with comma-separated values (e.g., `sanitize:"trim,lower"`)
   - Pair with `validator` package for complete input processing

## API Reference

### Functions

```go
func SanitizeStruct(s any)
```
Sanitizes the struct fields based on 'sanitize' tags. Takes a pointer to a struct as input.

```go
func RegisterSanitizer(tag string, fn SanitizeFunc)
```
Registers a custom sanitization function with the given tag name.

```go
func ResetSanitizers()
```
Resets all sanitizers to the default set. Useful for testing purposes.

### Types

```go
type SanitizeFunc func(fieldValue any, fieldType reflect.StructField, params []string) any
```
Function signature for sanitizers. Takes the field value, field type, and optional parameters.

### Built-in Sanitizers

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
