package sanitizer_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/sanitizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SpecialTestStruct struct {
	SlugField string `sanitize:"slug"`
	UUIDField string `sanitize:"uuid"`
	BoolField string `sanitize:"bool"`
}

func TestSpecialSanitizers(t *testing.T) {
	tests := []struct {
		name     string
		input    *SpecialTestStruct
		expected *SpecialTestStruct
	}{
	// Test slug sanitizer
	{
		name: "Slug with spaces and special chars",
		input: &SpecialTestStruct{
			SlugField: "Hello World! This is a Test 123",
			BoolField: "false", // Explicitly set to avoid empty string being converted
		},
		expected: &SpecialTestStruct{
			SlugField: "hello-world-this-is-a-test-123",
			BoolField: "false",
		},
	},
	{
		name: "Slug with diacritics",
		input: &SpecialTestStruct{
			SlugField: "Café au Lait",
			BoolField: "false", // Explicitly set to avoid empty string being converted
		},
		expected: &SpecialTestStruct{
			SlugField: "cafe-au-lait",
			BoolField: "false",
		},
	},

	// Test UUID sanitizer
	{
		name: "UUID with braces",
		input: &SpecialTestStruct{
			UUIDField: "{550E8400-E29B-41D4-A716-446655440000}",
			BoolField: "false", // Explicitly set to avoid empty string being converted
		},
		expected: &SpecialTestStruct{
			UUIDField: "550e8400-e29b-41d4-a716-446655440000",
			BoolField: "false",
		},
	},
	{
		name: "UUID without braces",
		input: &SpecialTestStruct{
			UUIDField: "550E8400-E29B-41D4-A716-446655440000",
			BoolField: "false", // Explicitly set to avoid empty string being converted
		},
		expected: &SpecialTestStruct{
			UUIDField: "550e8400-e29b-41d4-a716-446655440000",
			BoolField: "false",
		},
	},
	{
		name: "Invalid UUID",
		input: &SpecialTestStruct{
			UUIDField: "not-a-uuid",
			BoolField: "false", // Explicitly set to avoid empty string being converted
		},
		expected: &SpecialTestStruct{
			UUIDField: "not-a-uuid", // Should remain unchanged
			BoolField: "false",
		},
	},

	// Test bool sanitizer with strings
	{
		name: "Bool from string true",
		input: &SpecialTestStruct{
			BoolField: "true",
		},
		expected: &SpecialTestStruct{
			BoolField: "true",
		},
	},
	{
		name: "Bool from string yes",
		input: &SpecialTestStruct{
			BoolField: "yes",
		},
		expected: &SpecialTestStruct{
			BoolField: "true",
		},
	},
	{
		name: "Bool from string 1",
		input: &SpecialTestStruct{
			BoolField: "1",
		},
		expected: &SpecialTestStruct{
			BoolField: "true",
		},
	},
	{
		name: "Bool from string false",
		input: &SpecialTestStruct{
			BoolField: "false",
		},
		expected: &SpecialTestStruct{
			BoolField: "false",
		},
	},
	{
		name: "Bool from string no",
		input: &SpecialTestStruct{
			BoolField: "no",
		},
		expected: &SpecialTestStruct{
			BoolField: "false",
		},
	},
	{
		name: "Bool from string 0",
		input: &SpecialTestStruct{
			BoolField: "0",
		},
		expected: &SpecialTestStruct{
			BoolField: "false",
		},
	},
	{
		name: "Bool from invalid string",
		input: &SpecialTestStruct{
			BoolField: "maybe",
		},
		expected: &SpecialTestStruct{
			BoolField: "maybe", // Should remain unchanged
		},
	},

	// Test with empty values
	{
		name: "Empty values",
		input: &SpecialTestStruct{
			SlugField: "",
			UUIDField: "",
			BoolField: "",
		},
		expected: &SpecialTestStruct{
			SlugField: "",
			UUIDField: "",
			BoolField: "false", // Empty string converts to "false"
		},
	},
	}

	s, err := sanitizer.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SanitizeStruct(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestSpecialSanitizersWithMultipleRules(t *testing.T) {
	type TestStruct struct {
		SlugField string `sanitize:"trim;lower;slug"`
		UUIDField string `sanitize:"trim;uuid"`
	}

	tests := []struct {
		name     string
		input    *TestStruct
		expected *TestStruct
	}{
	{
		name: "Multiple rules with special sanitizers",
		input: &TestStruct{
			SlugField: "  Hello World! This is a Test  ",
			UUIDField: "  {550E8400-E29B-41D4-A716-446655440000}  ",
		},
		expected: &TestStruct{
			SlugField: "hello-world-this-is-a-test",
			UUIDField: "550e8400-e29b-41d4-a716-446655440000",
		},
	},
	}

	s, err := sanitizer.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SanitizeStruct(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}
