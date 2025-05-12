# Validator Package

A flexible, thread-safe validation system for Go structs using struct tags with customizable error messages.

## Installation

```bash
go get github.com/dmitrymomot/gokit/validator
```

## Overview

The `validator` package provides a robust validation system for Go structs using struct tags. It includes 30+ built-in validators for common validation scenarios and supports custom validation rules through a simple registration mechanism. The package is thread-safe for concurrent validation and offers flexible error message customization through translator functions.

## Features

- Simple struct tag-based validation with clear syntax
- 30+ built-in validators for common validation scenarios
- Custom validation rule support with simple registration mechanism
- Customizable error messages with translator function support
- Customizable field names in error messages using struct tags
- Recursive validation for nested structs and slices
- Thread-safe implementation for concurrent use

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

// Create a validator with default error messages
v, err := validator.New(validator.WithAllValidators())
if err != nil {
    // Handle initialization error
}

user := User{
    Username: "john_doe",
    Email: "invalid-email",
    Age: 15,
    Password: "weak",
    Phone: "123",
    Website: "example",
}

// Validate the struct
err = v.ValidateStruct(user)
if err != nil {
    // err will contain validation errors for multiple fields:
    // - Email: not a valid email format
    // - Age: not in range 18-100
    // - Password: doesn't meet complexity requirements
    // - Phone: invalid phone number format
    // - Website: not a valid URL
}
```

### Custom Error Translator

```go
import (
    "fmt"
    "github.com/dmitrymomot/gokit/validator"
)

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
    // Handle initialization error
}
// Now validation errors will use your custom messages
```

### Custom Field Names

```go
import "github.com/dmitrymomot/gokit/validator"

// Create a validator that uses JSON field names in error messages
v, err := validator.New(validator.WithFieldNameTag("json"), validator.WithAllValidators())
if err != nil {
    // Handle initialization error
}

type User struct {
    Username string `json:"user_name" validate:"required" label:"Username"`
    Email    string `json:"email_address" validate:"required,email" label:"Email"`
}

user := User{
    Username: "", // Empty, will trigger required validation
    Email:    "invalid-email",
}

err = v.ValidateStruct(user)
// Error will contain:
// - "user_name": "Username field is required"
// - "email_address": "Email must be a valid email"
// Instead of "Username" and "Email" field names
```

### Custom Validation Rules

```go
import (
    "errors"
    "reflect"
    "regexp"
    "github.com/dmitrymomot/gokit/validator"
)

// Register a custom zipcode validator
err := v.RegisterValidation("zipcode", func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
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
import "github.com/dmitrymomot/gokit/validator"

type Address struct {
    Street  string `validate:"required" label:"Street"`
    City    string `validate:"required" label:"City"`
    Country string `validate:"required" label:"Country"`
}

type Customer struct {
    Name      string    `validate:"required,fullname" label:"Full Name"`
    Email     string    `validate:"required,email" label:"Email"`
    Address   Address   `validate:"required"` // Validates nested struct
    ShipAddrs []Address `validate:"required"` // Validates slice of structs
}

customer := Customer{
    Name:  "John Doe",
    Email: "john.doe@example.com",
    Address: Address{
        Street: "123 Main St",
        // Missing City and Country fields
    },
    // Missing ShipAddrs slice
}

// Validate recursive structures
err := v.ValidateStruct(customer)
// Returns errors for:
// - Address.City: required field missing
// - Address.Country: required field missing
// - ShipAddrs: required field missing
```

### Error Handling

```go
import (
    "fmt"
    "github.com/dmitrymomot/gokit/validator"
)

err := v.ValidateStruct(user)
if err != nil {
    // Check if it's a validation error
    if validator.IsValidationError(err) {
        // Extract structured validation errors
        validationErrs := validator.ExtractValidationErrors(err)

        // Access specific field errors
        for field, fieldErr := range validationErrs.Values() {
            fmt.Printf("Field '%s': %s\n", field, fieldErr[0])
            // Field 'Email': must be a valid email address
            // Field 'Age': must be between 18 and 100
        }

        // Handle specific field error
        if emailErrs, ok := validationErrs.Values()["Email"]; ok {
            // Handle email-specific errors
        }
    } else {
        // Handle other types of errors
    }
}
```

## Best Practices

1. **Field Naming and Labeling**:

    - Use the `label` tag to provide user-friendly field names for error messages
    - Example: `label:"Email Address"` instead of just `label:"email"`
    - When no label is provided, the field name will be used
    - For consistent field naming in errors, use `WithFieldNameTag` to match JSON, XML, or form field names
    - Example: Configure with `WithFieldNameTag("json")` to use `json:"field_name"` tag values in errors

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

### Constructor Functions

```go
func New(options ...Option) (*Validator, error)
```

Creates a new Validator instance with the provided options. Returns an error if any option configuration is invalid.

### Options

```go
func WithErrorTranslator(translator ErrorTranslatorFunc) Option
```

Sets a custom error translator function for validation error messages.

```go
func WithSeparators(ruleSep, paramSep, paramListSep string) Option
```

Sets custom separators for parsing validation rules (default: "," for rules, ":" for params, "," for param lists).

```go
func WithValidators(validatorNames ...string) Option
```

Adds specific named validators to the Validator instance.

```go
func WithAllValidators() Option
```

Adds all built-in validators to the Validator instance (recommended for most use cases).

```go
func WithExcept(excludedNames ...string) Option
```

Adds all built-in validators except specifically excluded ones.

```go
func WithCustomValidator(name string, fn ValidationFunc) Option
```

Adds a custom validator function at initialization time.

```go
func WithFieldNameTag(tagName string) Option
```

Sets the tag name used for identifying field names in validation errors. If a field has this tag, its value will be used in error messages instead of the field name.

### Validator Methods

```go
func (v *Validator) RegisterValidation(tag string, fn ValidationFunc) error
```

Registers a custom validation function for use with the specified tag name.

```go
func (v *Validator) ValidateStruct(s any) error
```

Validates struct fields based on 'validate' tags and returns validation errors.

### Error Handling Functions

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

### Error Methods

```go
func (e ValidationErrors) Error() string
```

Returns the error message string, implementing the error interface.

```go
func (e ValidationErrors) Values() url.Values
```

Returns the underlying url.Values containing field errors for access by field name.

### Built-in Validators

| Validator     | Description                                      | Example Usage                  |
| ------------- | ------------------------------------------------ | ------------------------------ |
| required      | Field must not be empty                          | `validate:"required"`          |
| email         | Must be a valid email                            | `validate:"email"`             |
| realemail     | Valid email with stricter checks (alias: remail) | `validate:"realemail"`         |
| range         | Value must be within range                       | `validate:"range:1,100"`       |
| min           | Minimum value or length                          | `validate:"min:5"`             |
| max           | Maximum value or length                          | `validate:"max:100"`           |
| regex         | Must match regular expression                    | `validate:"regex:^[a-z]+$"`    |
| in            | Must be one of specified values                  | `validate:"in:a,b,c"`          |
| notin         | Must not be in specified values                  | `validate:"notin:a,b,c"`       |
| length        | Must be exact length (alias: len)                | `validate:"length:10"`         |
| between       | Value between min and max (alias: btw)           | `validate:"between:5,10"`      |
| url           | Valid URL                                        | `validate:"url"`               |
| phone         | Valid phone number                               | `validate:"phone"`             |
| uuid          | Valid UUID                                       | `validate:"uuid"`              |
| date          | Valid date in specified format                   | `validate:"date:2006-01-02"`   |
| pastdate      | Date must be in the past (alias: past)           | `validate:"pastdate"`          |
| futuredate    | Date must be in the future (alias: future)       | `validate:"futuredate"`        |
| workday       | Date must be a work day (alias: wday)            | `validate:"workday"`           |
| weekend       | Date must be a weekend (alias: wend)             | `validate:"weekend"`           |
| creditcard    | Valid credit card number (alias: cc)             | `validate:"creditcard"`        |
| hexcolor      | Valid hex color code (alias: hcolor)             | `validate:"hexcolor"`          |
| ip            | Valid IP address (v4 or v6)                      | `validate:"ip"`                |
| ipv4          | Valid IPv4 address                               | `validate:"ipv4"`              |
| ipv6          | Valid IPv6 address                               | `validate:"ipv6"`              |
| domain        | Valid domain name                                | `validate:"domain"`            |
| mac           | Valid MAC address                                | `validate:"mac"`               |
| port          | Valid port number                                | `validate:"port"`              |
| alpha         | Letters only                                     | `validate:"alpha"`             |
| alphanum      | Letters and numbers only                         | `validate:"alphanum"`          |
| alphaspace    | Letters and spaces (alias: aspace)               | `validate:"alphaspace"`        |
| alphaspacenum | Letters, numbers and spaces (alias: aspacenum)   | `validate:"alphaspacenum"`     |
| numeric       | Numbers only                                     | `validate:"numeric"`           |
| username      | Valid username format (alias: uname)             | `validate:"username"`          |
| password      | Meets password complexity rules (alias: pwd)     | `validate:"password"`          |
| fullname      | Valid full name format (alias: fname)            | `validate:"fullname"`          |
| name          | Valid name format                                | `validate:"name"`              |
| slug          | Valid URL slug                                   | `validate:"slug"`              |
| boolean       | Boolean value (alias: bool)                      | `validate:"boolean"`           |
| ascii         | Contains only ASCII characters                   | `validate:"ascii"`             |
| base64        | Valid Base64 string (alias: b64)                 | `validate:"base64"`            |
| json          | Valid JSON string                                | `validate:"json"`              |
| semver        | Valid semantic version                           | `validate:"semver"`            |
| extension     | File has specific extension (alias: ext)         | `validate:"extension:jpg,png"` |
| starts        | Starts with substring (alias: startswith)        | `validate:"starts:foo"`        |
| endswith      | Ends with substring (alias: ends)                | `validate:"endswith:bar"`      |
| contains      | Contains substring                               | `validate:"contains:example"`  |
| notcontains   | Does not contain substring (alias: notcont)      | `validate:"notcontains:xyz"`   |
| positive      | Number must be positive (alias: pos)             | `validate:"positive"`          |
| negative      | Number must be negative (alias: neg)             | `validate:"negative"`          |
| even          | Number must be even                              | `validate:"even"`              |
| odd           | Number must be odd                               | `validate:"odd"`               |
| multiple      | Number must be multiple of (alias: mult)         | `validate:"multiple:3"`        |
| eq            | Equal to value                                   | `validate:"eq:100"`            |
| ne            | Not equal to value                               | `validate:"ne:0"`              |
| gt            | Greater than value                               | `validate:"gt:0"`              |
| gte           | Greater than or equal to value                   | `validate:"gte:1"`             |
| lt            | Less than value                                  | `validate:"lt:100"`            |
| lte           | Less than or equal to value                      | `validate:"lte:99"`            |

### Error Types

```go
var ErrInvalidTagFormat = errors.New("invalid validation tag format")
var ErrValidatorNotRegistered = errors.New("validator not registered")
```
