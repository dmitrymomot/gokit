# Sanitizer Package

A flexible, tag-based sanitization package for Go structs.

## Installation

```bash
go get -u github.com/dmitrymomot/gokit/sanitizer
```

## Overview

The sanitizer package provides a flexible, tag-based way to clean and transform struct field values. It's particularly useful for sanitizing user input or normalizing data formats before further processing or storage. This package follows an instance-based design pattern, allowing for multiple independent sanitizer instances with different configurations to be used concurrently.

## Features

- **Instance-based Design**: Create isolated sanitizer instances with different configurations
- **Configurable Separators**: Customize rule, parameter, and parameter list separators
- **Field Name Tag Support**: Use custom struct tags for field identification
- **30+ Built-in Sanitizers**: Common text transformations ready to use
- **Custom Sanitizer Registration**: Add your own custom sanitizers
- **Thread-safe Operation**: Safe for concurrent use
- **Recursive Struct Handling**: Process nested structs automatically
- **Zero Value Handling**: Skip empty fields with `omitempty` option
- **Chainable Rules**: Apply multiple sanitization rules to a single field

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    
    "github.com/dmitrymomot/gokit/sanitizer"
)

type User struct {
    Username  string `sanitize:"trim;lower"`
    Email     string `sanitize:"trim;lower;email"`
    FirstName string `sanitize:"trim;capitalize"`
    LastName  string `sanitize:"trim;capitalize"`
    Bio       string `sanitize:"trim;striphtml"`
}

func main() {
    user := &User{
        Username:  "  JohnDoe123  ",
        Email:     " USER@EXAMPLE.COM ",
        FirstName: " john ",
        LastName:  " DOE ",
        Bio:       "<p>I am a <strong>developer</strong></p>",
    }
    
    // Create a new sanitizer instance
    s, err := sanitizer.New()
    if err != nil {
        fmt.Println("Error creating sanitizer:", err)
        return
    }
    
    // Sanitize the struct
    if err := s.SanitizeStruct(user); err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Output: {Username:johndoe123 Email:user@example.com FirstName:John LastName:Doe Bio:I am a developer}
}
```

### Creating Custom Sanitizer Instances

```go
package main

import (
    "fmt"
    "reflect"
    
    "github.com/dmitrymomot/gokit/sanitizer"
)

type Product struct {
    Name        string  `sanitize:"trim" json:"name"`
    Description string  `sanitize:"trim;striphtml" json:"description"`
    SKU         string  `sanitize:"trim;upper" json:"sku"`
    Price       float64 `json:"price"`
}

func main() {
    // Create a custom sanitizer with comma as rule separator
    s, err := sanitizer.New(
        sanitizer.WithRuleSeparator(","),
        sanitizer.WithFieldNameTag("json"),
    )
    if err != nil {
        fmt.Println("Error creating sanitizer:", err)
        return
    }
    
    // Register a custom sanitizer
    err = s.RegisterSanitizer("sku", func(fieldValue any, fieldType reflect.StructField, params []string) any {
        if v, ok := fieldValue.(string); ok {
            return "PRD-" + v
        }
        return fieldValue
    })
    if err != nil {
        fmt.Println("Error registering sanitizer:", err)
        return
    }
    
    product := &Product{
        Name:        "  Wireless Keyboard  ",
        Description: " <p>Great wireless keyboard</p> ",
        SKU:         "kb100",
        Price:       29.99,
    }
    
    if err := s.SanitizeStruct(product); err != nil {
        fmt.Println("Error sanitizing struct:", err)
        return
    }
    
    // Output: {Name:Wireless Keyboard Description:Great wireless keyboard SKU:PRD-KB100 Price:29.99}
}
```

### Using OmitEmpty and Nested Structs

```go
type Address struct {
    Street  string `sanitize:"trim;omitempty"`
    City    string `sanitize:"trim;capitalize"`
    State   string `sanitize:"trim;upper"`
    ZipCode string `sanitize:"trim;numeric"`
}

type Customer struct {
    Name    string  `sanitize:"trim"`
    Email   string  `sanitize:"trim;lower;email"`
    Address Address // Will be processed recursively
}

func main() {
    customer := &Customer{
        Name:  "  Jane Smith  ",
        Email: " JANE@EXAMPLE.COM ",
        Address: Address{
            Street:  "", // This will be skipped due to omitempty
            City:    " new york ",
            State:   " ny ",
            ZipCode: "10001-ABC",
        },
    }
    
    s, err := sanitizer.New()
    if err != nil {
        fmt.Println("Error creating sanitizer:", err)
        return
    }
    
    if err := s.SanitizeStruct(customer); err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Output: {Name:Jane Smith Email:jane@example.com Address:{Street: City:New York State:NY ZipCode:10001}}
}
```

### Error Handling

```go
s, err := sanitizer.New(
    sanitizer.WithRuleSeparator(""),  // Invalid: empty separator
)
if err != nil {
    switch {
    case errors.Is(err, sanitizer.ErrInvalidSanitizerConfiguration):
        // Handle configuration error
        fmt.Println("Invalid sanitizer configuration")
    default:
        // Handle other errors
        fmt.Println("Error:", err)
    }
}
```

## Built-in Sanitizers

| Name | Description | Example |
|------|-------------|--------|
| `trim` | Removes leading and trailing whitespace | `" hello " → "hello"` |
| `trimspace` | Removes all whitespace | `"hello world" → "helloworld"` |
| `lower` | Converts to lowercase | `"Hello" → "hello"` |
| `upper` | Converts to uppercase | `"hello" → "HELLO"` |
| `capitalize` | Capitalizes first letter | `"hello" → "Hello"` |
| `ucfirst` | Same as capitalize | `"hello" → "Hello"` |
| `lcfirst` | Lowercases first letter | `"Hello" → "hello"` |
| `camelcase` | Converts to camelCase | `"hello world" → "helloWorld"` |
| `pascalcase` | Converts to PascalCase | `"hello world" → "HelloWorld"` |
| `snakecase` | Converts to snake_case | `"helloWorld" → "hello_world"` |
| `kebabcase` | Converts to kebab-case | `"helloWorld" → "hello-world"` |
| `replace:old,new` | Replaces text | `replace:world,earth: "hello world" → "hello earth"` |
| `striphtml` | Removes HTML tags | `"<p>hello</p>" → "hello"` |
| `escape` | HTML escapes characters | `"<hello>" → "&lt;hello&gt;"` |
| `alphanum` | Removes non-alphanumeric characters | `"hello123!@#" → "hello123"` |
| `numeric` | Removes non-numeric characters | `"abc123" → "123"` |
| `truncate:N` | Limits string length to N | `truncate:5: "hello world" → "hello"` |
| `normalize` | Normalizes line endings to \n | `"hello\r\nworld" → "hello\nworld"` |
| `slug` | Creates URL-friendly slug | `"Café & World!" → "cafe-world"` |
| `email` | Trims and lowercases email | `" User@EXAMPLE.com " → "user@example.com"` |
| `uuid` | Normalizes UUID format | `"{ABCDEF-1234}" → "abcdef-1234"` |
| `bool` | Converts strings to boolean | `"yes" → "true", "no" → "false"` |

## Best Practices

1. **Use Instance-Based Design**:
   - Create separate sanitizer instances for different parts of your application to avoid configuration conflicts
   - Initialize sanitizers once and reuse them for better performance

2. **Error Handling**:
   - Always check error returns from sanitizer operations, especially when registering custom sanitizers
   - Wrap sanitization errors in context-specific errors for better debugging

3. **Custom Rules**:
   - Register custom sanitizers for domain-specific cleaning operations rather than chaining multiple built-in sanitizers
   - Keep custom sanitizer functions simple and focused on a single task

4. **Configuration**:
   - Set appropriate separators that don't conflict with your data content
   - Consider using different field tags for more flexibility in complex applications

5. **Field Tag Selection**:
   - Use `WithFieldNameTag` to leverage existing JSON/XML field tags for consistency
   - Create clear naming conventions for sanitizer tags to improve code readability

6. **Omit Empty Values**:
   - Use `omitempty` to skip sanitization of zero values when appropriate
   - Prefer this over conditional sanitization in your business logic

7. **Performance**:
   - For high-frequency sanitization, consider pre-creating and caching sanitizer instances
   - Group related sanitization operations together to minimize processing overhead

## API Reference

### Types

```go
// SanitizeFunc defines the signature of a sanitization function
type SanitizeFunc func(fieldValue any, fieldType reflect.StructField, params []string) any

// Option configures a Sanitizer instance
type Option func(*Sanitizer) error

// Sanitizer holds instance-specific configuration and sanitizer map
type Sanitizer struct {...}
```

### Functions

```go
// New creates a new Sanitizer with the given options
func New(options ...Option) (*Sanitizer, error)

// MustNew creates a new Sanitizer or panics on error
func MustNew(options ...Option) *Sanitizer
```

### Methods

```go
// RegisterSanitizer registers a custom sanitization function
func (s *Sanitizer) RegisterSanitizer(tag string, fn SanitizeFunc) error

// SanitizeStruct sanitizes the struct fields based on 'sanitize' tags
func (s *Sanitizer) SanitizeStruct(ptr any) error
```

### Configuration Options

```go
func WithRuleSeparator(separator string) Option
func WithParamSeparator(separator string) Option
func WithParamListSeparator(separator string) Option
func WithFieldNameTag(tag string) Option
func WithSanitizers(sanitizers map[string]SanitizeFunc) Option
```

### Errors

```go
var ErrInvalidSanitizerConfiguration = errors.New("invalid sanitizer configuration")
var ErrUnknownSanitizer = errors.New("unknown sanitizer")
```