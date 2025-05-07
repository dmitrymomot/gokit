package validator

import (
	"errors"
	"net"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	domainRegex   = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)
	macRegex      = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	phoneRegex    = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	usernameRegex = regexp.MustCompile(`^[A-Za-z0-9_]{3,16}$`)
	hexcolorRegex = regexp.MustCompile(`^#?([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	uuidRegex     = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
)

func urlValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	parsedURL, err := url.Parse(val)
	if err != nil || !parsedURL.IsAbs() {
		return errors.New(translator("validation.url", label, params...))
	}
	return nil
}

// ipv4Validator checks if a field is a valid IPv4 address.
func ipv4Validator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	ip := net.ParseIP(val)
	if ip == nil || ip.To4() == nil {
		return errors.New(translator("validation.ipv4", label, params...))
	}
	return nil
}

// ipv6Validator checks if a field is a valid IPv6 address.
func ipv6Validator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	ip := net.ParseIP(val)
	if ip == nil || ip.To4() != nil {
		return errors.New(translator("validation.ipv6", label, params...))
	}
	return nil
}

func ipValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if net.ParseIP(val) == nil {
		return errors.New(translator("validation.ip", label, params...))
	}
	return nil
}

func domainValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if !domainRegex.MatchString(val) {
		return errors.New(translator("validation.domain", label, params...))
	}
	return nil
}

func macValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if !macRegex.MatchString(val) {
		return errors.New(translator("validation.mac", label, params...))
	}
	return nil
}

func portValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	// Port can be string or numeric, so type check is different
	switch v := fieldValue.(type) {
	case string:
		if v == "" {
			return nil
		}
		portNum, err := strconv.Atoi(v)
		if err != nil {
			return errors.New(translator("validation.port", label, params...)) // Not a number
		}
		if portNum < 1 || portNum > 65535 {
			return errors.New(translator("validation.port", label, params...))
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		portNum := reflect.ValueOf(fieldValue).Convert(reflect.TypeOf(int64(0))).Int()
		if portNum < 1 || portNum > 65535 {
			return errors.New(translator("validation.port", label, params...))
		}
	default:
		// Not a string or any standard integer type
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	return nil
}

func phoneValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if !phoneRegex.MatchString(val) {
		return errors.New(translator("validation.phone", label, params...))
	}
	return nil
}

func usernameValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if !usernameRegex.MatchString(val) {
		return errors.New(translator("validation.username", label, params...))
	}
	return nil
}

func slugValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if len(val) < 3 {
		return errors.New(translator("validation.slug", label, params...))
	}
	if strings.HasPrefix(val, "_") || strings.HasPrefix(val, "-") ||
		strings.HasSuffix(val, "_") || strings.HasSuffix(val, "-") ||
		strings.Contains(val, "__") || strings.Contains(val, "--") {
		return errors.New(translator("validation.slug", label, params...))
	}
	return nil
}

func hexcolorValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if !hexcolorRegex.MatchString(val) {
		return errors.New(translator("validation.hexcolor", label, params...))
	}
	return nil
}

func extensionValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" || len(params) == 0 {
		return nil
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(val)), ".")
	for _, allowed := range params {
		if strings.ToLower(allowed) == ext {
			return nil
		}
	}
	return errors.New(translator("validation.extension", label, params...))
}

func uuidValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if val == "" {
		return nil
	}
	if !uuidRegex.MatchString(val) {
		return errors.New(translator("validation.uuid", label, params...))
	}
	return nil
}
