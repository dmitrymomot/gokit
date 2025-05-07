package validator

import (
	"errors"
	"fmt"
	"net/url"
)

// ErrInvalidSeparatorConfiguration is returned when validator separator characters are not distinct.
var ErrInvalidSeparatorConfiguration = errors.New("rule, parameter, and parameter list separators must be distinct characters")

// ValidationError represents a validation error.
// It contains the field name and the error message.
type ValidationErrors url.Values

// Error returns the error message.
// Implements the error interface.
func (e ValidationErrors) Error() string {
	return fmt.Sprintf("%+v", url.Values(e))
}

// Values returns the underlying url.Values.
func (e ValidationErrors) Values() url.Values {
	return url.Values(e)
}

// defaultErrorTranslator returns the key as the default message.
func defaultErrorTranslator(key string, label string, params ...string) string {
	return key
}

// NewValidationError creates a new validation error.
// It returns a map with the field name as the key and the error message as the value.
func NewValidationError(args ...string) ValidationErrors {
	ve := make(ValidationErrors)
	maxPairs := len(args) / 2 * 2 // Ensure even number to process complete pairs only
	for i := 0; i < maxPairs; i += 2 {
		ve[args[i]] = []string{args[i+1]}
	}
	return ve
}

// ExtractValidationErrors extracts the validation errors from the error.
// It returns the validation errors if the error is a validation error.
// It returns nil otherwise.
func ExtractValidationErrors(err error) ValidationErrors {
	var ve ValidationErrors
	if errors.As(err, &ve) {
		return ve
	}
	return nil
}

// IsValidationError checks if the error is a validation error.
// It returns true if the error is a validation error.
// It returns false otherwise.
func IsValidationError(err error) bool {
	var ve ValidationErrors
	return errors.As(err, &ve) && len(ve) > 0
}
