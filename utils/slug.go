package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateSlug converts a string to a URL-friendly slug.
// It removes diacritics, replaces all non-alphanumeric characters with hyphens,
// and converts the string to lowercase.
//
// Example:
//
//	GenerateSlug("Hello World") // returns "hello-world"
//	GenerateSlug("Café & Restaurant") // returns "cafe-restaurant"
func GenerateSlug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Create a transformer that removes diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	// Apply the transformation
	result, _, _ := transform.String(t, s)

	// Replace non-alphanumeric characters with a hyphen
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	result = reg.ReplaceAllString(result, "-")

	// Remove leading and trailing hyphens
	result = strings.Trim(result, "-")

	return result
}

// Deprecated: ToSlug is deprecated, use GenerateSlug instead.
func ToSlug(s string) string {
	return GenerateSlug(s)
}
