# Sanitizer Package

## Installation

```bash
go get -u github.com/dmitrymomot/gokit/sanitizer
```

## Overview

The sanitizer package provides a flexible, tag-based way to clean and transform struct field values. It's particularly useful for sanitizing user input or normalizing data formats before further processing or storage.

This package follows an instance-based design pattern, allowing for multiple independent sanitizer instances with different configurations to be used concurrently.

## Features

* **Instance-based Design**: Create isolated sanitizer instances with different configurations
* **Configurable Separators**: Customize rule, parameter, and parameter list separators
* **Field Name Tag Support**: Use custom struct tags for field identification
* **30+ Built-in Sanitizers**: Common text transformations ready to use
* **Custom Sanitizer Registration**: Add your own custom sanitizers
* **Thread-safe Operation**: Safe for concurrent use
* **Recursive Struct Handling**: Process nested structs automatically
* **Zero Value Handling**: Skip empty fields with `omitempty` option
* **Chainable Rules**: Apply multiple sanitization rules to a single field
* **Type-safe Operations**: Preserves Go's type safety

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
    
    fmt.Printf("%+v\n", user)
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
    
    fmt.Printf("%+v\n", product)
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
    Address Address `sanitize:""` // Will be processed recursively
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
    
    fmt.Printf("%+v\n", customer)
    // Output: {Name:Jane Smith Email:jane@example.com Address:{Street: City:New York State:NY ZipCode:10001}}
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
| `slug` | Creates URL-friendly slug (removes diacritics, converts to lowercase, replaces non-alphanumeric chars with hyphens) | `"Café & World!" → "cafe-world"` |
| `email` | Trims and lowercases email | `" User@EXAMPLE.com " → "user@example.com"` |
| `uuid` | Normalizes UUID format | `"{ABCDEF-1234}" → "abcdef-1234"` |
| `bool` | Converts strings to boolean | `"yes" → true, "no" → false` |

## Best Practices

1. **Use Instance-Based Design**: Create separate sanitizer instances for different parts of your application to avoid configuration conflicts.

2. **Error Handling**: Always check error returns from sanitizer operations, especially when registering custom sanitizers.

3. **Custom Rules**: Register custom sanitizers for domain-specific cleaning operations rather than chaining multiple built-in sanitizers.

4. **Configuration**: Set appropriate separators that don't conflict with your data content.

5. **Field Tag Selection**: Use the most appropriate field tag for your application context.

6. **Omit Empty Values**: Use `omitempty` to skip sanitization of zero values when appropriate.

7. **Performance**: For high-frequency sanitization, consider pre-creating and caching sanitizer instances rather than creating them on demand.

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

// Global convenience functions using DefaultSanitizer
func RegisterSanitizer(tag string, fn SanitizeFunc) error
func SanitizeStruct(ptr any) error
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
    Username string `sanitize:"trim;lower"`
    Email    string `sanitize:"trim;lower"`
    Bio      string `sanitize:"striphtml;trim"`
    Website  string `sanitize:"trim;replace:http,https"`
}

func main() {
    user := &User{
        Username: "  JohnDoe  ",
        Email:    "  JOHN@EXAMPLE.COM  ",
        Bio:      "<p>Hello World</p>  ",
        Website:  "http://example.com",
    }

    // Create a new sanitizer instance
    s, err := sanitizer.New()
    if err != nil {
        fmt.Println("Error creating sanitizer:", err)
        return
    }
    
    // Apply all sanitization rules
    if err := s.SanitizeStruct(user); err != nil {
        fmt.Println("Error:", err)
        return
    }

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

func main() {
    // Create a new sanitizer instance
    s, err := sanitizer.New()
    if err != nil {
        fmt.Println("Error creating sanitizer:", err)
        return
    }
    
    // Register the custom sanitizer
    if err := s.RegisterSanitizer("profanity", profanityFilterSanitizer); err != nil {
        fmt.Println("Error registering sanitizer:", err)
        return
    }

// Use it in your structs
type Comment struct {
    Text string `sanitize:"trim;profanity"`
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
    
    // Create a new sanitizer instance
    s, err := sanitizer.New()
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    
    // Sanitize all fields based on struct tags
    if err := s.SanitizeStruct(&user); err != nil {
        http.Error(w, "Invalid input data", http.StatusBadRequest)
        return
    }
    
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
   - Register custom sanitizers with your sanitizer instance when initializing your application
   - Keep custom sanitizer functions simple and focused on a single task
   - Check for errors when registering custom sanitizers

4. **Performance**:
   - Keep sanitizer functions efficient, especially for high-traffic applications
   - Consider applying sanitization only on the fields that need it

5. **Usage Patterns**:
   - Use multiple sanitizers with semicolon-separated values (e.g., `sanitize:"trim;lower;replace:old:new"`)
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
| replace     | Replaces strings (comma-separated params) | `sanitize:"replace:old,new"`   |
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
