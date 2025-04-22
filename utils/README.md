# Utils Package

## Overview

The Utils package provides a collection of common utility functions that are useful across various Go applications. It offers helper functions for working with strings, emails, pointers, function reflection, and more.

## Documentation

### Email Utilities

#### `NormalizeDisplayName(email string) string`

Extracts a username from an email address and normalizes it for use as a display name.

- **Parameters**:
  - `email`: The email address to process
- **Returns**:
  - A normalized display name extracted from the email address

#### `GetNormalizedDisplayName(email string) string` (Deprecated)

Deprecated alias for `NormalizeDisplayName`.

### Pointer Utilities

#### `Ptr[T any](v T) *T`

A generic utility function that returns a pointer to the provided value.

- **Parameters**:
  - `v`: A value of any type
- **Returns**:
  - A pointer to the input value

### String Utilities

#### `GenerateSlug(s string) string`

Converts a string to a URL-friendly slug by removing diacritics, replacing non-alphanumeric characters with hyphens, and converting to lowercase.

- **Parameters**:
  - `s`: The string to convert
- **Returns**:
  - A URL-friendly slug

#### `ToSlug(s string) string` (Deprecated)

Deprecated alias for `GenerateSlug`.

### Reflection Utilities

#### `QualifiedFuncName(v any) string`

Returns a function's fully qualified name in the format `[package].[func name]`.

- **Parameters**:
  - `v`: A function
- **Returns**:
  - The fully qualified name of the function, or an empty string if the input is not a function

#### `FullyQualifiedFuncName(v any) string` (Deprecated)

Deprecated alias for `QualifiedFuncName`.

### Struct Utilities

#### `GetStructName(v any) string`

Gets the name of a struct type.

- **Parameters**:
  - `v`: A struct instance or a pointer to a struct
- **Returns**:
  - The name of the struct type

## Usage Examples

### Email Utilities

```go
package main

import (
	"fmt"
	
	"github.com/dmitrymomot/gokit/utils"
)

func main() {
	email := "john.doe123@example.com"
	name := utils.NormalizeDisplayName(email)
	fmt.Println(name) // Output: "John Doe"
}
```

### Pointer Utilities

```go
package main

import (
	"fmt"
	
	"github.com/dmitrymomot/gokit/utils"
)

func main() {
	// Create a pointer to a string value
	strPtr := utils.Ptr("hello")
	fmt.Println(*strPtr) // Output: "hello"
	
	// Useful in struct initialization
	type Config struct {
		Name    string
		Enabled *bool
		Count   *int
	}
	
	config := Config{
		Name:    "Example",
		Enabled: utils.Ptr(true),
		Count:   utils.Ptr(42),
	}
	
	fmt.Printf("Config: %+v\n", config)
}
```

### String Utilities

```go
package main

import (
	"fmt"
	
	"github.com/dmitrymomot/gokit/utils"
)

func main() {
	input := "Hello World & Café!"
	slug := utils.GenerateSlug(input)
	fmt.Println(slug) // Output: "hello-world-cafe"
}
```

### Reflection Utilities

```go
package main

import (
	"fmt"
	
	"github.com/dmitrymomot/gokit/utils"
)

func exampleFunction() {
	// Function body
}

func main() {
	funcName := utils.QualifiedFuncName(exampleFunction)
	fmt.Println(funcName) // Output: "main.exampleFunction"
}
```
