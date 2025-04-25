package utils

import (
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NormalizeDisplayName extracts a username from an email address and normalizes it for use as a name.
// 
// Example:
//
//	john.doe@example.com -> "John Doe"
//	user123@example.com -> "User"
func NormalizeDisplayName(email string) string {
	// Extract the username part (before @)
	username := email
	if atIndex := strings.Index(email, "@"); atIndex > 0 {
		username = email[:atIndex]
	}

	// Remove special characters and replace with spaces
	reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	username = reg.ReplaceAllString(username, " ")

	// Trim spaces and convert to title case
	username = strings.TrimSpace(username)
	caser := cases.Title(language.English)
	username = caser.String(strings.ToLower(username))

	// Remove extra spaces
	spaceReg := regexp.MustCompile(`\s+`)
	username = spaceReg.ReplaceAllString(username, " ")

	// Remove numbers from start and end of username
	username = regexp.MustCompile(`^\d+`).ReplaceAllString(username, "")  // Remove numbers from start
	username = regexp.MustCompile(`\d+$`).ReplaceAllString(username, "")  // Remove numbers from end
	username = strings.TrimSpace(username)  // Trim any spaces that might have been created

	return username
}

// Deprecated: GetNormalizedDisplayName is deprecated, use NormalizeDisplayName instead.
func GetNormalizedDisplayName(email string) string {
	return NormalizeDisplayName(email)
}
