# Privacy Package

A flexible data masking library for secure handling of sensitive information in Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/privacy
```

## Overview

The `privacy` package provides tools for masking sensitive data in Go applications. It offers a comprehensive set of strategies for data masking, including specialized implementations for common sensitive data types (emails, credit cards, phones) and a registry system for easy management of masking rules. The package is designed to be thread-safe, flexible, and easy to integrate into existing applications.

## Features

- Multiple masking strategies (redaction, partial visibility, tokenization, pseudonymization)
- Type-specific maskers for common sensitive data types (emails, credit cards, phones)
- Category-based registry for organized masking rules
- Auto-detection of sensitive data patterns
- Flexible configuration via functional options
- Thread-safe implementation for concurrent usage
- Comprehensive error handling with specific error types

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
ctx := context.Background()
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
ctx := context.Background()
result, err := masker.Mask(ctx, "4111 1111 1111 1111")
if err != nil {
	panic(err)
}

fmt.Println(result) // Output: **** **** **** 1111

// Also handles special formats like AMEX
amexResult, _ := masker.Mask(ctx, "3782 822463 10005")
fmt.Println(amexResult) // Output: **** ***** *0005
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
ctx := context.Background()
emailResult, _ := registry.MaskByCategory(ctx, privacy.CategoryEmail, "user@example.com")
cardResult, _ := registry.MaskByCategory(ctx, privacy.CategoryCreditCard, "4111 1111 1111 1111")
phoneResult, _ := registry.MaskByCategory(ctx, privacy.CategoryPhone, "(555) 123-4567")

fmt.Println(emailResult)  // Masked email output
fmt.Println(cardResult)   // Masked card output
fmt.Println(phoneResult)  // Masked phone output
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
ctx := context.Background()
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

## Best Practices

1. **Security Considerations**:
   - Always sanitize logs and outputs containing sensitive information
   - Choose appropriate masking strategies based on data sensitivity
   - Don't rely solely on masking for highly sensitive data (consider encryption)
   - Regularly review and update detection patterns for evolving data formats

2. **Performance Optimization**:
   - Cache masker instances for reuse across requests
   - Consider using the registry pattern to organize maskers for complex applications
   - For high-volume applications, limit auto-detection to essential cases
   - Pre-register all required maskers during application initialization

3. **Error Handling**:
   - Always check for errors when masking sensitive data
   - Use the predefined error types to handle specific failure cases
   - Log masking failures appropriately without revealing sensitive data
   - Implement graceful fallbacks when masking fails

4. **Configuration Best Practices**:
   - Use default maskers for simple cases
   - Customize maskers based on your application's security requirements
   - Follow industry standards (e.g., PCI-DSS for credit cards)
   - Document which masking strategies are used for different data types

## API Reference

### Masker Interface

```go
// Masker defines the interface for all masking operations
type Masker interface {
	Mask(ctx context.Context, data any) (any, error)
	CanMask(data any) bool
}
```

### Masking Strategies

```go
// Available masking strategies
const (
	StrategyRedact      MaskingStrategy = "redact"       // Complete replacement
	StrategyPartialMask MaskingStrategy = "partial"      // Show portions of data
	StrategyTokenize    MaskingStrategy = "tokenize"     // Replace with token
	StrategyPseudonymize MaskingStrategy = "pseudonymize" // Replace with pseudonym
	StrategyEncrypt     MaskingStrategy = "encrypt"      // Encrypt the data
	StrategyNoise       MaskingStrategy = "noise"        // Add statistical noise
)
```

### Data Categories

```go
// Predefined data categories
const (
	// General categories
	CategoryPII           DataCategory = "pii"
	CategoryFinancial     DataCategory = "financial"
	CategoryHealth        DataCategory = "health"
	CategoryCredentials   DataCategory = "credentials"
	CategoryLocation      DataCategory = "location"
	CategoryCommunication DataCategory = "communication"
	
	// Specific data categories
	CategoryCreditCard    DataCategory = "credit_card"
	CategoryEmail         DataCategory = "email"
	CategoryPhone         DataCategory = "phone"
	CategorySSN           DataCategory = "ssn"
	CategoryName          DataCategory = "name"
	CategoryAddress       DataCategory = "address"
	CategoryPassport      DataCategory = "passport"
	CategoryDriverLicense DataCategory = "driver_license"
)
```

### Registry Interfaces

```go
// MaskerRegistry manages multiple maskers by category
type MaskerRegistry interface {
	RegisterMasker(category DataCategory, masker Masker) error
	GetMasker(category DataCategory) (Masker, error)
	MaskByCategory(ctx context.Context, category DataCategory, data any) (any, error)
}

// AutoMaskingRegistry extends MaskingRegistry with detection capabilities
type AutoMaskingRegistry struct {
	*MaskingRegistry
	detectionRules map[DataCategory]DetectionRule
}

// DetectionRule is used to detect if data belongs to a specific category
type DetectionRule func(data any) bool
```

### Error Types

```go
// Common error types for precise error handling
var (
	// General masking errors
	ErrInvalidData         = errors.New("privacy: invalid data for masking")
	ErrInvalidMask         = errors.New("privacy: invalid mask configuration")
	ErrUnsupportedType     = errors.New("privacy: unsupported data type")
	ErrMaskingFailed       = errors.New("privacy: masking operation failed")
	
	// Registry errors
	ErrMaskerNotFound      = errors.New("privacy: masker not found for category")
	ErrInvalidMasker       = errors.New("privacy: invalid masker")
	ErrInvalidRule         = errors.New("privacy: invalid detection rule")
	ErrRuleNotFound        = errors.New("privacy: detection rule not found")
	ErrCategoryNotDetected = errors.New("privacy: data category could not be detected")
	
	// Configuration errors
	ErrInvalidRegex        = errors.New("privacy: invalid regex pattern")
	ErrNegativeCharCount   = errors.New("privacy: visible character count cannot be negative")
	ErrNegativeDigitCount  = errors.New("privacy: visible digit count cannot be negative")
	ErrTooManyVisibleDigits = errors.New("privacy: showing more than 4 digits is not allowed for security reasons")
	ErrNegativeMinLength   = errors.New("privacy: minimum length cannot be negative")
)
```

### Available Maskers

```go
// Create specialized maskers with their configuration options
func NewStringMasker(options ...Option) (*StringMasker, error)
func NewEmailMasker(options ...Option) (*EmailMasker, error)
func NewCardMasker(options ...Option) (*CardMasker, error)
func NewPhoneMasker(options ...Option) (*PhoneMasker, error)
```

### Registry Creation

```go
// Create new registries
func NewMaskingRegistry(defaultMasker Masker) *MaskingRegistry
func NewAutoMaskingRegistry(defaultMasker Masker) *AutoMaskingRegistry
