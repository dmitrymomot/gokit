package privacy

import "context"

// Masker is the interface that all data masking implementations must implement.
type Masker interface {
	// Mask applies the masking strategy to the given data.
	// It returns the masked data or an error if masking fails.
	Mask(ctx context.Context, data any) (any, error)

	// CanMask checks if the masker can handle the given data type.
	// Returns true if the data can be masked, false otherwise.
	CanMask(data any) bool
}

// Option is a functional option for configuring maskers.
type Option func(m any) error

// MaskingStrategy is an identifier for different masking strategies.
type MaskingStrategy string

// Predefined masking strategies.
const (
	// StrategyRedact completely redacts the value, replacing it with a fixed string.
	StrategyRedact MaskingStrategy = "redact"

	// StrategyPartialMask keeps parts of the data visible, masking the rest.
	StrategyPartialMask MaskingStrategy = "partial"

	// StrategyTokenize replaces the original value with a token that can be
	// mapped back to the original value in a secure environment.
	StrategyTokenize MaskingStrategy = "tokenize"

	// StrategyPseudonymize replaces the original value with a pseudonym
	// that is consistent for the same input value.
	StrategyPseudonymize MaskingStrategy = "pseudonymize"

	// StrategyEncrypt encrypts the data with a cryptographic algorithm.
	StrategyEncrypt MaskingStrategy = "encrypt"

	// StrategyNoise adds statistical noise to numerical data.
	StrategyNoise MaskingStrategy = "noise"
)

// DataCategory represents the category of data for classification and policy enforcement.
type DataCategory string

// Predefined data categories.
const (
	// CategoryPII represents Personally Identifiable Information.
	CategoryPII DataCategory = "pii"

	// CategoryFinancial represents financial information like credit card numbers.
	CategoryFinancial DataCategory = "financial"

	// CategoryHealth represents health-related information (e.g., HIPAA in the US).
	CategoryHealth DataCategory = "health"

	// CategoryCredentials represents authentication credentials.
	CategoryCredentials DataCategory = "credentials"

	// CategoryLocation represents geolocation data.
	CategoryLocation DataCategory = "location"

	// CategoryCommunication represents communication data like emails, phone numbers.
	CategoryCommunication DataCategory = "communication"

	// Specific data categories
	
	// CategoryCreditCard represents credit card information.
	CategoryCreditCard DataCategory = "credit_card"
	
	// CategoryEmail represents email addresses.
	CategoryEmail DataCategory = "email"
	
	// CategoryPhone represents phone numbers.
	CategoryPhone DataCategory = "phone"
	
	// CategorySSN represents Social Security Numbers.
	CategorySSN DataCategory = "ssn"
	
	// CategoryName represents personal names.
	CategoryName DataCategory = "name"
	
	// CategoryAddress represents physical addresses.
	CategoryAddress DataCategory = "address"
	
	// CategoryPassport represents passport numbers.
	CategoryPassport DataCategory = "passport"
	
	// CategoryDriverLicense represents driver's license numbers.
	CategoryDriverLicense DataCategory = "driver_license"
)

// MaskerRegistry manages a collection of maskers for different data types and categories.
type MaskerRegistry interface {
	// RegisterMasker adds a masker for a specific data category.
	RegisterMasker(category DataCategory, masker Masker) error

	// GetMasker retrieves a masker for a specific data category.
	GetMasker(category DataCategory) (Masker, error)

	// MaskByCategory masks data according to its category.
	MaskByCategory(ctx context.Context, category DataCategory, data any) (any, error)
}
