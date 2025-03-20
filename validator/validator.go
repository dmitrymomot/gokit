package validator

import (
	"net/url"
	"reflect"
	"strings"
)

// ValidationFunc defines the signature of a validation function.
type ValidationFunc func(fieldValue interface{}, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error

// Validator struct holds instance-specific data like the error translator function.
type Validator struct {
	errorTranslator ErrorTranslatorFunc
}

// ErrorTranslatorFunc defines the signature for error translation functions.
type ErrorTranslatorFunc func(key string, label string, params ...string) string

// NewValidator creates a new Validator instance with the provided error translator function.
func NewValidator(errorTranslator ErrorTranslatorFunc) *Validator {
	if errorTranslator == nil {
		errorTranslator = defaultErrorTranslator
	}
	return &Validator{
		errorTranslator: errorTranslator,
	}
}

// RegisterValidation registers a custom validation function globally.
func RegisterValidation(tag string, fn ValidationFunc) {
	if tag == "" || fn == nil {
		return
	}
	validatorsMutex.Lock()
	defer validatorsMutex.Unlock()
	if validators == nil {
		validators = make(map[string]ValidationFunc)
	}
	validators[tag] = fn
}

// ValidateStruct validates the struct fields based on 'validate' tags.
// It returns a url.Values containing validation errors.
func (v *Validator) ValidateStruct(s interface{}) error {
	errors := url.Values{}
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return ValidationErrors(errors)
	}

	v.validateFields(val, val.Type(), "", errors)
	if len(errors) > 0 {
		return ValidationErrors(errors)
	}

	return nil
}

// validateFields is a helper function to validate struct fields recursively
func (v *Validator) validateFields(val reflect.Value, typ reflect.Type, prefix string, errors url.Values) {
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}

		// Handle nested structs
		if fieldVal.Kind() == reflect.Struct {
			v.validateFields(fieldVal, fieldType.Type, prefix, errors)
			continue
		}

		validationTag := fieldType.Tag.Get("validate")
		if validationTag == "" {
			continue
		}

		// Get label or field name
		label := fieldType.Tag.Get("label")
		if label == "" {
			label = fieldType.Name
		}

		rules := strings.Split(validationTag, "|")
		for _, rule := range rules {
			ruleName, params := parseRule(rule)
			validatorsMutex.RLock()
			validatorFunc, ok := validators[ruleName]
			validatorsMutex.RUnlock()
			if ok {
				err := validatorFunc(fieldVal.Interface(), fieldType, params, label, v.errorTranslator)
				if err != nil {
					fieldName := fieldType.Name
					if prefix != "" {
						fieldName = prefix + "." + fieldName
					}
					errors.Add(fieldName, err.Error())
				}
			}
		}
	}
}

// parseRule splits a rule into its name and parameters.
func parseRule(rule string) (string, []string) {
	parts := strings.SplitN(rule, ":", 2)
	ruleName := parts[0]
	var params []string
	if len(parts) > 1 {
		// Don't split the parameters if they contain commas inside a single parameter
		// For example: "regex:^[a-z,A-Z]+$" should not be split
		if ruleName == "regex" {
			params = []string{parts[1]}
		} else {
			// For other rules, split by comma but preserve empty parts
			params = strings.Split(parts[1], ",")
			// Trim spaces from parameters
			for i := range params {
				params[i] = strings.TrimSpace(params[i])
			}
		}
	}
	return ruleName, params
}
