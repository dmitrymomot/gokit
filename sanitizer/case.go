package sanitizer

import (
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

// capitalizeSanitizer capitalizes the first letter of a string.
func capitalizeSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok && len(v) > 0 {
		return strings.ToUpper(string(v[0])) + v[1:]
	}
	return fieldValue
}

// camelCaseSanitizer converts a string to lowerCamelCase.
func camelCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToLowerCamel(v)
	}
	return fieldValue
}

// pascalCaseSanitizer converts a string to PascalCase.
func pascalCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToCamel(v)
	}
	return fieldValue
}

// snakeCaseSanitizer converts a string to snake_case.
func snakeCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToSnake(v)
	}
	return fieldValue
}

// kebabCaseSanitizer converts a string to kebab-case.
func kebabCaseSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok {
		return strcase.ToKebab(v)
	}
	return fieldValue
}

// ucfirstSanitizer capitalizes the first letter of a string.
// This is an alias for capitalizeSanitizer for compatibility.
func ucfirstSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	return capitalizeSanitizer(fieldValue, fieldType, params)
}

// lcfirstSanitizer lowercases the first letter of a string.
func lcfirstSanitizer(fieldValue any, fieldType reflect.StructField, params []string) any {
	if v, ok := fieldValue.(string); ok && len(v) > 0 {
		return strings.ToLower(string(v[0])) + v[1:]
	}
	return fieldValue
}
