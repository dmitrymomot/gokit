# README Creation Guide for Go Packages

## Purpose

Create a concise, informative README.md file that clearly explains your package's purpose, features, and usage to other developers. A well-structured README serves as an excellent context source for providing accurate code suggestions.

**IMPORTANT: ALWAYS examine the codebase FIRST. ONLY document functionality that EXISTS in the code and NEVER include speculative or planned features.**

## README Structure Template

```markdown
# Package Name

A one-line description of the package's core purpose.

## Installation

```bash
go get package-name
```

Take package name from the go.mod file and import path if applicable.

## Overview

A 2-3 sentence paragraph explaining what the package does and its primary use case. Mention if the package is thread-safe and any other important high-level characteristics.

## Features

- Bullet point list of key capabilities (4-6 items)
- Focus on what makes the package useful
- Keep each point brief (1 line per feature)
- Include important technical characteristics (thread-safety, performance considerations)

## Usage

### Basic Example

```go
import "package/path"

// Code example showing the most common usage pattern
// Include comments that explain expected behavior rather than fmt.Println outputs
// Example:
result := package.Function("input")
// Returns: "expected output"
```

### Additional Usage Scenarios

```go
// Examples of other important usage patterns
// Use comments to show expected outputs rather than print statements
```

### Error Handling

```go
// Show how to properly handle errors from the package
// Include examples of checking specific error types
if err != nil {
    switch {
    case errors.Is(err, package.ErrSpecificError):
        // Handle specific error
    default:
        // Handle general error
    }
}
```

## Best Practices

1. **Category Name**:
   - Specific recommendation
   - Related recommendation

2. **Another Category**:
   - Group related recommendations together
   - Provide specific, actionable advice

## API Reference

### Configuration Variables

```go
// List package-level variables that users can configure
var ConfigOption = defaultValue // Brief explanation
```

### Types

```go
// Use code blocks for type definitions to make them more parsable
type TypeName struct {
    Field1 string // Field description
    Field2 int    // Field description
}
```

### Functions

```go
// List functions with full signatures and brief descriptions
func FunctionName(param1 Type, param2 Type) (ReturnType, error)
```

Brief explanation of what the function does, its parameters, and return values.

### Methods

```go
// List methods with full signatures
func (receiver *Type) MethodName(param Type) ReturnType
```

Brief explanation of the method's purpose.

### Error Types

```go
// List all exported error variables
var ErrSpecificError = errors.New("specific error description")
```

## Writing Guidelines

1. **Package Name and Description**
   - Use a clear, descriptive title (e.g., "API Key Package")
   - Provide a concise one-line description that explains the package's purpose

2. **Features Section**
   - List 4-6 key features of the package
   - Focus on capabilities, not implementation details
   - Keep bullet points short and start with action verbs
   - Include information about thread-safety and performance characteristics

3. **Usage Section**
   - Start with the most common use case
   - Include complete, runnable code examples with imports
   - Add comments within code to explain key steps and expected outputs
   - Use comments instead of print statements to show function outputs
   - Show comprehensive error handling in examples
   - Include multiple usage scenarios for complex packages

4. **API Reference**
   - Document all exported types, functions, methods, and errors
   - Use proper Go function signature format in code blocks
   - Group related functions/methods together
   - List error variables separately for easy reference
   - Include configuration variables that can be modified

5. **Best Practices**
   - Organize best practices into logical categories
   - Include security considerations where applicable
   - Add performance tips for large-scale usage
   - Provide data consistency recommendations

6. **General Style**
   - Use consistent markdown formatting
   - Prefer short sentences and active voice
   - Keep the entire README under 300 lines
   - Include code fences with `go` language identifier
   - ENSURE all documented functions and types ACTUALLY EXIST in the codebase
   - VERIFY all function signatures MATCH the actual implementation

7. **Accuracy Verification**
   - ALWAYS verify the existence of functions, types, or variables before documenting them
   - DOUBLE-CHECK all function signatures against their actual implementation
   - DO NOT document planned or future functionality that doesn't yet exist
   - ONLY document features that are implemented and working in the current version

## Anti-Hallucination Checklist

BEFORE finalizing any README, VERIFY the following:

1. **EXISTENCE CHECK**: CONFIRM every function, method, type, and variable mentioned actually exists in the code
2. **SIGNATURE CHECK**: VERIFY all function signatures exactly match what's in the code
3. **BEHAVIOR CHECK**: VALIDATE that code examples match the actual implementation behavior
4. **ERROR CHECK**: ENSURE all documented error types and values exist in the code
5. **CONSISTENCY CHECK**: VERIFY parameter names and types are consistent throughout documentation
6. **EXAMPLE CHECK**: TEST all code examples against the current implementation
7. **CLAIM CHECK**: CONFIRM all performance, security, and thread-safety claims are supported by the code

NEVER document features based solely on function names or comments without verifying the implementation.

## Example README

Here's an abbreviated example of a well-structured README:

```markdown
# API Key Package

A secure API key generation, hashing, and validation package for authentication systems.

## Installation

```bash
go get github.com/dmitrymomot/gokit/apikey
```

## Overview

The `apikey` package provides tools for creating and managing secure API keys for authentication systems. It offers cryptographically secure key generation with multiple methods, secure hashing for storage, and constant-time validation to prevent timing attacks. This package is thread-safe and suitable for concurrent use.

## Features

- Cryptographically secure API key generation
- Time-ordered keys using UUID V7 for chronological sorting
- Secure hashing with HMAC-SHA256
- Constant-time comparison to prevent timing attacks
- Simple validation against stored hashes
- Thread-safe implementation for concurrent usage
- Comprehensive error handling

## Usage

### Generating API Keys

```go
import (
    "github.com/dmitrymomot/gokit/apikey"
)

// Generate a random API key (256 bits, hex-encoded)
apiKey, err := apikey.GenerateRandom()
if err != nil {
    // Handle error
    return fmt.Errorf("failed to generate API key: %w", err)
}
// apiKey = "a1b2c3d4..." (64-character hexadecimal string)

// Alternative: Generate a time-ordered API key using UUID V7
// Useful for keys that need to be sortable by creation time
orderedKey, err := apikey.GenerateTimeOrdered()
if err != nil {
    // Handle error
    return fmt.Errorf("failed to generate time-ordered key: %w", err)
}
// orderedKey = "0188f8e8-..." (UUID v7 format)

// Create a hash for storage
hash, err := apikey.HashKey(apiKey, secretKey)
if err != nil {
    // Handle error
    return fmt.Errorf("failed to hash key: %w", err)
}
// hash = "sha256:..." (storable hash string)
```

### Validating API Keys

```go
// When a client sends an API key, validate it against the stored hash
isValid := apikey.ValidateKey(apiKey, storedHash, secretKey)
if isValid {
    // Allow access
} else {
    // Deny access
}
```

### Error Handling

```go
apiKey, err := apikey.GenerateRandom()
if err != nil {
    switch {
    case errors.Is(err, apikey.ErrInsufficientEntropy):
        // Handle insufficient system entropy
        log.Fatal("System has insufficient entropy for secure key generation")
    default:
        // Handle other errors
        log.Fatalf("Failed to generate API key: %v", err)
    }
}
```

## Best Practices

1. **Security**:
   - Never store raw API keys - only store hashed versions
   - Use a strong secret key - your secret key should be long, random, and kept secure
   - Rotate API keys periodically for sensitive systems

2. **Performance**:
   - Cache validation results for frequently used keys
   - For high-traffic systems, consider implementing a key cache

3. **Error Handling**:
   - Always check for errors when generating keys
   - Implement appropriate logging for failed validation attempts

## API Reference

### Configuration Variables

```go
var DefaultKeyLength = 32 // Default length of random API keys in bytes
```

### Functions

```go
func GenerateRandom() (string, error)
```
Creates a new API key with a secure random value.

```go
func GenerateTimeOrdered() (string, error)
```
Creates a time-ordered API key using UUID V7 format.

```go
func HashKey(apiKey, secretKey string) (string, error)
```
Creates a secure hash of the API key for storage.

```go
func ValidateKey(apiKey, hash, secretKey string) bool
```
Checks if the API key matches the hash using the secret key.

### Error Types

```go
var ErrInsufficientEntropy = errors.New("insufficient system entropy")
var ErrInvalidKeyFormat = errors.New("invalid API key format")
```

Remember to adjust the README to match your package's complexity and specific needs. The best READMEs are comprehensive yet structured, with clear sections, code examples with comments showing expected outputs, and detailed API documentation including error types and configuration options.
