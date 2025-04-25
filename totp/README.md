# TOTP Package

A secure Time-based One-Time Password (TOTP) implementation with encryption.

## Installation

```bash
go get github.com/dmitrymomot/gokit/totp
```

## Overview

The `totp` package provides a complete solution for implementing secure two-factor authentication (2FA) using Time-based One-Time Passwords (TOTP). It offers functionality for generating and validating TOTP codes, managing secrets with AES-256-GCM encryption, and handling recovery codes.

## Features

- TOTP generation and validation compliant with RFC 6238
- AES-256-GCM encryption for secure storage of TOTP secrets
- Recovery code generation and validation
- QR code URI generation for easy mobile app setup
- HMAC-based one-time password (HOTP) support
- Comprehensive error handling

## Usage

### Setup and Secret Generation

```go
import (
	"encoding/base64"
	"github.com/dmitrymomot/gokit/totp"
)

// Generate a new encryption key (store this securely)
key, err := totp.GenerateEncryptionKey()
if err != nil {
	// Handle error
}
encodedKey := base64.StdEncoding.EncodeToString(key)
// Save this key securely (environment variable, key vault, etc.)

// Generate a new TOTP secret for a user
secret, err := totp.GenerateSecretKey()
if err != nil {
	// Handle error
}

// Encrypt the secret before storage
encryptedSecret, err := totp.EncryptSecret(secret, key)
if err != nil {
	// Handle error
}
// Store encryptedSecret in your database
```

### Generating QR Code URI

```go
// Generate a URI for QR code display
// Parameters: secret, user identifier, issuer name
uri := totp.GetTOTPURI(secret, "user@example.com", "MyApp")

// Use this URI with a QR code generator library
// Example with github.com/skip2/go-qrcode:
// qrCode, _ := qrcode.New(uri, qrcode.Medium)
// png, _ := qrCode.PNG(256)
```

### Validating TOTP Codes

```go
// Retrieve encrypted secret from storage and decrypt it
key, err := totp.LoadEncryptionKey() // Load from environment
if err != nil {
	// Handle error
}

secret, err := totp.DecryptSecret(encryptedSecret, key)
if err != nil {
	// Handle error
}

// When user submits a TOTP code for verification
userProvidedCode := "123456" // From user input

valid, err := totp.ValidateTOTP(secret, userProvidedCode)
if err != nil {
	// Handle error
}

if valid {
	// Authentication successful, grant access
} else {
	// Authentication failed
}
```

### Recovery Codes

```go
// Generate a set of recovery codes (typically done during 2FA setup)
recoveryCodes, err := totp.GenerateRecoveryCodes(8) // Generate 8 codes
if err != nil {
	// Handle error
}

// Hash and store codes in database
var hashedCodes []string
for _, code := range recoveryCodes {
	hashedCode := totp.HashRecoveryCode(code)
	hashedCodes = append(hashedCodes, hashedCode)
	// Store hashedCode in database
}

// Provide the original unhashed codes to user for backup

// Validating a recovery code during account recovery
userProvidedCode := "ABCD-1234-EFGH" // From user input
validCode := false

// Check against all stored hashed codes
for _, hashedCode := range hashedCodes {
	if totp.ValidateRecoveryCode(userProvidedCode, hashedCode) {
		validCode = true
		// Remove this code from database after use
		break
	}
}

if validCode {
	// Allow access and reset 2FA
} else {
	// Invalid recovery code
}
```

### Custom Configuration

```go
// Use custom TOTP parameters (default is 6 digits, 30 seconds)
valid, err := totp.ValidateTOTPWithParams(
	secret,
	userCode,
	8,        // 8 digits instead of 6
	60,       // 60 second period instead of 30
	1,        // Allow 1 period skew before/after
)

// Generate custom TOTP URI
customURI := totp.GetTOTPURIWithParams(
	secret,
	"user@example.com",
	"MyApp",
	8,        // 8 digits
	60,       // 60 second period
)
```

## API Reference

### Core Functions

```go
// Generate a new TOTP secret key
func GenerateSecretKey() (string, error)

// Create a TOTP URI for QR code generation
func GetTOTPURI(secret, accountName, issuer string) string

// Create a TOTP URI with custom parameters
func GetTOTPURIWithParams(secret, accountName, issuer string, digits, period int) string

// Validate a TOTP code against a secret
func ValidateTOTP(secret, code string) (bool, error)

// Validate with custom parameters
func ValidateTOTPWithParams(secret, code string, digits, period, skew int) (bool, error)
```

### Encryption Functions

```go
// Generate a new 32-byte encryption key
func GenerateEncryptionKey() ([]byte, error)

// Load encryption key from environment variable
func LoadEncryptionKey() ([]byte, error)

// Encrypt a TOTP secret
func EncryptSecret(secret string, key []byte) (string, error)

// Decrypt a TOTP secret
func DecryptSecret(encryptedSecret string, key []byte) (string, error)
```

### Recovery Code Functions

```go
// Generate recovery codes
func GenerateRecoveryCodes(count int) ([]string, error)

// Hash a recovery code for storage
func HashRecoveryCode(code string) string

// Validate a recovery code against its hash
func ValidateRecoveryCode(code, hash string) bool
```

### Configuration

Set the following environment variable for encryption key management:

```
TOTP_ENCRYPTION_KEY=your_base64_encoded_32_byte_key
```

## Security Best Practices

1. **Secret Management**
   - Always encrypt TOTP secrets before storage
   - Use a secure key management system for encryption keys
   - Never store encryption keys in code or version control
   - Implement key rotation procedures

2. **Authentication Security**
   - Implement rate limiting for TOTP verification attempts
   - Use HTTPS for all authentication endpoints
   - Log authentication attempts for security monitoring
   - Set appropriate session timeouts after authentication

3. **Recovery Code Security**
   - Store only hashed recovery codes
   - Invalidate recovery codes after use
   - Generate new recovery codes whenever TOTP is reset
   - Notify users when recovery codes are used

4. **Implementation Details**
   - Use a secure random number generator for all crypto operations
   - Follow the RFC 6238 specification for TOTP implementation
   - Regularly update dependencies and security libraries
   - Perform security audits of your authentication flow
