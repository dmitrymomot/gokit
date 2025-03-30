# Privacy Package

The `privacy` package provides a comprehensive set of tools for data masking and privacy protection in Go applications. It offers a flexible and extensible API for masking sensitive information across different data types and categories.

## Features

- **Flexible Masking Strategies**: Support for different data masking approaches (redaction, partial masking)
- **Type-Specific Maskers**: Specialized maskers for common data types (strings, emails, credit cards, phone numbers)
- **Customizable Options**: Fine-grained control over masking behavior via functional options pattern
- **Category-Based Masking**: Organize masking rules by data categories (PII, sensitive data, etc.)
- **Auto-Detection**: Automatically detect and mask sensitive data types
- **Context-Aware**: All operations support context for cancellation and timeouts
- **Thread-Safe**: All components are designed to be safe for concurrent use
- **Robust Error Handling**: Comprehensive set of predefined errors for precise error reporting

## Installation

```bash
go get github.com/dmitrymomot/gokit/privacy
```

## Usage

### Basic String Masking

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dmitrymomot/gokit/privacy"
)

func main() {
    ctx := context.Background()
    
    // Create a basic string masker that shows first and last character
    masker, err := privacy.NewStringMasker(
        privacy.WithStrategy(privacy.StrategyPartialMask),
        privacy.WithReplacement('*'),
        privacy.WithVisibleChars(1, 1),
    )
    if err != nil {
        panic(err)
    }
    
    // Mask a sensitive string
    result, err := masker.Mask(ctx, "SecretPassword")
    if err != nil {
        panic(err)
    }
    
    fmt.Println(result) // Output: S***********d
}
```

### Email Masking

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dmitrymomot/gokit/privacy"
)

func main() {
    ctx := context.Background()
    
    // Create an email masker with custom configuration
    masker, err := privacy.NewEmailMasker(
        privacy.WithEmailReplacement('*'),       // Character to use for masking
        privacy.WithVisibleLocalChars(2),        // Show first 2 characters of local part
        privacy.WithVisibleDomainChars(0),       // Show 0 characters of domain (before extension)
        privacy.WithShowDomainExt(true),         // Show domain extension
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
}
```

### Credit Card Masking

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dmitrymomot/gokit/privacy"
)

func main() {
    ctx := context.Background()
    
    // Create a credit card masker with custom configuration
    masker, err := privacy.NewCardMasker(
        privacy.WithCardReplacement('*'),        // Character to use for masking
        privacy.WithVisibleEndDigits(4),         // Show last 4 digits (PCI-DSS compliant)
        privacy.WithCardFormatting(true),        // Preserve original formatting with spaces
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
    
    // Handles special formats like AMEX
    amexResult, _ := masker.Mask(ctx, "3782 822463 10005")
    fmt.Println(amexResult) // Output: **** ***** *0005
}
```

### Phone Number Masking

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dmitrymomot/gokit/privacy"
)

func main() {
    ctx := context.Background()
    
    // Create a phone masker with custom configuration
    masker, err := privacy.NewPhoneMasker(
        privacy.WithPhoneReplacement('*'),       // Character to use for masking
        privacy.WithVisiblePhoneDigits(4),       // Show last 4 digits
        privacy.WithPreserveFormat(true),        // Preserve original formatting
        privacy.WithPreserveCountryCode(true),   // Keep country code visible
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
    
    // Example without preserving country code
    maskerNoCountry, _ := privacy.NewPhoneMasker(
        privacy.WithPreserveCountryCode(false),
    )
    noCountryResult, _ := maskerNoCountry.Mask(ctx, "+1 555-123-4567")
    fmt.Println(noCountryResult) // Output: ** ***-***-4567
}
```

### Using the Masking Registry

The masking registry provides a centralized way to manage different maskers and apply them based on data categories.

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/dmitrymomot/gokit/privacy"
)

func main() {
    ctx := context.Background()
    
    // Create a default string masker to use when no specific masker is found
    defaultMasker, _ := privacy.NewStringMasker(
        privacy.WithStrategy(privacy.StrategyRedact),
    )
    
    // Create masking registry with default masker
    registry := privacy.NewMaskingRegistry(defaultMasker)
    
    // Create and register maskers for different data categories
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
    
    // Using the default masker for an unregistered category
    otherResult, _ := registry.MaskByCategory(ctx, privacy.CategoryName, "John Smith")
    fmt.Println(otherResult)  // Output: [REDACTED]
}
```

### Auto-Detection and Masking

The auto-masking registry extends the standard registry with automatic detection capabilities:

```go
package main

import (
    "context"
    "fmt"
    "regexp"
    
    "github.com/dmitrymomot/gokit/privacy"
)

func main() {
    ctx := context.Background()
    
    // Create auto-masking registry
    registry := privacy.NewAutoMaskingRegistry(nil)
    
    // Register maskers
    emailMasker, _ := privacy.NewEmailMasker()
    cardMasker, _ := privacy.NewCardMasker()
    phoneMasker, _ := privacy.NewPhoneMasker()
    
    registry.RegisterMasker(privacy.CategoryEmail, emailMasker)
    registry.RegisterMasker(privacy.CategoryCreditCard, cardMasker)
    registry.RegisterMasker(privacy.CategoryPhone, phoneMasker)
    
    // Define detection rules
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    cardRegex := regexp.MustCompile(`^[\d\s-]{13,19}$`)
    phoneRegex := regexp.MustCompile(`^[\d\s\+\-\(\)]{7,15}$`)
    
    // Register detection rules
    registry.RegisterDetectionRule(privacy.CategoryEmail, func(data any) bool {
        str, ok := data.(string)
        return ok && emailRegex.MatchString(str)
    })
    
    registry.RegisterDetectionRule(privacy.CategoryCreditCard, func(data any) bool {
        str, ok := data.(string)
        return ok && cardRegex.MatchString(str)
    })
    
    registry.RegisterDetectionRule(privacy.CategoryPhone, func(data any) bool {
        str, ok := data.(string)
        return ok && phoneRegex.MatchString(str)
    })
    
    // Auto-detect and mask data
    email := "user@example.com"
    card := "4111 1111 1111 1111"
    phone := "+1 (555) 123-4567"
    
    emailResult, _ := registry.AutoMask(ctx, email)
    cardResult, _ := registry.AutoMask(ctx, card)
    phoneResult, _ := registry.AutoMask(ctx, phone)
    
    fmt.Println(emailResult)  // Will auto-detect as email and mask accordingly
    fmt.Println(cardResult)   // Will auto-detect as credit card and mask accordingly
    fmt.Println(phoneResult)  // Will auto-detect as phone number and mask accordingly
}
```

## Error Handling

The privacy package provides a comprehensive set of predefined errors for precise error reporting:

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

## Available Maskers

The package includes the following specialized maskers:

1. **StringMasker**: General-purpose string masking with flexible visibility options
2. **EmailMasker**: Email-specific masking with control over local part and domain visibility
3. **CardMasker**: Credit card masking, PCI-DSS compliant with format preservation
4. **PhoneMasker**: Phone number masking with international format support

## Masking Strategies

The package supports multiple masking strategies:

- **StrategyRedact**: Completely replaces the data with a fixed string
- **StrategyPartialMask**: Keeps parts of the data visible, masking the rest
- **StrategyTokenize**: (Reserved for future implementation)
- **StrategyPseudonymize**: (Reserved for future implementation)
- **StrategyEncrypt**: (Reserved for future implementation)
- **StrategyNoise**: (Reserved for future implementation)
