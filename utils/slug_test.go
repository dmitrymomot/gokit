package utils_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic conversion",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "special characters",
			input:    "Hello@#$World!!!",
			expected: "hello-world",
		},
		{
			name:     "multiple spaces",
			input:    "Hello    World",
			expected: "hello-world",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
		{
			name:     "multiple hyphens",
			input:    "hello---world",
			expected: "hello-world",
		},
		{
			name:     "mixed case",
			input:    "HeLLo WoRLD",
			expected: "hello-world",
		},
		{
			name:     "numbers",
			input:    "Hello World 123",
			expected: "hello-world-123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "@#$%^&*",
			expected: "",
		},
		{
			name:     "unicode characters",
			input:    "Hello Värld",
			expected: "hello-varld",
		},
		{
			name:     "long text with various characters",
			input:    "This is a Long Text with Various Characters!!! 123",
			expected: "this-is-a-long-text-with-various-characters-123",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "URL-like string",
			input:    "http://example.com/page",
			expected: "http-example-com-page",
		},
		{
			name:     "email-like string",
			input:    "user@example.com",
			expected: "user-example-com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GenerateSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateSlug_ConsistencyCheck(t *testing.T) {
	// Test that running GenerateSlug multiple times on the same input produces the same result
	input := "Hello World!"
	firstResult := utils.GenerateSlug(input)
	secondResult := utils.GenerateSlug(firstResult)

	assert.Equal(t, firstResult, secondResult, "GenerateSlug should be idempotent")
}

func TestDeprecatedToSlug(t *testing.T) {
	// Test that the deprecated function still works correctly
	t.Run("calls GenerateSlug", func(t *testing.T) {
		input := "Hello World!"
		oldResult := utils.ToSlug(input)
		newResult := utils.GenerateSlug(input)
		
		assert.Equal(t, newResult, oldResult, "Deprecated ToSlug should return same result as GenerateSlug")
	})
}

func BenchmarkGenerateSlug(b *testing.B) {
	input := "This is a Long Text with Various Characters!!! 123"
	for i := 0; i < b.N; i++ {
		utils.GenerateSlug(input)
	}
}

func BenchmarkDeprecatedToSlug(b *testing.B) {
	input := "This is a Long Text with Various Characters!!! 123"
	for i := 0; i < b.N; i++ {
		utils.ToSlug(input)
	}
}
