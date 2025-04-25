# Privacy Package

A flexible data masking library for secure handling of sensitive information in Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/privacy
```

## Overview

The `privacy` package provides tools for masking sensitive data in Go applications. It includes specialized maskers for common data types (emails, credit cards, phone numbers, etc.) with configurable masking behaviors, category-based registry, and auto-detection capabilities.

## Features

- Multiple masking strategies (redaction, partial visibility)
- Type-specific maskers for common sensitive data types
- Category registry for organized masking rules
- Auto-detection of sensitive data patterns
- Flexible configuration via functional options
- Context support for cancellation and timeouts
- Thread-safe implementation
- Comprehensive error handling

## Usage

### Basic String Masking

```go
import (
	"context"
	"fmt"
	"github.com/dmitrymomot/gokit/privacy"
)

func main() {
	// Create a string masker that shows first and last character
	masker, err := privacy.NewStringMasker(
		privacy.WithStrategy(privacy.StrategyPartialMask),
		privacy.WithReplacement('*'),
		privacy.WithVisibleChars(1, 1),
	)
	if err != nil {
		panic(err)
	}
	
	// Mask a sensitive string
	ctx := context.Background()
	result, err := masker.Mask(ctx, "SecretPassword")
	if err != nil {
		panic(err)
	}
	
	fmt.Println(result) // Output: S***********d
}
```

### Email Masking

```go
// Create an email masker with custom configuration
masker, err := privacy.NewEmailMasker(
	privacy.WithEmailReplacement('*'),      // Character for masking
	privacy.WithVisibleLocalChars(2),       // Show first 2 chars of local part
	privacy.WithVisibleDomainChars(0),      // Hide domain (except extension)
	privacy.WithShowDomainExt(true),        // Show domain extension
)
if err != nil {
	panic(err)
}

// Mask an email address
result, err := masker.Mask(ctx, "john.doe@example.com")
if err != nil {
	panic(err)
}

fmt.Println(result) // Output: jo******@example.com
```

### Credit Card Masking

```go
// Create a credit card masker (PCI-DSS compliant)
masker, err := privacy.NewCardMasker(
	privacy.WithCardReplacement('*'),       // Character for masking
	privacy.WithVisibleEndDigits(4),        // Show last 4 digits
	privacy.WithCardFormatting(true),       // Preserve formatting
)
if err != nil {
	panic(err)
}

// Mask a credit card number
result, err := masker.Mask(ctx, "4111 1111 1111 1111")
if err != nil {
	panic(err)
}

fmt.Println(result) // Output: **** **** **** 1111

// Also handles special formats like AMEX
amexResult, _ := masker.Mask(ctx, "3782 822463 10005")
fmt.Println(amexResult) // Output: **** ***** *0005
```

### Phone Number Masking

```go
// Create a phone number masker
masker, err := privacy.NewPhoneMasker(
	privacy.WithPhoneReplacement('*'),      // Character for masking
	privacy.WithVisiblePhoneDigits(4),      // Show last 4 digits
	privacy.WithPreserveFormat(true),       // Keep formatting
	privacy.WithPreserveCountryCode(true),  // Keep country code visible
)
if err != nil {
	panic(err)
}

// Mask a phone number
result, err := masker.Mask(ctx, "+1 (555) 123-4567")
if err != nil {
	panic(err)
}

fmt.Println(result) // Output: +1 (***) ***-4567
```

### Using the Masking Registry

```go
// Create a masking registry with default fallback masker
defaultMasker, _ := privacy.NewStringMasker(
	privacy.WithStrategy(privacy.StrategyRedact),
)
registry := privacy.NewMaskingRegistry(defaultMasker)

// Register specialized maskers for different data categories
emailMasker, _ := privacy.NewEmailMasker()
cardMasker, _ := privacy.NewCardMasker()
phoneMasker, _ := privacy.NewPhoneMasker()

registry.RegisterMasker(privacy.CategoryEmail, emailMasker)
registry.RegisterMasker(privacy.CategoryCreditCard, cardMasker)
registry.RegisterMasker(privacy.CategoryPhone, phoneMasker)

// Mask data by category
emailResult, _ := registry.MaskByCategory(ctx, privacy.CategoryEmail, "user@example.com")
cardResult, _ := registry.MaskByCategory(ctx, privacy.CategoryCreditCard, "4111 1111 1111 1111")
phoneResult, _ := registry.MaskByCategory(ctx, privacy.CategoryPhone, "(555) 123-4567")

fmt.Println(emailResult)  // Output: u***@example.com
fmt.Println(cardResult)   // Output: **** **** **** 1111
fmt.Println(phoneResult)  // Output: (***) ***-4567

// Fallback to default masker for unregistered categories
nameResult, _ := registry.MaskByCategory(ctx, privacy.CategoryName, "John Smith")
fmt.Println(nameResult)  // Output: [REDACTED]
```

### Auto-Detection and Masking

```go
import (
	"context"
	"fmt"
	"regexp"
	"github.com/dmitrymomot/gokit/privacy"
)

// Create auto-masking registry
registry := privacy.NewAutoMaskingRegistry(nil)

// Register maskers
emailMasker, _ := privacy.NewEmailMasker()
cardMasker, _ := privacy.NewCardMasker()
phoneMasker, _ := privacy.NewPhoneMasker()

registry.RegisterMasker(privacy.CategoryEmail, emailMasker)
registry.RegisterMasker(privacy.CategoryCreditCard, cardMasker)
registry.RegisterMasker(privacy.CategoryPhone, phoneMasker)

// Register detection rules using regex patterns
emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
registry.RegisterDetectionRule(privacy.CategoryEmail, func(data any) bool {
	str, ok := data.(string)
	return ok && emailRegex.MatchString(str)
})

// Define and register other detection rules...

// Auto-detect and mask data
email := "user@example.com"
card := "4111 1111 1111 1111"
phone := "+1 (555) 123-4567"

emailResult, _ := registry.AutoMask(ctx, email)
cardResult, _ := registry.AutoMask(ctx, card)
phoneResult, _ := registry.AutoMask(ctx, phone)

fmt.Println(emailResult)  // Auto-detected and masked as email
fmt.Println(cardResult)   // Auto-detected and masked as credit card
fmt.Println(phoneResult)  // Auto-detected and masked as phone number
```

## API Reference

### Masker Interface

```go
// Masker defines the interface for all masking operations
type Masker interface {
	Mask(ctx context.Context, data any) (string, error)
}
```

### Available Maskers

```go
// Create specialized maskers with their configuration options
StringMasker   // General purpose string masking
EmailMasker    // Email-specific masking
CardMasker     // Credit card number masking (PCI-DSS compliant)
PhoneMasker    // Phone number masking
```

### Masking Registry

```go
// MaskingRegistry manages multiple maskers by category
type MaskingRegistry interface {
	RegisterMasker(category Category, masker Masker)
	MaskByCategory(ctx context.Context, category Category, data any) (string, error)
	GetMasker(category Category) Masker
}

// AutoMaskingRegistry extends MaskingRegistry with detection capabilities
type AutoMaskingRegistry interface {
	MaskingRegistry
	RegisterDetectionRule(category Category, rule DetectionRule)
	AutoMask(ctx context.Context, data any) (string, error)
}
```

### Available Categories

```go
// Predefined data categories
const (
	CategoryEmail      Category = "email"
	CategoryCreditCard Category = "credit_card"
	CategoryPhone      Category = "phone"
	CategoryName       Category = "name"
	CategoryAddress    Category = "address"
	CategoryPassword   Category = "password"
	CategorySSN        Category = "ssn"
	CategoryCustom     Category = "custom"
)
```

### Masking Strategies

```go
// Available masking strategies
const (
	StrategyRedact      Strategy = "redact"       // Complete replacement
	StrategyPartialMask Strategy = "partial_mask" // Show portions of data
)
```

### Error Types

```go
// Common error types for precise error handling
var (
	ErrInvalidData      = errors.New("invalid data")
	ErrUnsupportedType  = errors.New("unsupported type")
	ErrInvalidMask      = errors.New("invalid mask configuration")
	ErrInvalidCategory  = errors.New("invalid category")
	ErrMaskerNotFound   = errors.New("masker not found")
)
```

## Error Handling

```go
// Handle masking errors appropriately
result, err := masker.Mask(ctx, data)
if err != nil {
	switch {
	case errors.Is(err, privacy.ErrInvalidData):
		// Handle invalid data error
	case errors.Is(err, privacy.ErrUnsupportedType):
		// Handle unsupported type error
	case errors.Is(err, privacy.ErrInvalidMask):
		// Handle invalid mask configuration
	default:
		// Handle other errors
	}
}
```

## Best Practices

1. **Security**:
   - Always sanitize logs and outputs containing sensitive information
   - Choose appropriate masking strategies based on data sensitivity
   - Don't rely solely on masking for highly sensitive data (consider encryption)

2. **Performance**:
   - Cache masker instances for reuse across requests
   - Consider using the registry pattern to organize maskers for complex applications
   - For high-volume applications, limit auto-detection to essential cases

3. **Error Handling**:
   - Always check for errors when masking sensitive data
   - Use the predefined error types to handle specific failure cases
   - Log masking failures appropriately without revealing sensitive data

4. **Configuration**:
   - Use default maskers for simple cases
   - Customize maskers based on your application's security requirements
   - Follow industry standards (e.g., PCI-DSS for credit cards)
