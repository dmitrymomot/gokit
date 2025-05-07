package validator

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// ValidationFunc defines the signature of a validation function.
type ValidationFunc func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error

// Validator struct holds instance-specific data like the error translator function,
// and string separators for rules and parameters.
type Validator struct {
	errorTranslator    ErrorTranslatorFunc
	ruleSeparator      string // e.g., ";"
	paramSeparator     string // e.g., ":"
	paramListSeparator string // e.g., ","
}

// ErrorTranslatorFunc defines the signature for error translation functions.
type ErrorTranslatorFunc func(key string, label string, params ...string) string

// NewValidator creates a new Validator instance with the provided error translator function
// and default separators (";" for rules, ":" for params, "," for param lists).
// It panics if the default separators are somehow invalid, which indicates an internal issue.
func NewValidator(errorTranslator ErrorTranslatorFunc) *Validator {
	v, err := NewValidatorWithSeparators(errorTranslator, ";", ":", ",")
	if err != nil {
		// This should not happen with default separators
		panic("gokit/validator: internal error with default separator configuration: " + err.Error())
	}
	return v
}

// NewValidatorWithSeparators creates a new Validator instance with the provided error translator
// function and custom string separators.
// It returns an error if separators are empty, not single characters, or not distinct.
func NewValidatorWithSeparators(errorTranslator ErrorTranslatorFunc, ruleSep, paramSep, paramListSep string) (*Validator, error) {
	// Validate separators: must be non-empty, single-character strings
	if ruleSep == "" || paramSep == "" || paramListSep == "" {
		return nil, ErrInvalidSeparatorConfiguration // Or a more specific error like ErrEmptySeparator
	}
	if len(ruleSep) != 1 || len(paramSep) != 1 || len(paramListSep) != 1 {
		return nil, ErrInvalidSeparatorConfiguration // Or a more specific error like ErrMultiCharSeparator
	}

	if ruleSep == paramSep || ruleSep == paramListSep || paramSep == paramListSep {
		return nil, ErrInvalidSeparatorConfiguration
	}

	if errorTranslator == nil {
		errorTranslator = defaultErrorTranslator
	}
	return &Validator{
		errorTranslator:    errorTranslator,
		ruleSeparator:      ruleSep,
		paramSeparator:     paramSep,
		paramListSeparator: paramListSep,
	}, nil
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
func (v *Validator) ValidateStruct(s any) error {
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
	for i := range val.NumField() {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}

		validationTag := fieldType.Tag.Get("validate")
		if validationTag == "" {
			// Only handle nested structs if there is no validation tag
			if fieldVal.Kind() == reflect.Struct {
				v.validateFields(fieldVal, fieldType.Type, prefix, errors)
			}
			continue
		}

		// Get label or field name
		label := fieldType.Tag.Get("label")
		if label == "" {
			label = fieldType.Name
		}

		// Split rules by the ruleSeparator
		var rules []string
		for ruleStr := range strings.SplitSeq(validationTag, v.ruleSeparator) {
			rules = append(rules, ruleStr)
		}

		// Check for 'omitempty' rule first
		hasOmitempty := false
		for _, ruleStr := range rules {
			if strings.TrimSpace(ruleStr) == "omitempty" {
				hasOmitempty = true
				break
			}
		}

		if hasOmitempty && isZero(fieldVal.Interface()) {
			continue // Skip validation for this field
		}

		for _, ruleStr := range rules {
			trimmedRuleStr := strings.TrimSpace(ruleStr)
			if trimmedRuleStr == "" || trimmedRuleStr == "omitempty" { // Also skip omitempty here as it's handled
				continue
			}
			ruleName, params := v.parseRule(trimmedRuleStr)
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
			} else {
				// Rule not found, add a specific error for developers
				fieldName := fieldType.Name
				if prefix != "" {
					fieldName = prefix + "." + fieldName
				}
				errors.Add(fieldName, fmt.Sprintf("validation: unknown rule '%s' for field '%s'", ruleName, fieldName))
			}
		}
	}
}

// isZero checks if the value is the zero value for its type.
func isZero(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}

// parseRule splits a rule into its name and parameters using the configured paramSeparator.
// It also handles splitting of parameter lists using the configured paramListSeparator.
func (v *Validator) parseRule(rule string) (string, []string) {
	// Mimic SplitN(rule, v.paramSeparator, 2)
	// to separate rule name from the full parameter string.
	var partsSlice []string
	for part := range strings.SplitSeq(rule, v.paramSeparator) {
		partsSlice = append(partsSlice, part)
	}

	ruleName := strings.TrimSpace(partsSlice[0])
	var paramsStr string
	if len(partsSlice) > 1 {
		// Re-join the rest if paramSeparator was in the params part, effectively taking only the first split.
		paramsStr = strings.TrimSpace(strings.Join(partsSlice[1:], v.paramSeparator))
	}

	var params []string
	if paramsStr != "" {
		// Special handling for "regex" to treat its entire param string as a single parameter.
		if ruleName == "regex" {
			params = []string{paramsStr}
		} else {
			// For other rules, split the paramsStr by the paramListSeparator (e.g., comma).
			var paramPartsSlice []string
			for p := range strings.SplitSeq(paramsStr, v.paramListSeparator) {
				paramPartsSlice = append(paramPartsSlice, p)
			}
			for _, p := range paramPartsSlice {
				params = append(params, strings.TrimSpace(p))
			}
		}
	}
	return ruleName, params
}
