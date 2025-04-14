# Binder Package

The binder package provides a simple and efficient way to bind HTTP request data to Go structs. It automatically handles different content types including JSON, form data, and query parameters.

## Features

- Content-type based binding (JSON, form data)
- Query parameter binding for GET requests
- Form data binding for POST requests
- Tag-based field mapping (`json` or `form` tags)
- Support for slice types in query parameters and form values
- Support for nested structs with dot notation
- Support for time.Time fields with multiple formats
- Support for maps with string keys
- Support for custom types with TextUnmarshaler interface
- Comprehensive error handling with proper error wrapping

## Installation

```go
import "github.com/dmitrymomot/gokit/binder"
```

## Usage

### Basic Usage

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/binder"
)

// User represents a simple user structure
type User struct {
	Name  string   `json:"name" form:"name"`
	Email string   `json:"email" form:"email"`
	Age   int      `json:"age" form:"age"`
	Roles []string `json:"roles" form:"roles"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Initialize an empty user struct
	var user User
	
	// Bind request data to the user struct
	if err := binder.Bind(r, &user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Process the user data...
	log.Printf("Created user: %+v", user)
	
	// Return success response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "User created",
	})
}

func main() {
	http.HandleFunc("/users", createUserHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Binding from Different Sources

The binder will automatically select the appropriate binding method based on the request's content type:

```go
// For JSON requests (Content-Type: application/json)
// POST /users with body: {"name":"John","email":"john@example.com","age":30,"roles":["admin","user"]}
// 
// For form submissions (Content-Type: application/x-www-form-urlencoded)
// POST /users with body: name=John&email=john@example.com&age=30&roles=admin&roles=user
//
// For query parameters (GET requests)
// GET /users?name=John&email=john@example.com&age=30&roles=admin&roles=user
```

### Specialized Binders

You can also use specialized binders for specific content types:

```go
// JSON binding example:
var user User
if err := binder.BindJSON(r, &user); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

// Query parameter binding example:
var searchParams struct {
    Query  string `form:"q"`
    Limit  int    `form:"limit"`
    Page   int    `form:"page"`
    SortBy string `form:"sort_by"`
}
if err := binder.BindQuery(r, &searchParams); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}

// Form data binding example:
var contactForm struct {
    Name    string `form:"name"`
    Email   string `form:"email"`
    Message string `form:"message"`
}
if err := binder.BindForm(r, &contactForm); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

### Nested Struct Binding

The binder supports nested structs with dot notation for form data and query parameters:

```go
// Define your nested structs
type Address struct {
    Street  string `json:"street" form:"street"`
    City    string `json:"city" form:"city"`
    ZipCode string `json:"zip_code" form:"zip_code"`
}

type User struct {
    Name    string  `json:"name" form:"name"`
    Email   string  `json:"email" form:"email"`
    Address Address `json:"address" form:"address"` // Nested struct
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := binder.Bind(r, &user); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    log.Printf("User: %s, Email: %s", user.Name, user.Email)
    log.Printf("Address: %s, %s, %s", user.Address.Street, user.Address.City, user.Address.ZipCode)
    
    // Process the user...
}
```

#### How to Submit Nested Data

For JSON requests, use regular nested JSON:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "address": {
    "street": "123 Main St",
    "city": "San Francisco",
    "zip_code": "94105"
  }
}
```

For form data or query parameters, use dot notation:

```
name=John+Doe&email=john@example.com&address.street=123+Main+St&address.city=San+Francisco&address.zip_code=94105
```

### Time.Time Binding

The binder automatically handles `time.Time` fields with multiple common formats:

```go
type Event struct {
    Title     string    `json:"title" form:"title"`
    StartDate time.Time `json:"start_date" form:"start_date"`
    EndDate   time.Time `json:"end_date" form:"end_date"`
}

func createEventHandler(w http.ResponseWriter, r *http.Request) {
    var event Event
    if err := binder.Bind(r, &event); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    log.Printf("Event: %s from %v to %v", 
        event.Title, 
        event.StartDate.Format(time.RFC3339),
        event.EndDate.Format(time.RFC3339))
    
    // Process the event...
}
```

For form data or query parameters, the package accepts various time formats:

```
title=Conference&start_date=2025-05-15T09:00:00Z&end_date=2025-05-17T18:00:00Z
```

Or simpler formats:

```
title=Conference&start_date=2025-05-15&end_date=2025-05-17
```

### Error Handling

The package provides specific error types to help identify binding issues:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    var data MyStruct
    err := binder.Bind(r, &data)
    
    if err != nil {
        switch {
        case errors.Is(err, binder.ErrInvalidContentType):
            http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
            return
            
        case errors.Is(err, binder.ErrEmptyBody):
            http.Error(w, "Request body is empty", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrInvalidJSON):
            http.Error(w, "Invalid JSON format", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrInvalidFormData):
            http.Error(w, "Invalid form data", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrUnsupportedType):
            http.Error(w, "Unsupported data type", http.StatusBadRequest)
            return
            
        case errors.Is(err, binder.ErrUnsupportedTimeFormat):
            http.Error(w, "Invalid date/time format", http.StatusBadRequest)
            return
            
        default:
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }
    
    // Process the bound data...
}
