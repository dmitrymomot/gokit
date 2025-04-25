package privacy

import "errors"

// Predefined errors for the privacy package.
var (
	// ErrInvalidData indicates that the provided data is invalid for masking.
	ErrInvalidData = errors.New("privacy: invalid data for masking")

	// ErrInvalidMask indicates that the masking configuration is invalid.
	ErrInvalidMask = errors.New("privacy: invalid mask configuration")

	// ErrUnsupportedType indicates the provided data type is not supported by the masker.
	ErrUnsupportedType = errors.New("privacy: unsupported data type")

	// ErrInvalidRegex indicates a regex pattern is invalid.
	ErrInvalidRegex = errors.New("privacy: invalid regex pattern")

	// ErrMaskingFailed indicates a general failure during the masking process.
	ErrMaskingFailed = errors.New("privacy: masking operation failed")

	// ErrMaskerNotFound indicates that no masker exists for the requested category.
	ErrMaskerNotFound = errors.New("privacy: masker not found for category")

	// ErrInvalidMasker indicates that the provided masker is invalid or nil.
	ErrInvalidMasker = errors.New("privacy: invalid masker")

	// ErrInvalidRule indicates that the provided detection rule is invalid or nil.
	ErrInvalidRule = errors.New("privacy: invalid detection rule")

	// ErrRuleNotFound indicates that no detection rule exists for the requested category.
	ErrRuleNotFound = errors.New("privacy: detection rule not found")

	// ErrCategoryNotDetected indicates that the data category could not be automatically detected.
	ErrCategoryNotDetected = errors.New("privacy: data category could not be detected")

	// Character count errors

	// ErrNegativeCharCount indicates that a visible character count cannot be negative.
	ErrNegativeCharCount = errors.New("privacy: visible character count cannot be negative")

	// ErrNegativeDigitCount indicates that a visible digit count cannot be negative.
	ErrNegativeDigitCount = errors.New("privacy: visible digit count cannot be negative")

	// ErrTooManyVisibleDigits indicates that showing too many digits is a security risk.
	ErrTooManyVisibleDigits = errors.New("privacy: showing more than 4 digits is not allowed for security reasons")

	// ErrNegativeMinLength indicates that a minimum length cannot be negative.
	ErrNegativeMinLength = errors.New("privacy: minimum length cannot be negative")

	// ErrPartialMaskRequiresVisibleChars indicates that partial masking requires visible characters.
	ErrPartialMaskRequiresVisibleChars = errors.New("privacy: partial masking requires at least some visible characters")

	// Format errors

	// ErrInvalidEmailFormat indicates that the email format is invalid.
	ErrInvalidEmailFormat = errors.New("privacy: invalid email format")

	// ErrInvalidCardNumberLength indicates that a credit card number length is invalid.
	ErrInvalidCardNumberLength = errors.New("privacy: invalid credit card number length")

	// Registry errors

	// ErrNilMasker indicates that a masker cannot be nil.
	ErrNilMasker = errors.New("privacy: masker cannot be nil")

	// ErrNilDetectionRule indicates that a detection rule cannot be nil.
	ErrNilDetectionRule = errors.New("privacy: detection rule cannot be nil")
)
