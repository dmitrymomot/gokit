package validator

import (
	"fmt"
	"maps"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

// ValidationFunc defines the signature of a validation function.
type ValidationFunc func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error

// Validator struct holds instance-specific data like the error translator function,
// string separators for rules and parameters, and its own validators map.
type Validator struct {
	errorTranslator    ErrorTranslatorFunc
	ruleSeparator      string                    // e.g., ";"
	paramSeparator     string                    // e.g., ":"
	paramListSeparator string                    // e.g., ","
	validators         map[string]ValidationFunc // Instance-specific validators
	validatorsMutex    sync.RWMutex              // Instance-specific mutex
}

// ErrorTranslatorFunc defines the signature for error translation functions.
type ErrorTranslatorFunc func(key string, label string, params ...string) string

// New creates a new Validator instance with the provided options
func New(options ...Option) (*Validator, error) {
	// Create with default values
	v := &Validator{
		errorTranslator:    defaultErrorTranslator,
		ruleSeparator:      ";",
		paramSeparator:     ":",
		paramListSeparator: ",",
		validators:         make(map[string]ValidationFunc),
	}

	// Copy built-in validators to the instance
	// This ensures that each instance has access to the global validators
	// without modifying the global map, allowing for instance-specific overrides
	// or additions via RegisterValidation.
	v.validatorsMutex.Lock()
	maps.Copy(v.validators, builtInValidators)
	v.validatorsMutex.Unlock()

	// Apply all provided options
	for _, option := range options {
		if err := option(v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

// RegisterValidation registers a custom validation function for this validator instance
func (v *Validator) RegisterValidation(tag string, fn ValidationFunc) error {
	if tag == "" || fn == nil {
		return ErrInvalidValidatorConfiguration
	}
	v.validatorsMutex.Lock()
	defer v.validatorsMutex.Unlock()
	v.validators[tag] = fn
	return nil
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
			if fieldVal.Kind() == reflect.Struct {
				// Propagate prefix for nested fields
				nestedPrefix := fieldType.Name
				if prefix != "" {
					nestedPrefix = prefix + "." + fieldType.Name
				}
				v.validateFields(fieldVal, fieldType.Type, nestedPrefix, errors)
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
		for _, ruleStr := range strings.Split(validationTag, v.ruleSeparator) {
			rules = append(rules, ruleStr)
		}

		// Only skip validation if there is a standalone 'omitempty' rule
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
			if trimmedRuleStr == "" || trimmedRuleStr == "omitempty" {
				continue
			}
			ruleName, params := v.parseRule(trimmedRuleStr)
			v.validatorsMutex.RLock()
			validatorFunc, ok := v.validators[ruleName]
			v.validatorsMutex.RUnlock()
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
