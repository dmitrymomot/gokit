package sanitizer

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// trimSanitizer removes leading and trailing whitespace.
func trimSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.TrimSpace(v)
	}
	return fieldValue
}

// lowerSanitizer converts a string to lowercase.
func lowerSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.ToLower(v)
	}
	return fieldValue
}

// upperSanitizer converts a string to uppercase.
func upperSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.ToUpper(v)
	}
	return fieldValue
}

// replaceSanitizer replaces all occurrences of old with new.
func replaceSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if len(params) < 2 {
		return fieldValue
	}
	if v, ok := fieldValue.(string); ok {
		oldValue := params[0]
		newValue := params[1]
		return strings.ReplaceAll(v, oldValue, newValue)
	}
	return fieldValue
}

// stripHTMLSanitizer removes HTML tags from a string.
func stripHTMLSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		// Simple regex to remove HTML tags
		re := regexp.MustCompile("<[^>]*>")
		return re.ReplaceAllString(v, "")
	}
	return fieldValue
}

// escapeSanitizer replaces special HTML characters with their HTML entities.
func escapeSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		// Replace special HTML characters
		replacer := strings.NewReplacer(
			"&", "&amp;",
			"<", "&lt;",
			">", "&gt;",
			`"`, "&quot;",
			"'", "&#39;",
			"`", "&#96;",
			"!", "&#33;",
			"@", "&#64;",
			"$", "&#36;",
			"%", "&#37;",
			"(", "&#40;",
			")", "&#41;",
			"=", "&#61;",
			"+", "&#43;",
			"{", "&#123;",
			"}", "&#125;",
			"[", "&#91;",
			"]", "&#93;",
		)
		return replacer.Replace(v)
	}
	return fieldValue
}

// alphanumSanitizer removes all non-alphanumeric characters.
func alphanumSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		var sb strings.Builder
		for _, r := range v {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				sb.WriteRune(r)
			}
		}
		return sb.String()
	}
	return fieldValue
}

// numericSanitizer removes all non-numeric characters.
func numericSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		var sb strings.Builder
		for _, r := range v {
			if unicode.IsDigit(r) {
				sb.WriteRune(r)
			}
		}
		return sb.String()
	}
	return fieldValue
}

// truncateSanitizer truncates a string to a maximum length.
func truncateSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if len(params) == 0 {
		return fieldValue
	}
	maxLen, err := strconv.Atoi(params[0])
	if err != nil || maxLen < 0 {
		return fieldValue
	}
	if v, ok := fieldValue.(string); ok {
		if len(v) > maxLen {
			return v[:maxLen]
		}
	}
	return fieldValue
}

// normalizeSanitizer normalizes line endings to \n.
func normalizeSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		// Replace different types of line endings with \n
		v = strings.ReplaceAll(v, "\r\n", "\n")
		v = strings.ReplaceAll(v, "\r", "\n")
		return v
	}
	return fieldValue
}

// trimspaceSanitizer removes all whitespace from a string.
func trimspaceSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		var sb strings.Builder
		for _, r := range v {
			if !unicode.IsSpace(r) {
				sb.WriteRune(r)
			}
		}
		return sb.String()
	}
	return fieldValue
}

// emailSanitizer trims spaces and converts email to lowercase.
func emailSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.ToLower(strings.TrimSpace(v))
	}
	return fieldValue
}
