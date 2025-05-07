# Validator Package

A flexible, tag-based validation system for Go structs with customizable error messages.

## Installation

```bash
go get github.com/dmitrymomot/gokit/validator
```

## Overview

The `validator` package provides a robust validation system for Go structs using struct tags. It includes 30+ built-in validators and supports custom validation rules through a simple registration mechanism. The package is thread-safe for concurrent use and offers flexible error message customization through translator functions.

## Features

- Simple struct tag-based validation with clear syntax
- 30+ built-in validators for common validation scenarios
- Custom validation rule support with simple registration mechanism
- Customizable error messages with translator function support
- Recursive validation for nested structs and slices
- Thread-safe implementation for concurrent validation
- Comprehensive field type support including strings, numbers, slices and maps

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
    v, err := validator.New(validator.WithAllValidators())
    if err != nil {
        panic(err)
    }
    
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
        // err will be of type *validator.ValidationErrors
    }
}
```

### Custom Error Translator

```go
// Create a custom error translator function
translator := func(key string, label string, params ...string) string {
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
v, err := validator.New(validator.WithErrorTranslator(translator), validator.WithAllValidators())
if err != nil {
    panic(err)
}
```

### Custom Validation Rules

```go
import (
    "reflect"
    "regexp"
    "github.com/dmitrymomot/gokit/validator"
)

// Register the custom validator with this validator instance
err := v.RegisterValidation("custom", func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
    // Convert field value to string
    val, ok := fieldValue.(string)
    if !ok {
        return errors.New(translator("validation.type_mismatch", label, params...))
    }
    
    // Validate US zip code format
    zipcodeRegex := regexp.MustCompile(`^\d{5}(-\d{4})?$`)
    if !zipcodeRegex.MatchString(val) {
        return errors.New(translator("validation.zipcode", label, params...))
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

### Error Handling

```go
err := v.ValidateStruct(user)
if err != nil {
    // Type assertion to access validation errors
    if validationErrs, ok := validator.ExtractValidationErrors(err); ok {
        // Access specific field errors
        for field, fieldErr := range validationErrs.Values() {
            fmt.Printf("Field '%s': %s\n", field, fieldErr[0])
        }
    }
}

// Alternative checking method
if validator.IsValidationError(err) {
    // Handle validation errors
    validationErrs := validator.ExtractValidationErrors(err)
    // Process errors...
}
```

## Best Practices

1. **Field Labeling**: 
   - Use the `label` tag to provide user-friendly field names for error messages
   - Example: `label:"Email Address"` instead of just `label:"email"`
   - When no label is provided, the field name will be used

2. **Validation Rules Organization**:
   - Group related validators together in a logical order
   - Start with `required` if the field is mandatory
   - Example: `validate:"required,email,max:100"`
   - Separate rules with commas, no spaces

3. **Custom Validations**:
   - Register custom validators for business-specific rules
   - Keep custom validation functions small and focused
   - Return specific error types for better error handling
   - Use the error translator for consistent error messages

4. **Error Handling**:
   - Implement a custom error translator for user-friendly messages
   - Return all validation errors at once rather than stopping at the first error
   - Use the helper functions like `ExtractValidationErrors` and `IsValidationError`
   - Check specific field errors using the `Values()` method

5. **Performance Considerations**:
   - Create validator instances once and reuse them
   - Consider caching validation results for frequently validated data
   - For high-performance needs, validate only modified fields rather than entire structs

## API Reference

### Types

```go
type Validator struct {
    // Contains unexported fields
}

type ValidationFunc func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error

type ErrorTranslatorFunc func(key string, label string, params ...string) string

type ValidationErrors url.Values
```

### Functions

```go
func New(options ...Option) (*Validator, error)
```
Creates a new Validator instance with the provided options. Returns an error if any option configuration is invalid.

```go
func WithErrorTranslator(translator ErrorTranslatorFunc) Option
```
Sets a custom error translator for the validator.

```go
func WithSeparators(ruleSep, paramSep, paramListSep string) Option
```
Sets custom separators for parsing validation rules.

```go
func WithValidators(validatorNames ...string) Option
```
Adds specific validators to the Validator instance.

```go
func WithAllValidators() Option
```
Adds all built-in validators to the Validator instance.

```go
func WithExcept(excludedNames ...string) Option
```
Adds all built-in validators except specified ones.

```go
func WithCustomValidator(name string, fn ValidationFunc) Option
```
Adds a custom validator function.

```go
func (v *Validator) RegisterValidation(tag string, fn ValidationFunc) error
```
Registers a custom validation function for this validator instance for use with the specified tag.

```go
func NewValidationError(args ...string) ValidationErrors
```
Creates a new validation error with the specified field and message pairs.

```go
func ExtractValidationErrors(err error) ValidationErrors
```
Extracts validation errors from an error if it's a validation error type.

```go
func IsValidationError(err error) bool
```
Checks if an error is a validation error.

### Methods

```go
func (v *Validator) ValidateStruct(s any) error
```
Validates the struct fields based on 'validate' tags and returns validation errors.

```go
func (e ValidationErrors) Error() string
```
Returns the error message string, implementing the error interface.

```go
func (e ValidationErrors) Values() url.Values
```
Returns the underlying url.Values containing field errors.

### Built-in Validators

| Validator   | Description                           | Example Usage                |
|-------------|---------------------------------------|------------------------------|
| required    | Field must not be empty               | `validate:"required"`        |
| email       | Must be a valid email                 | `validate:"email"`           |
| realemail   | Valid email with stricter checks      | `validate:"realemail"`       |
| range       | Value must be within range            | `validate:"range:1,100"`     |
| min         | Minimum value or length               | `validate:"min:5"`           |
| max         | Maximum value or length               | `validate:"max:100"`         |
| regex       | Must match regular expression         | `validate:"regex:^[a-z]+$"`  |
| in          | Must be one of specified values       | `validate:"in:a,b,c"`        |
| notin       | Must not be in specified values       | `validate:"notin:a,b,c"`     |
| length      | Must be exact length                  | `validate:"length:10"`       |
| between     | Value between min and max (inclusive) | `validate:"between:5,10"`    |
| url         | Valid URL                             | `validate:"url"`             |
| phone       | Valid phone number                    | `validate:"phone"`           |
| uuid        | Valid UUID                            | `validate:"uuid"`            |
| date        | Valid date in specified format        | `validate:"date:2006-01-02"` |
| creditcard  | Valid credit card number              | `validate:"creditcard"`      |
| hexcolor    | Valid hex color code                  | `validate:"hexcolor"`        |
| ip          | Valid IP address                      | `validate:"ip"`              |
| alpha       | Letters only                          | `validate:"alpha"`           |
| alphanum    | Letters and numbers only              | `validate:"alphanum"`        |
| numeric     | Numbers only                          | `validate:"numeric"`         |
| username    | Valid username format                 | `validate:"username"`        |
| password    | Meets password complexity rules       | `validate:"password"`        |
| fullname    | Valid full name format                | `validate:"fullname"`        |
| slug        | Valid URL slug                        | `validate:"slug"`            |
| boolean     | Boolean value                         | `validate:"boolean"`         |
| eq          | Equal to value                        | `validate:"eq:100"`          |
| ne          | Not equal to value                    | `validate:"ne:0"`            |
| gt          | Greater than value                    | `validate:"gt:0"`            |
| gte         | Greater than or equal to value        | `validate:"gte:1"`           |
| lt          | Less than value                       | `validate:"lt:100"`          |
| lte         | Less than or equal to value           | `validate:"lte:99"`          |
