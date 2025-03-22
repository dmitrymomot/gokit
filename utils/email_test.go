package utils_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/utils"
)

func TestNormalizeDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "simple email",
			email:    "john.doe@example.com",
			expected: "John Doe",
		},
		{
			name:     "email with dots and underscores",
			email:    "john_doe.smith@example.com",
			expected: "John Doe Smith",
		},
		{
			name:     "email with numbers",
			email:    "user123@example.com",
			expected: "User",
		},
		{
			name:     "email with hyphens",
			email:    "john-doe-smith@example.com",
			expected: "John Doe Smith",
		},
		{
			name:     "email with mixed case",
			email:    "JohnDoe@example.com",
			expected: "Johndoe",
		},
		{
			name:     "email with special characters",
			email:    "john.doe+test@example.com",
			expected: "John Doe Test",
		},
		{
			name:     "email with multiple spaces after normalization",
			email:    "john...doe___smith@example.com",
			expected: "John Doe Smith",
		},
		{
			name:     "email without @ symbol",
			email:    "johndoe",
			expected: "Johndoe",
		},
		{
			name:     "empty email",
			email:    "",
			expected: "",
		},
		{
			name:     "email with numbers in middle",
			email:    "user123name@example.com",
			expected: "User123name",
		},
		{
			name:     "email with numbers at start",
			email:    "123user@example.com",
			expected: "User",
		},
		{
			name:     "email with numbers at end",
			email:    "user123@example.com",
			expected: "User",
		},
		{
			name:     "email with numbers at start and end",
			email:    "123user456@example.com",
			expected: "User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.NormalizeDisplayName(tt.email)
			if result != tt.expected {
				t.Errorf("NormalizeDisplayName(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}

// Test to ensure the deprecated function still works correctly
func TestGetNormalizedDisplayName(t *testing.T) {
	testCases := []struct {
		email    string
		expected string
	}{
		{"john.doe@example.com", "John Doe"},
		{"user123@example.com", "User"},
	}

	for _, tc := range testCases {
		result := utils.GetNormalizedDisplayName(tc.email)
		if result != tc.expected {
			t.Errorf("GetNormalizedDisplayName(%q) = %q, want %q", tc.email, result, tc.expected)
		}
	}
}
