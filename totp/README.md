# TOTP Package

The `totp` package provides functionality for implementing Time-based One-Time Password (TOTP) authentication and recovery codes in Go applications. This package can be used to add two-factor authentication (2FA) to your applications, with secure encryption of TOTP secrets.

## Features

- Generate TOTP secret keys
- Create TOTP URIs for QR codes
- Validate TOTP codes
- Generate and hash recovery codes
- HMAC-based one-time password (HOTP) generation
- AES-256-GCM encryption for TOTP secrets
- Secure key management for encryption

## Configuration

The package requires an encryption key to be set in your environment:

```env
TOTP_ENCRYPTION_KEY=your_base64_encoded_32_byte_key
```

You can generate a new encryption key using:

```go
key, err := totp.GenerateEncryptionKey()
if err != nil {
    log.Fatal(err)
}
// Encode the key to base64 and store it securely
encodedKey := base64.StdEncoding.EncodeToString(key)
```

## Usage

### Generating a Secret Key

```go
secret, err := totp.GenerateSecretKey()
if err != nil {
    log.Fatal(err)
}

// Encrypt the secret before storing
key, err := totp.LoadEncryptionKey()
if err != nil {
    log.Fatal(err)
}

encryptedSecret, err := totp.EncryptSecret(secret, key)
if err != nil {
    log.Fatal(err)
}
// Store the encrypted secret securely
```

### Decrypting and Using a Stored Secret

```go
// Retrieve the encrypted secret and decrypt it
key, err := totp.LoadEncryptionKey()
if err != nil {
    log.Fatal(err)
}

secret, err := totp.DecryptSecret(encryptedSecret, key)
if err != nil {
    log.Fatal(err)
}
```

### Creating a TOTP URI

```go
// Create a URI that can be encoded as a QR code
uri := totp.GetTOTPURI(secret, "user@example.com", "YourApp")
// Use this URI with a QR code generator
```

### Validating TOTP Codes

```go
// When user provides a code
code := "123456" // Example code from user's authenticator app
valid, err := totp.ValidateTOTP(secret, code)
if err != nil {
    log.Fatal(err)
}
if valid {
    // Code is valid, proceed with authentication
} else {
    // Invalid code
}
```

### Generating Recovery Codes

```go
// Generate 8 recovery codes
codes, err := totp.GenerateRecoveryCodes(8)
if err != nil {
    log.Fatal(err)
}
// Store the hashed versions of these codes
for _, code := range codes {
    hashedCode := totp.HashRecoveryCode(code)
    // Store hashedCode in your database securely
}
// Provide the unhashed codes to the user
```

## Security Considerations

- Always store TOTP secrets in encrypted form using the provided encryption functions
- Keep your encryption key secure and separate from your application code
- Use HTTPS for all authentication requests
- Implement rate limiting for TOTP validation attempts
- Consider implementing backup methods for account recovery
- Regularly rotate encryption keys and re-encrypt secrets
- Monitor for suspicious authentication patterns
- Implement proper key management procedures for your encryption keys

## Best Practices

1. **Key Management**

    - Store encryption keys in a secure key management system
    - Implement key rotation procedures
    - Never store encryption keys in code or version control

2. **Secret Storage**

    - Always encrypt TOTP secrets before storage
    - Use secure database systems with encryption at rest
    - Implement proper access controls

3. **Authentication Flow**
    - Implement proper session management
    - Use rate limiting to prevent brute force attacks
    - Log authentication attempts for security monitoring
