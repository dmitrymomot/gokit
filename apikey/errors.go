package apikey

import "errors"

var (
	// ErrEmptyInput is returned when the API key or secret key is empty
	ErrEmptyInput = errors.New("empty api key or secret key")
	// ErrGeneration is returned when API key generation fails
	ErrGeneration = errors.New("failed to generate api key")
	// ErrInvalidHash is returned when the hash format is invalid
	ErrInvalidHash = errors.New("invalid hash format")
)
