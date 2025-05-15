package sanitizer

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/dmitrymomot/gokit/utils"
)

// slugSanitizer converts a string to a URL-friendly slug.
func slugSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return utils.GenerateSlug(v)
	}
	return fieldValue
}

// uuidSanitizer normalizes a UUID string (lowercase, remove braces).
func uuidSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		// Remove braces if present
		v = strings.TrimPrefix(v, "{")
		v = strings.TrimSuffix(v, "}")
		// Convert to lowercase
		v = strings.ToLower(v)
		// Validate basic UUID format
		if validateUUID(v) {
			return v
		}
	}
	return fieldValue
}

// validateUUID checks if a string has a valid UUID format.
func validateUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")
	return r.MatchString(uuid)
}

// boolSanitizer converts string representations to boolean values.
func boolSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		v = strings.TrimSpace(strings.ToLower(v))
		switch v {
		case "true", "t", "yes", "y", "1", "on":
			return true
		case "false", "f", "no", "n", "0", "off", "":
			return false
		}
	}
	return fieldValue
}
