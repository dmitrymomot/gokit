package sanitizer

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/iancoleman/strcase"
)

var (
	// sanitizers holds the registered sanitization functions.
	sanitizers = make(map[string]SanitizeFunc)

	// sanitizersMutex is used to synchronize access to the sanitizers map.
	sanitizersMutex sync.RWMutex
)

// ResetSanitizers restores the default sanitizers map.
// This is primarily useful for testing purposes.
func ResetSanitizers() {
	sanitizersMutex.Lock()
	defer sanitizersMutex.Unlock()
	sanitizers = map[string]SanitizeFunc{
		"trim":       trimSanitizer,
		"lower":      lowerSanitizer,
		"upper":      upperSanitizer,
		"replace":    replaceSanitizer,
		"striphtml":  stripHTMLSanitizer,
		"escape":     escapeSanitizer,
		"alphanum":   alphanumSanitizer,
		"numeric":    numericSanitizer,
		"truncate":   truncateSanitizer,
		"normalize":  normalizeSanitizer,
		"capitalize": capitalizeSanitizer,
		"camelcase":  camelCaseSanitizer,
		"snakecase":  snakeCaseSanitizer,
		"kebabcase":  kebabCaseSanitizer,
		"ucfirst":    ucfirstSanitizer,
	}
}

func trimSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.TrimSpace(v)
	}
	return fieldValue
}

func lowerSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.ToLower(v)
	}
	return fieldValue
}

func upperSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strings.ToUpper(v)
	}
	return fieldValue
}

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

func stripHTMLSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		// Simple regex to remove HTML tags
		re := regexp.MustCompile("<[^>]*>")
		return re.ReplaceAllString(v, "")
	}
	return fieldValue
}

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

func normalizeSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		// Replace different types of line endings with \n
		v = strings.ReplaceAll(v, "\r\n", "\n")
		v = strings.ReplaceAll(v, "\r", "\n")
		return v
	}
	return fieldValue
}

func capitalizeSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok && len(v) > 0 {
		return strings.ToUpper(string(v[0])) + v[1:]
	}
	return fieldValue
}

func camelCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToLowerCamel(v)
	}
	return fieldValue
}

func snakeCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToSnake(v)
	}
	return fieldValue
}

func kebabCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToKebab(v)
	}
	return fieldValue
}

func ucfirstSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok && len(v) > 0 {
		return strings.ToUpper(string(v[0])) + v[1:]
	}
	return fieldValue
}
