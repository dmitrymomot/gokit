package privacy

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// StringMasker implements the Masker interface for string data.
type StringMasker struct {
	// replacement is the character used to mask the original characters.
	replacement rune

	// strategy defines how the string should be masked.
	strategy MaskingStrategy

	// visibleStartChars is the number of characters from the start to keep visible.
	visibleStartChars int

	// visibleEndChars is the number of characters from the end to keep visible.
	visibleEndChars int

	// minLength is the minimum length of string that can be masked.
	minLength int
}

// StringMaskerOption is a function that configures a StringMasker.
type StringMaskerOption func(*StringMasker) error

// WithReplacement sets the character used for masking.
func WithReplacement(char rune) StringMaskerOption {
	return func(sm *StringMasker) error {
		sm.replacement = char
		return nil
	}
}

// WithStrategy sets the masking strategy.
func WithStrategy(strategy MaskingStrategy) StringMaskerOption {
	return func(sm *StringMasker) error {
		sm.strategy = strategy
		return nil
	}
}

// WithVisibleChars sets the number of characters to keep visible from
// the start and end of the string.
func WithVisibleChars(start, end int) StringMaskerOption {
	return func(sm *StringMasker) error {
		if start < 0 || end < 0 {
			return errors.Join(ErrInvalidMask, ErrNegativeCharCount)
		}
		sm.visibleStartChars = start
		sm.visibleEndChars = end
		return nil
	}
}

// WithMinLength sets the minimum length of string that can be masked.
func WithMinLength(length int) StringMaskerOption {
	return func(sm *StringMasker) error {
		if length < 0 {
			return errors.Join(ErrInvalidMask, ErrNegativeMinLength)
		}
		sm.minLength = length
		return nil
	}
}

// NewStringMasker creates a new StringMasker with the given options.
func NewStringMasker(opts ...StringMaskerOption) (*StringMasker, error) {
	// Default configuration
	sm := &StringMasker{
		replacement:       '*',
		strategy:          StrategyPartialMask,
		visibleStartChars: 0,
		visibleEndChars:   0,
		minLength:         1,
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(sm); err != nil {
			return nil, err
		}
	}

	// Validate configuration
	if sm.strategy == StrategyPartialMask && sm.visibleStartChars+sm.visibleEndChars <= 0 {
		return nil, errors.Join(ErrInvalidMask, ErrPartialMaskRequiresVisibleChars)
	}

	return sm, nil
}

// CanMask checks if the masker can mask the given data.
func (sm *StringMasker) CanMask(data any) bool {
	_, ok := data.(string)
	return ok
}

// Mask applies the masking strategy to the given string.
func (sm *StringMasker) Mask(ctx context.Context, data any) (any, error) {
	// Check if context is canceled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	str, ok := data.(string)
	if !ok {
		return nil, ErrUnsupportedType
	}

	// Check if string meets minimum length requirement
	if utf8.RuneCountInString(str) < sm.minLength {
		return str, nil
	}

	// Apply masking based on strategy
	switch sm.strategy {
	case StrategyRedact:
		return "[REDACTED]", nil

	case StrategyPartialMask:
		return sm.partialMask(str), nil

	default:
		return nil, errors.Join(ErrInvalidMask,
			fmt.Errorf("unsupported strategy: %s", sm.strategy))
	}
}

// partialMask implements the partial masking strategy.
func (sm *StringMasker) partialMask(s string) string {
	runes := []rune(s)
	length := len(runes)

	// Handle special case where desired visible characters exceed string length
	if sm.visibleStartChars+sm.visibleEndChars >= length {
		// If the string is too short, keep all characters visible
		return s
	}

	// Calculate how many characters to mask
	maskLen := length - sm.visibleStartChars - sm.visibleEndChars
	if maskLen <= 0 {
		return s
	}

	// Create the masked string
	var result strings.Builder
	result.Grow(length)

	// Copy visible prefix
	for i := 0; i < sm.visibleStartChars; i++ {
		result.WriteRune(runes[i])
	}

	// Add mask characters
	for i := 0; i < maskLen; i++ {
		result.WriteRune(sm.replacement)
	}

	// Copy visible suffix
	for i := length - sm.visibleEndChars; i < length; i++ {
		result.WriteRune(runes[i])
	}

	return result.String()
}
