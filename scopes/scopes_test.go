package scopes_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/scopes"
	"github.com/stretchr/testify/assert"
)

func TestParseScopes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single scope",
			input:    "read",
			expected: []string{"read"},
		},
		{
			name:     "multiple scopes",
			input:    "read write delete",
			expected: []string{"read", "write", "delete"},
		},
		{
			name:     "extra spaces",
			input:    "  read   write  ",
			expected: []string{"read", "write"},
		},
		{
			name:     "hierarchical scopes",
			input:    "admin.read user.write system.*",
			expected: []string{"admin.read", "user.write", "system.*"},
		},
		{
			name:     "mixed scopes with wildcards",
			input:    "* admin.read user.*",
			expected: []string{"*", "admin.read", "user.*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.ParseScopes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		expected string
	}{
		{
			name:     "empty scopes",
			scopes:   []string{},
			expected: "",
		},
		{
			name:     "nil scopes",
			scopes:   nil,
			expected: "",
		},
		{
			name:     "single scope",
			scopes:   []string{"read"},
			expected: "read",
		},
		{
			name:     "multiple scopes",
			scopes:   []string{"read", "write"},
			expected: "read write",
		},
		{
			name:     "hierarchical scopes",
			scopes:   []string{"admin.read", "user.write", "system.*"},
			expected: "admin.read user.write system.*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.JoinScopes(tt.scopes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsScope(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		scope    string
		expected bool
	}{
		{
			name:     "empty scopes",
			scopes:   []string{},
			scope:    "read",
			expected: false,
		},
		{
			name:     "exact match",
			scopes:   []string{"read", "write"},
			scope:    "read",
			expected: true,
		},
		{
			name:     "no match",
			scopes:   []string{"read", "write"},
			scope:    "delete",
			expected: false,
		},
		{
			name:     "global wildcard",
			scopes:   []string{"*"},
			scope:    "anything",
			expected: true,
		},
		{
			name:     "namespace wildcard match",
			scopes:   []string{"admin.*"},
			scope:    "admin.read",
			expected: true,
		},
		{
			name:     "namespace wildcard no match",
			scopes:   []string{"admin.*"},
			scope:    "user.read",
			expected: false,
		},
		{
			name:     "deep scope match",
			scopes:   []string{"admin.users.*"},
			scope:    "admin.users.read",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.ContainsScope(tt.scopes, tt.scope)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasAllScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		required []string
		expected bool
	}{
		{
			name:     "empty required",
			scopes:   []string{"read"},
			required: []string{},
			expected: true,
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			required: []string{"read"},
			expected: false,
		},
		{
			name:     "has all required",
			scopes:   []string{"read", "write"},
			required: []string{"read"},
			expected: true,
		},
		{
			name:     "missing required",
			scopes:   []string{"read"},
			required: []string{"write"},
			expected: false,
		},
		{
			name:     "global wildcard",
			scopes:   []string{"*"},
			required: []string{"read", "write", "admin.users"},
			expected: true,
		},
		{
			name:     "namespace wildcard",
			scopes:   []string{"admin.*"},
			required: []string{"admin.read", "admin.write"},
			expected: true,
		},
		{
			name:     "mixed wildcards",
			scopes:   []string{"admin.*", "user.read"},
			required: []string{"admin.write", "user.read"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.HasAllScopes(tt.scopes, tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasAnyScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		required []string
		expected bool
	}{
		{
			name:     "empty required",
			scopes:   []string{"read"},
			required: []string{},
			expected: true,
		},
		{
			name:     "empty scopes",
			scopes:   []string{},
			required: []string{"read"},
			expected: false,
		},
		{
			name:     "has one required",
			scopes:   []string{"read", "write"},
			required: []string{"write", "delete"},
			expected: true,
		},
		{
			name:     "has none required",
			scopes:   []string{"read"},
			required: []string{"write", "delete"},
			expected: false,
		},
		{
			name:     "global wildcard",
			scopes:   []string{"*"},
			required: []string{"anything", "whatever"},
			expected: true,
		},
		{
			name:     "namespace wildcard match",
			scopes:   []string{"admin.*"},
			required: []string{"admin.read", "user.write"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.HasAnyScopes(tt.scopes, tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEqualScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes1  []string
		scopes2  []string
		expected bool
	}{
		{
			name:     "empty scopes",
			scopes1:  []string{},
			scopes2:  []string{},
			expected: true,
		},
		{
			name:     "nil scopes",
			scopes1:  nil,
			scopes2:  nil,
			expected: true,
		},
		{
			name:     "same scopes different order",
			scopes1:  []string{"read", "write"},
			scopes2:  []string{"write", "read"},
			expected: true,
		},
		{
			name:     "different scopes",
			scopes1:  []string{"read"},
			scopes2:  []string{"write"},
			expected: false,
		},
		{
			name:     "different lengths",
			scopes1:  []string{"read", "write"},
			scopes2:  []string{"read"},
			expected: false,
		},
		{
			name:     "with wildcards",
			scopes1:  []string{"admin.*", "user.read"},
			scopes2:  []string{"user.read", "admin.*"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.EqualScopes(tt.scopes1, tt.scopes2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateScopes(t *testing.T) {
	tests := []struct {
		name        string
		scopes      []string
		validScopes []string
		expected    bool
	}{
		{
			name:        "empty scopes",
			scopes:      []string{},
			validScopes: []string{"read", "write"},
			expected:    true,
		},
		{
			name:        "empty valid scopes",
			scopes:      []string{"read"},
			validScopes: []string{},
			expected:    false,
		},
		{
			name:        "all valid",
			scopes:      []string{"read", "write"},
			validScopes: []string{"read", "write", "delete"},
			expected:    true,
		},
		{
			name:        "invalid scope",
			scopes:      []string{"read", "invalid"},
			validScopes: []string{"read", "write"},
			expected:    false,
		},
		{
			name:        "wildcard in valid scopes",
			scopes:      []string{"custom.scope", "another.scope"},
			validScopes: []string{"*"},
			expected:    true,
		},
		{
			name:        "namespace wildcard in valid scopes",
			scopes:      []string{"admin.read", "admin.write"},
			validScopes: []string{"admin.*"},
			expected:    true,
		},
		{
			name:        "mixed wildcards and explicit scopes",
			scopes:      []string{"admin.read", "user.write"},
			validScopes: []string{"admin.*", "user.write"},
			expected:    true,
		},
		{
			name:        "invalid with wildcards",
			scopes:      []string{"admin.read", "user.write"},
			validScopes: []string{"admin.*", "system.*"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.ValidateScopes(tt.scopes, tt.validScopes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original string
	}{
		{
			name:     "simple scopes",
			original: "read write delete",
		},
		{
			name:     "hierarchical scopes",
			original: "admin.read user.write system.*",
		},
		{
			name:     "mixed wildcards",
			original: "* admin.* user.read.write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopeSlice := scopes.ParseScopes(tt.original)
			result := scopes.JoinScopes(scopeSlice)
			assert.Equal(t, tt.original, result)
		})
	}
}

func TestNormalizeScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   []string
		expected []string
	}{
		{
			name:     "empty scopes",
			scopes:   []string{},
			expected: nil,
		},
		{
			name:     "nil scopes",
			scopes:   nil,
			expected: nil,
		},
		{
			name:     "no duplicates",
			scopes:   []string{"read", "write", "delete"},
			expected: []string{"delete", "read", "write"},
		},
		{
			name:     "with duplicates",
			scopes:   []string{"read", "write", "read", "delete", "write"},
			expected: []string{"delete", "read", "write"},
		},
		{
			name:     "already sorted",
			scopes:   []string{"admin", "delete", "read", "write"},
			expected: []string{"admin", "delete", "read", "write"},
		},
		{
			name:     "with wildcards",
			scopes:   []string{"user.*", "admin.*", "*", "admin.*"},
			expected: []string{"*", "admin.*", "user.*"},
		},
		{
			name:     "with hierarchical scopes",
			scopes:   []string{"admin.write", "user.read", "admin.read", "user.read"},
			expected: []string{"admin.read", "admin.write", "user.read"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scopes.NormalizeScopes(tt.scopes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomDelimitersWithGlobalVars(t *testing.T) {
	// Save original values to restore after test
	originalSeparator := scopes.ScopeSeparator
	originalWildcard := scopes.ScopeWildcard
	originalDelimiter := scopes.ScopeDelimiter

	// Restore original values after test
	defer func() {
		scopes.ScopeSeparator = originalSeparator
		scopes.ScopeWildcard = originalWildcard
		scopes.ScopeDelimiter = originalDelimiter
	}()

	t.Run("custom separator", func(t *testing.T) {
		// Set custom separator
		scopes.ScopeSeparator = ","

		// Test parsing with comma separator
		parsed := scopes.ParseScopes("read,write,admin")
		assert.Equal(t, []string{"read", "write", "admin"}, parsed)

		// Test joining with comma separator
		joined := scopes.JoinScopes([]string{"read", "write", "admin"})
		assert.Equal(t, "read,write,admin", joined)

		// Test round trip
		original := "read,write,admin.users"
		roundTrip := scopes.JoinScopes(scopes.ParseScopes(original))
		assert.Equal(t, original, roundTrip)
	})

	t.Run("custom delimiter and wildcard", func(t *testing.T) {
		// Set custom delimiter and wildcard
		scopes.ScopeDelimiter = ":"
		scopes.ScopeWildcard = "?"

		// Test wildcard matching
		hasScope := scopes.ContainsScope([]string{"admin:?"}, "admin:read")
		assert.True(t, hasScope)

		// Test hierarchical scopes with custom delimiter
		hasAll := scopes.HasAllScopes(
			[]string{"admin:?", "read"}, // Added comma here
			[]string{"admin:users", "read"},
		)
		assert.True(t, hasAll)

		// Test validation with custom delimiter and wildcard
		isValid := scopes.ValidateScopes(
			[]string{"admin:read", "user:write"},
			[]string{"admin:?", "user:?"},
		)
		assert.True(t, isValid)
	})
}

func TestRoundTripWithCustomSeparator(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		separator string
	}{
		{
			name:      "comma separator",
			original:  "read,write,delete",
			separator: ",",
		},
		{
			name:      "semicolon separator",
			original:  "admin:read;user:write;system:*",
			separator: ";",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values to restore after test
			originalSeparator := scopes.ScopeSeparator
			defer func() {
				scopes.ScopeSeparator = originalSeparator
			}()

			// Set custom separator
			scopes.ScopeSeparator = tt.separator

			scopeSlice := scopes.ParseScopes(tt.original)
			result := scopes.JoinScopes(scopeSlice)
			assert.Equal(t, tt.original, result)
		})
	}
}
