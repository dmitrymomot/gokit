# Validator Package

A flexible, tag-based validation system for Go structs with customizable error messages.

## Installation

```bash
go get github.com/dmitrymomot/gokit/validator
```

## Overview

The `validator` package provides a robust validation system for Go structs using struct tags. It includes a wide range of built-in validators and supports custom validation rules, with flexibility for error message customization.

## Features

- Simple struct tag-based validation
- 25+ built-in validators covering common use cases
- Custom validation rule support with simple registration
- Customizable error messages with translation support
- Support for nested structs and complex validation rules
- Thread-safe operation for concurrent validation
- Comprehensive field type support

## Usage

### Basic Validation

```go
import "github.com/dmitrymomot/gokit/validator"

type User struct {
    Username  string `validate:"required,username" label:"Username"`
    Email     string `validate:"required,email" label:"Email Address"`
    Age       int    `validate:"required,range:18,100" label:"Age"`
    Password  string `validate:"required,password" label:"Password"`
    Phone     string `validate:"phone" label:"Phone Number"`
    Website   string `validate:"url" label:"Website"`
}

func main() {
    // Create a validator with default error messages
    v := validator.NewValidator(nil)
    
    user := User{
        Username: "john_doe",
        Email: "invalid-email",
        Age: 15,
        Password: "weak",
        Phone: "123", // Invalid phone
        Website: "example", // Invalid URL
    }
    
    // Validate the struct
    err := v.ValidateStruct(user)
    if err != nil {
        // Handle validation errors
        // err will contain all validation failures
        fmt.Println(err)
    }
}
```

### Custom Error Translator

```go
// Create a custom error translator function
translator := func(key string, label string, params ...any) string {
    switch key {
    case "validation.required":
        return fmt.Sprintf("%s field cannot be empty", label)
    case "validation.email":
        return fmt.Sprintf("%s must be a valid email (e.g., user@example.com)", label)
    case "validation.range":
        min, max := params[0], params[1]
        return fmt.Sprintf("%s must be between %v and %v", label, min, max)
    default:
        return fmt.Sprintf("Invalid %s", label)
    }
}

// Create a validator with custom error messages
v := validator.NewValidator(translator)
```

### Custom Validation Rules

```go
import (
    "reflect"
    "regexp"
    "github.com/dmitrymomot/gokit/validator"
)

// Register a custom validation function
validator.RegisterValidation("zipcode", func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
    // Convert field value to string
    val, ok := fieldValue.(string)
    if !ok {
        return validator.NewValidationError("validation.type_mismatch", label)
    }
    
    // Validate US zip code format
    zipcodeRegex := regexp.MustCompile(`^\d{5}(-\d{4})?$`)
    if !zipcodeRegex.MatchString(val) {
        return validator.NewValidationError("validation.zipcode", label)
    }
    
    return nil
})

// Use in struct definition
type Address struct {
    Street  string `validate:"required" label:"Street Address"`
    City    string `validate:"required" label:"City"`
    State   string `validate:"required,length:2" label:"State"`
    ZipCode string `validate:"required,zipcode" label:"ZIP Code"`
}
```

### Validating Nested Structs

```go
type Address struct {
    Street  string `validate:"required" label:"Street"`
    City    string `validate:"required" label:"City"`
    Country string `validate:"required" label:"Country"`
}

type Customer struct {
    Name      string   `validate:"required,fullname" label:"Full Name"`
    Email     string   `validate:"required,email" label:"Email"`
    Address   Address  `validate:"required"` // Validates nested struct
    ShipAddrs []Address `validate:"required"` // Validates slice of structs
}

// Validate
customer := Customer{
    Name: "John Doe",
    Email: "john.doe@example.com",
    Address: Address{
        Street: "123 Main St",
        // Missing City and Country - will fail validation
    },
}

err := v.ValidateStruct(customer)
// err will contain nested validation errors
```

## Validation Options

### Common Validators

| Validator   | Description                           | Example Usage                |
|-------------|---------------------------------------|------------------------------|
| required    | Field must not be empty               | `validate:"required"`        |
| email       | Must be a valid email                 | `validate:"email"`           |
| range       | Value must be within range            | `validate:"range:1,100"`     |
| min         | Minimum value or length               | `validate:"min:5"`           |
| max         | Maximum value or length               | `validate:"max:100"`         |
| regex       | Must match regular expression         | `validate:"regex:^[a-z]+$"`  |
| in          | Must be one of specified values       | `validate:"in:a,b,c"`        |
| length      | Must be exact length                  | `validate:"length:10"`       |

### Special Formats

| Validator   | Description                           | Example Usage                |
|-------------|---------------------------------------|------------------------------|
| url         | Valid URL                             | `validate:"url"`             |
| phone       | Valid phone number                    | `validate:"phone"`           |
| uuid        | Valid UUID                            | `validate:"uuid"`            |
| date        | Valid date in specified format        | `validate:"date:2006-01-02"` |
| creditcard  | Valid credit card number              | `validate:"creditcard"`      |
| hexcolor    | Valid hex color code                  | `validate:"hexcolor"`        |
| ip          | Valid IP address                      | `validate:"ip"`              |

### Content Rules

| Validator   | Description                           | Example Usage                |
|-------------|---------------------------------------|------------------------------|
| alpha       | Letters only                          | `validate:"alpha"`           |
| alphanum    | Letters and numbers only              | `validate:"alphanum"`        |
| numeric     | Numbers only                          | `validate:"numeric"`         |
| username    | Valid username format                 | `validate:"username"`        |
| password    | Meets password complexity rules       | `validate:"password"`        |
| fullname    | Valid full name format                | `validate:"fullname"`        |
| slug        | Valid URL slug                        | `validate:"slug"`            |

### Comparison Rules

| Validator   | Description                           | Example Usage                |
|-------------|---------------------------------------|------------------------------|
| eq          | Equal to value                        | `validate:"eq:100"`          |
| ne          | Not equal to value                    | `validate:"ne:0"`            |
| gt          | Greater than value                    | `validate:"gt:0"`            |
| gte         | Greater than or equal to value        | `validate:"gte:1"`           |
| lt          | Less than value                       | `validate:"lt:100"`          |
| lte         | Less than or equal to value           | `validate:"lte:99"`          |

## Error Handling

The validator returns detailed error information that can be used to display user-friendly messages:

```go
err := v.ValidateStruct(user)
if err != nil {
    // Type assertion to access validation errors
    if validationErr, ok := err.(*validator.ValidationErrors); ok {
        // Access specific field errors
        for field, fieldErr := range validationErr.Errors {
            fmt.Printf("Field '%s': %s\n", field, fieldErr)
        }
    }
}
```

## Best Practices

1. **Descriptive Labels**: 
   - Use the `label` tag to provide user-friendly field names for error messages
   - Example: `label:"Email Address"` instead of just `label:"email"`

2. **Validation Groups**:
   - Group related validators together in a logical order
   - Start with `required` if the field is mandatory
   - Example: `validate:"required,email,max:100"`

3. **Custom Validations**:
   - Register custom validators for business-specific rules
   - Keep custom validation functions small and focused
   - Return specific error types for better error handling

4. **Error Handling**:
   - Implement a custom error translator for user-friendly messages
   - Consider internationalization (i18n) for multi-language support
   - Return all validation errors at once rather than stopping at the first error
