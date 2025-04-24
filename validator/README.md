# Validator Package

The validator package provides a robust and flexible validation system for Go structs. It supports a wide range of built-in validators and allows for custom validation rules.

## Features

- Simple struct tag-based validation
- 25+ built-in validators
- Custom validation rules support
- Customizable error messages with translation support
- Thread-safe validation
- Support for nested structs
- Comprehensive field type support

## Installation

```bash
go get github.com/dmitrymomot/gokit/validator
```

## Basic Usage

```go
package main

import (
    "fmt"
    "github.com/dmitrymomot/gokit/validator"
)

type User struct {
    Username  string `validate:"required,username" label:"Username"`
    Email     string `validate:"required,email" label:"Email Address"`
    Age       int    `validate:"required,range:18,100" label:"Age"`
    Password  string `validate:"required,password" label:"Password"`
    Phone     string `validate:"phone" label:"Phone Number"`
    Website   string `validate:"url" label:"Website"`
}

func main() {
    // Create a new validator instance
    v := validator.NewValidator(nil) // Use default error translator

    user := User{
        Username: "john_doe",
        Email: "invalid-email",
        Age: 15,
        Password: "weak",
        Phone: "invalid-phone",
        Website: "invalid-url",
    }

    // Validate the struct
    if err := v.ValidateStruct(user); err != nil {
        // Handle validation errors
        fmt.Printf("Validation errors: %v\n", err)
    }
}
```

## Built-in Validators

| Validator  | Description                           | Example Usage                |
| ---------- | ------------------------------------- | ---------------------------- |
| required   | Field must not be empty               | `validate:"required"`        |
| email      | Must be a valid email address         | `validate:"email"`           |
| min        | Minimum value/length                  | `validate:"min:5"`           |
| max        | Maximum value/length                  | `validate:"max:100"`         |
| range      | Value must be within range            | `validate:"range:1,10"`      |
| regex      | Must match regular expression         | `validate:"regex:^[0-9]+$"`  |
| numeric    | Must be numeric                       | `validate:"numeric"`         |
| alpha      | Must contain only letters             | `validate:"alpha"`           |
| alphanum   | Must contain only letters and numbers | `validate:"alphanum"`        |
| url        | Must be a valid URL                   | `validate:"url"`             |
| ip         | Must be a valid IP address            | `validate:"ip"`              |
| date       | Must be a valid date                  | `validate:"date:2006-01-02"` |
| in         | Must be one of the values             | `validate:"in:a,b,c"`        |
| notin      | Must not be one of the values         | `validate:"notin:x,y,z"`     |
| length     | Must have exact length                | `validate:"length:10"`       |
| between    | Must be between min and max           | `validate:"between:5,10"`    |
| boolean    | Must be a boolean                     | `validate:"boolean"`         |
| uuid       | Must be a valid UUID                  | `validate:"uuid"`            |
| creditcard | Must be a valid credit card number    | `validate:"creditcard"`      |
| password   | Must meet password requirements       | `validate:"password"`        |
| phone      | Must be a valid phone number          | `validate:"phone"`           |
| username   | Must be a valid username              | `validate:"username"`        |
| slug       | Must be a valid slug                  | `validate:"slug"`            |
| hexcolor   | Must be a valid hex color             | `validate:"hexcolor"`        |
| fullname   | Must be a valid full name             | `validate:"fullname"`        |
| eq         | Must be equal to the specified value         | `validate:"eq:expected"`         |
| ne         | Must not be equal to the specified value     | `validate:"ne:unexpected"`       |
| lt         | Must be less than the specified value        | `validate:"lt:10"`               |
| lte        | Must be less than or equal to the value      | `validate:"lte:10"`              |
| gt         | Must be greater than the specified value     | `validate:"gt:1"`                |
| gte        | Must be greater than or equal to the value   | `validate:"gte:1"`               |
| len        | Must have exact length (alias for length)    | `validate:"len:10"`              |
| realemail  | Must be a real, deliverable email address    | `validate:"realemail"`           |

## Custom Validation

You can register custom validation functions:

```go
validator.RegisterValidation("custom", func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
    // Your custom validation logic here
    return nil
})
```

## Custom Error Messages

You can provide custom error messages by implementing a translator function:

```go
translator := func(key string, label string, params ...any) string {
    switch key {
    case "validation.required":
        return fmt.Sprintf("%s is required", label)
    case "validation.email":
        return fmt.Sprintf("%s must be a valid email address", label)
    default:
        return fmt.Sprintf("Invalid %s", label)
    }
}

v := validator.NewValidator(translator)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
