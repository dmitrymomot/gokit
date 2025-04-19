# Sanitizer Package

The sanitizer package provides a flexible and extensible way to sanitize struct fields using tags. It offers various built-in sanitizers and allows registration of custom sanitization functions.

## Features

- Simple tag-based field sanitization
- Thread-safe operation
- Multiple sanitization rules per field
- Extensible with custom sanitizers
- Various built-in sanitizers

## Built-in Sanitizers

| Sanitizer  | Description                             | Example Usage                |
| ---------- | --------------------------------------- | ---------------------------- |
| trim       | Removes leading and trailing whitespace | `sanitize:"trim"`            |
| lower      | Converts string to lowercase            | `sanitize:"lower"`           |
| upper      | Converts string to uppercase            | `sanitize:"upper"`           |
| replace    | Replaces occurrences of strings         | `sanitize:"replace:old:new"` |
| striphtml  | Removes HTML tags                       | `sanitize:"striphtml"`       |
| escape     | Escapes special HTML characters         | `sanitize:"escape"`          |
| alphanum   | Removes non-alphanumeric characters     | `sanitize:"alphanum"`        |
| numeric    | Keeps only numeric characters           | `sanitize:"numeric"`         |
| truncate   | Limits string length                    | `sanitize:"truncate:10"`     |
| normalize  | Normalizes line endings                 | `sanitize:"normalize"`       |
| capitalize | Capitalizes first character             | `sanitize:"capitalize"`      |
| camelcase  | Converts to camelCase                   | `sanitize:"camelcase"`       |
| snakecase  | Converts to snake_case                  | `sanitize:"snakecase"`       |
| kebabcase  | Converts to kebab-case                  | `sanitize:"kebabcase"`       |
| ucfirst    | Uppercase first character               | `sanitize:"ucfirst"`         |

## Installation

```bash
go get github.com/dmitrymomot/gokit/pkg/sanitizer
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/yourusername/gokit/pkg/sanitizer"
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
        Email: "  JOHN@EXAMPLE.COM  ",
        Bio: "<p>Hello World</p>  ",
        Website: "http://example.com",
    }

    sanitizer.SanitizeStruct(user)

    fmt.Printf("Username: %q\n", user.Username) // "johndoe"
    fmt.Printf("Email: %q\n", user.Email)       // "john@example.com"
    fmt.Printf("Bio: %q\n", user.Bio)          // "Hello World"
    fmt.Printf("Website: %q\n", user.Website)   // "https://example.com"
}
```

### Custom Sanitizers

You can register your own custom sanitizers:

```go
package main

import (
    "reflect"
    "strings"
    "github.com/yourusername/gokit/pkg/sanitizer"
)

// Custom sanitizer that removes specific words
func removeBadWordsSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
    if v, ok := fieldValue.(string); ok {
        badWords := []string{"bad", "words"}
        result := v
        for _, word := range badWords {
            result = strings.ReplaceAll(result, word, "***")
        }
        return result
    }
    return fieldValue
}

func main() {
    // Register the custom sanitizer
    sanitizer.RegisterSanitizer("removebadwords", removeBadWordsSanitizer)

    type Comment struct {
        Text string `sanitize:"trim,removebadwords"`
    }

    comment := &Comment{
        Text: "This has bad words in it",
    }

    sanitizer.SanitizeStruct(comment)
    // Output: "This has *** words in it"
}
```

## Thread Safety

The sanitizer package is thread-safe and can be used concurrently.
