package scopes

import (
	"sort"
	"strings"
)

var (
	// ScopeSeparator is used to separate multiple scopes in a string
	// This can be modified to use a different separator (e.g., ",")
	ScopeSeparator = " "

	// ScopeWildcard represents a wildcard scope that matches everything
	// This can be modified to use a different wildcard character (e.g., "?")
	ScopeWildcard = "*"

	// ScopeDelimiter is used to separate scope parts (e.g., "admin.read")
	// This can be modified to use a different delimiter (e.g., ":")
	ScopeDelimiter = "."
)

const (
	// Common scope constants
	ScopeRead   = "read"
	ScopeWrite  = "write"
	ScopeDelete = "delete"
	ScopeAdmin  = "admin"
)

// ParseScopes converts a space-separated string of scopes into a string slice.
//
// It handles trimming of extra spaces and removes empty entries. Returns nil if the input is empty.
//
// Example:
//
//	scopes := scopes.ParseScopes("read write admin.users")
//	// Returns: []string{"read", "write", "admin.users"}
func ParseScopes(scopesStr string) []string {
	if scopesStr == "" {
		return nil
	}

	parts := strings.Split(strings.TrimSpace(scopesStr), ScopeSeparator)
	scopes := make([]string, 0, len(parts))

	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			scopes = append(scopes, trimmed)
		}
	}

	return scopes
}

// JoinScopes converts a slice of scopes back to a space-separated string.
//
// Returns an empty string if the input slice is empty or nil.
//
// Example:
//
//	str := scopes.JoinScopes([]string{"read", "write", "admin.*"})
//	// Returns: "read write admin.*"
func JoinScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	return strings.Join(scopes, ScopeSeparator)
}

// matchScope checks if a single scope matches a pattern.
// It supports wildcards (*) and hierarchical scopes (scope1.scope2).
//
// Pattern matching rules:
// - Direct match: "read" matches "read"
// - Global wildcard: "*" matches any scope
// - Namespace wildcard: "admin.*" matches any scope starting with "admin."
func matchScope(scope, pattern string) bool {
	// Direct match or full wildcard
	if scope == pattern || pattern == ScopeWildcard {
		return true
	}

	// Handle wildcard suffix (e.g., "admin.*")
	if strings.HasSuffix(pattern, ScopeWildcard) {
		prefix := strings.TrimSuffix(pattern, ScopeWildcard)
		prefix = strings.TrimSuffix(prefix, ScopeDelimiter)
		return strings.HasPrefix(scope, prefix+ScopeDelimiter)
	}

	return false
}

// ContainsScope checks if scopes contain a specific scope.
//
// Supports wildcards and hierarchical scope matching.
//
// Example:
//
//	hasScope := scopes.ContainsScope([]string{"admin.*", "read"}, "admin.users")
//	// Returns: true (because "admin.*" matches "admin.users")
func ContainsScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if matchScope(scope, s) {
			return true
		}
	}
	return false
}

// HasAllScopes checks if scopes contain all of the required scopes.
//
// Returns true if:
// - The required slice is empty
// - The scopes include a global wildcard "*"
// - Each scope in required is matched by at least one scope in scopes
//
// Example:
//
//	hasAll := scopes.HasAllScopes(
//	    []string{"admin.*", "read", "write"},
//	    []string{"admin.users", "read"}
//	)
//	// Returns: true
func HasAllScopes(scopes, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(scopes) == 0 {
		return false
	}

	// Check for global wildcard in scopes
	for _, s := range scopes {
		if s == ScopeWildcard {
			return true
		}
	}

	for _, req := range required {
		if !ContainsScope(scopes, req) {
			return false
		}
	}
	return true
}

// HasAnyScopes checks if scopes contain any of the required scopes.
//
// Returns true if:
// - The required slice is empty
// - The scopes include a global wildcard "*"
// - At least one scope in required is matched by at least one scope in scopes
//
// Example:
//
//	hasAny := scopes.HasAnyScopes(
//	    []string{"read", "write"},
//	    []string{"delete", "read"}
//	)
//	// Returns: true
func HasAnyScopes(scopes, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(scopes) == 0 {
		return false
	}

	// Check for global wildcard in scopes
	for _, s := range scopes {
		if s == ScopeWildcard {
			return true
		}
	}

	for _, req := range required {
		if ContainsScope(scopes, req) {
			return true
		}
	}
	return false
}

// EqualScopes checks if two scope collections are identical (same scopes, regardless of order).
//
// It sorts both collections before comparison to handle different ordering.
//
// Example:
//
//	equal := scopes.EqualScopes(
//	    []string{"read", "write"},
//	    []string{"write", "read"}
//	)
//	// Returns: true
func EqualScopes(scopes1, scopes2 []string) bool {
	if len(scopes1) != len(scopes2) {
		return false
	}

	// Create copies to sort
	s1 := make([]string, len(scopes1))
	s2 := make([]string, len(scopes2))
	copy(s1, scopes1)
	copy(s2, scopes2)

	// Sort both copies
	sort.Strings(s1)
	sort.Strings(s2)

	// Compare sorted slices
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// ValidateScopes checks if all scopes are valid according to the provided valid scopes.
//
// A scope is considered valid if it matches any of the validScopes (including wildcards).
// Empty scopes are always considered valid, but empty validScopes will cause validation to fail.
//
// Example:
//
//	valid := scopes.ValidateScopes(
//	    []string{"admin.read", "user.write"},
//	    []string{"admin.*", "user.*", "system.*"}
//	)
//	// Returns: true
func ValidateScopes(scopes, validScopes []string) bool {
	if len(scopes) == 0 {
		return true
	}
	if len(validScopes) == 0 {
		return false
	}

	// Check for global wildcard in valid scopes
	for _, vs := range validScopes {
		if vs == ScopeWildcard {
			return true
		}
	}

	for _, scope := range scopes {
		isValid := false
		for _, validScope := range validScopes {
			if matchScope(scope, validScope) {
				isValid = true
				break
			}
		}
		if !isValid {
			return false
		}
	}
	return true
}

// NormalizeScopes removes duplicate scopes and sorts them alphabetically.
//
// This is useful for consistent scope handling and storage.
// Returns nil if the input slice is nil or empty.
//
// Example:
//
//	normalized := scopes.NormalizeScopes([]string{"write", "read", "read", "admin.*"})
//	// Returns: []string{"admin.*", "read", "write"}
func NormalizeScopes(inputScopes []string) []string {
	if len(inputScopes) == 0 {
		return nil
	}

	// Create a map to remove duplicates
	uniqueScopes := make(map[string]struct{}, len(inputScopes))
	for _, scopeValue := range inputScopes {
		uniqueScopes[scopeValue] = struct{}{}
	}

	// Create a new slice with unique scopes
	normalizedScopes := make([]string, 0, len(uniqueScopes))
	for uniqueScopeValue := range uniqueScopes {
		normalizedScopes = append(normalizedScopes, uniqueScopeValue)
	}

	// Sort the slice for consistent output
	sort.Strings(normalizedScopes)

	return normalizedScopes
}
