# Binder Package

The binder package provides a simple and efficient way to bind HTTP request data to Go structs. It automatically handles different content types including JSON, form data, and query parameters.

## Features

- Content-type based binding (JSON, form data)
- Query parameter binding for GET requests
- Form data binding for POST requests
- Tag-based field mapping (`json` or `form` tags)
- Support for slice types in query parameters and form values
- Comprehensive error handling

## Installation

```go
import "github.com/dmitrymomot/gokit/binder"
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/dmitrymomot/gokit/binder"
)

type User struct {
	Name  string   `json:"name" form:"name"`
	Age   int      `json:"age" form:"age"`
	Email string   `json:"email" form:"email"`
	Tags  []string `json:"tags" form:"tags"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := binder.Bind(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	fmt.Fprintf(w, "User: %+v", user)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
```

### Specialized Binders

You can also use specialized binders for specific content types:

```go
// JSON binding
var user User
if err := binder.BindJSON(r, &user); err != nil {
	// Handle error
}

// Query parameter binding
var filters Filters
if err := binder.BindQuery(r, &filters); err != nil {
	// Handle error
}

// Form data binding
var formData FormData
if err := binder.BindForm(r, &formData); err != nil {
	// Handle error
}
```

### Error Handling

The package provides specific error types to help identify binding issues:

```go
if err := binder.Bind(r, &user); err != nil {
	switch {
	case errors.Is(err, binder.ErrInvalidContentType):
		// Handle unsupported content type
	case errors.Is(err, binder.ErrEmptyBody):
		// Handle empty request body
	case errors.Is(err, binder.ErrInvalidJSON):
		// Handle invalid JSON
	case errors.Is(err, binder.ErrInvalidFormData):
		// Handle invalid form data
	case errors.Is(err, binder.ErrUnsupportedType):
		// Handle unsupported target type
	default:
		// Handle other errors
	}
}
```

## Integration Examples

### With Standard HTTP Handlers

```go
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := binder.Bind(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Process the user data...
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}
```
