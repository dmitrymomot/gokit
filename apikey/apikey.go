package apikey

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/agentstation/uuidkey"
	"github.com/google/uuid"
)

const (
	APIKeyLength = 32 // 256 bits
)

// GenerateRandom creates a new API key with a secure random value.
// Returns a hex-encoded string of 64 characters (32 bytes) or an error if generation fails.
func GenerateRandom() (string, error) {
	bytes := make([]byte, APIKeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", ErrGeneration
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateTimeOrdered creates a new API key using a UUID V7 and encodes it.
// It generates a time-ordered key that can be chronologically sorted.
// Returns the encoded API key as a string or an error if generation or encoding fails.
func GenerateTimeOrdered() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", errors.Join(ErrGeneration, err)
	}
	apiKey, err := uuidkey.Encode(id.String())
	if err != nil {
		return "", errors.Join(ErrGeneration, err)
	}
	return apiKey.String(), nil
}

// HashKey hashes the API key using HMAC-SHA256 with a secret key.
// Both apiKey and secretKey must be non-empty strings.
// Returns the hex-encoded hash string or an error if inputs are invalid.
func HashKey(apiKey, secretKey string) (string, error) {
	if apiKey == "" || secretKey == "" {
		return "", ErrEmptyInput
	}
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(apiKey))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SecureCompare performs a constant-time comparison of two strings to prevent timing attacks.
// Returns true if the strings are equal, false otherwise.
func SecureCompare(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return hmac.Equal([]byte(a), []byte(b))
}

// ValidateKey checks if the API key matches the hash using the secret key.
// Returns true if the API key matches the hash, false otherwise.
func ValidateKey(apiKey, hash, secretKey string) bool {
	if apiKey == "" || hash == "" || secretKey == "" {
		return false
	}
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(apiKey))
	expected := h.Sum(nil)
	actual, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, actual)
}
