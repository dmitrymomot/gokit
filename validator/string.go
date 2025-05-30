package validator

import (
	"errors"
	"net"
	"reflect"
	"regexp"
	"strings"
)

var (
	emailRegex         = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}` + `[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}` + `[a-zA-Z0-9])?)*$`)
	alphaRegex         = regexp.MustCompile(`^[A-Za-z]+$`)
	alphanumRegex      = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	alphaSpaceRegex    = regexp.MustCompile(`^[A-Za-z ]+$`)
	alphaSpaceNumRegex = regexp.MustCompile(`^[A-Za-z0-9 ]+$`)
	asciiRegex         = regexp.MustCompile(`^[\x00-\x7F]*$`)
)

func regexValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 || strings.TrimSpace(params[0]) == "" {
		return errors.New(translator("validation.internal.missing_regex_pattern", label))
	}
	pattern := params[0]
	re, err := regexp.Compile(pattern)
	if err != nil {
		return errors.New(translator("validation.internal.invalid_regex_pattern", label, pattern))
	}
	str, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if !re.MatchString(str) {
		return errors.New(translator("validation.regex", label, params...))
	}
	return nil
}

func emailValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	
	// Basic email format check using regex
	if !emailRegex.MatchString(val) {
		return errors.New(translator("validation.email", label, params...))
	}
	
	// More thorough checks for the email
	parts := strings.Split(val, "@")
	if len(parts) != 2 {
		return errors.New(translator("validation.email", label, params...))
	}
	
	// Check local part (username)
	localPart := parts[0]
	if len(localPart) == 0 || len(localPart) > 64 {
		return errors.New(translator("validation.email", label, params...))
	}
	
	// Check domain part
	domain := parts[1]
	if len(domain) < 3 || len(domain) > 255 {
		return errors.New(translator("validation.email", label, params...))
	}
	
	// Check domain format - must have at least one dot
	if !strings.Contains(domain, ".") {
		return errors.New(translator("validation.email", label, params...))
	}
	
	// Check TLD format
	domainParts := strings.Split(domain, ".")
	tld := domainParts[len(domainParts)-1]
	
	// TLD must be at least 2 characters
	if len(tld) < 2 {
		return errors.New(translator("validation.email", label, params...))
	}
	
	// TLD should be letters only
	if !alphaRegex.MatchString(tld) {
		return errors.New(translator("validation.email", label, params...))
	}
	
	return nil
}

var nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z ]*[a-zA-Z]$`)

// fullnameValidator checks if a field is a valid full name.
func fullnameValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if len(val) < 3 {
		return errors.New(translator("validation.fullname", label, params...))
	}
	if strings.Contains(val, "  ") || strings.Contains(val, "--") {
		return errors.New(translator("validation.fullname", label, params...))
	}
	if strings.HasPrefix(val, " ") || strings.HasPrefix(val, "-") || strings.HasSuffix(val, " ") || strings.HasSuffix(val, "-") {
		return errors.New(translator("validation.fullname", label, params...))
	}
	if strings.ContainsAny(val, "0123456789") {
		return errors.New(translator("validation.fullname", label, params...))
	}
	return nil
}

// nameValidator checks if a field is a valid name (only letters and spaces, min length 2).
func nameValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if len(val) < 2 {
		return errors.New(translator("validation.name", label, params...))
	}
	if !nameRegex.MatchString(val) {
		return errors.New(translator("validation.name", label, params...))
	}
	return nil
}

func realEmailValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	
	// First perform basic email validation
	if err := emailValidator(fieldValue, fieldType, params, label, translator); err != nil {
		return errors.New(translator("validation.realemail", label, params...))
	}
	
	// Then check if domain has MX records
	parts := strings.Split(val, "@")
	domain := parts[1]
	
	// Check if domain has valid MX records
	_, err := net.LookupMX(domain)
	if err != nil {
		return errors.New(translator("validation.realemail", label, params...))
	}
	
	return nil
}

func alphaValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if alphaRegex.MatchString(val) {
		return nil
	}
	return errors.New(translator("validation.alpha", label, params...))
}

func alphanumValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if alphanumRegex.MatchString(val) {
		return nil
	}
	return errors.New(translator("validation.alphanum", label, params...))
}

func alphaSpaceValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if alphaSpaceRegex.MatchString(val) {
		return nil
	}
	return errors.New(translator("validation.alphaspace", label, params...))
}

func alphaSpaceNumValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if alphaSpaceNumRegex.MatchString(val) {
		return nil
	}
	return errors.New(translator("validation.alphaspacenum", label, params...))
}

func startsWithValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if strings.HasPrefix(val, params[0]) {
		return nil
	}
	return errors.New(translator("validation.startswith", label, params...))
}

func endsWithValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if strings.HasSuffix(val, params[0]) {
		return nil
	}
	return errors.New(translator("validation.endswith", label, params...))
}

func containsValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}

	if len(params) == 0 {
		return nil
	}

	if strings.Contains(val, params[0]) {
		return nil
	}
	return errors.New(translator("validation.contains", label, params...))
}

func notContainsValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if !strings.Contains(val, params[0]) {
		return nil
	}
	return errors.New(translator("validation.notcontains", label, params...))
}

func asciiValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if asciiRegex.MatchString(val) {
		return nil
	}
	return errors.New(translator("validation.ascii", label, params...))
}
