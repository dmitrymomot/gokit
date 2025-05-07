package validator

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
)

// base64Validator checks if a field is valid Base64-encoded data.
func base64Validator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if _, err := base64.StdEncoding.DecodeString(val); err != nil {
		return errors.New(translator("validation.base64", label, params...))
	}
	return nil
}

// jsonValidator checks if a field is valid JSON.
func jsonValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	var js json.RawMessage
	if err := json.Unmarshal([]byte(val), &js); err != nil {
		return errors.New(translator("validation.json", label, params...))
	}
	return nil
}

var semverRegex = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:[-+][0-9A-Za-z.-]+)?$`)

// semverValidator checks if a field is a valid semantic version.
func semverValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.type_mismatch", label, params...))
	}
	if !semverRegex.MatchString(val) {
		return errors.New(translator("validation.semver", label, params...))
	}
	return nil
}
