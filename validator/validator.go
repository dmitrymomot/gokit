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
	fieldNameTag       string                    // e.g., "json", "xml", "form", etc.
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
		paramSeparator:     ":", // Default parameter separator
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
				nestedPrefix := v.getFieldNameByTag(fieldType)
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
		rules := strings.Split(validationTag, v.ruleSeparator)

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
					fieldName := v.getFieldNameByTag(fieldType)
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

// getFieldNameByTag retrieves the field name from a specified tag.
// If the tag value is empty, it returns the fallback value.
func (v *Validator) getFieldNameByTag(field reflect.StructField) string {
	if tagValue := field.Tag.Get(v.fieldNameTag); tagValue != "" {
		return tagValue
	}
	return field.Name
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
	parts := strings.SplitN(rule, v.paramSeparator, 2)
	ruleName := strings.TrimSpace(parts[0])
	var paramsStr string
	if len(parts) > 1 {
		paramsStr = strings.TrimSpace(parts[1])
	}

	var params []string
	if paramsStr != "" {
		// Special handling for "regex" to treat its entire param string as a single parameter.
		if ruleName == "regex" {
			params = []string{paramsStr}
		} else {
			// For other rules, split the paramsStr by the paramListSeparator (e.g., comma).
			paramParts := strings.Split(paramsStr, v.paramListSeparator)
			for _, p := range paramParts {
				params = append(params, strings.TrimSpace(p))
			}
		}
	}
	return ruleName, params
}
