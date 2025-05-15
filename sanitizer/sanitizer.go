package sanitizer

import (
	"maps"
	"reflect"
	"strings"
	"sync"
)

// SanitizeFunc defines the signature of a sanitization function.
type SanitizeFunc func(fieldValue any, fieldType reflect.StructField, params []string) any

// Sanitizer holds instance-specific configuration and sanitizer map.
type Sanitizer struct {
	ruleSeparator      string
	paramSeparator     string
	paramListSeparator string
	fieldNameTag       string

	sanitizers   map[string]SanitizeFunc
	sanitizersMu sync.RWMutex
}

// New creates a new Sanitizer with the given options.
func New(options ...Option) (*Sanitizer, error) {
	s := &Sanitizer{
		ruleSeparator:      ";",
		paramSeparator:     ":",
		paramListSeparator: ",",
		sanitizers:         make(map[string]SanitizeFunc),
	}

	s.sanitizersMu.Lock()
	maps.Copy(s.sanitizers, builtInSanitizers)
	s.sanitizersMu.Unlock()

	// Apply options
	for _, opt := range options {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// MustNew creates a new Sanitizer with the given options and panics on error.
func MustNew(options ...Option) *Sanitizer {
	s, err := New(options...)
	if err != nil {
		panic(err)
	}
	return s
}

// RegisterSanitizer registers a custom sanitization function.
func (s *Sanitizer) RegisterSanitizer(tag string, fn SanitizeFunc) error {
	if tag == "" || fn == nil {
		return ErrInvalidSanitizerConfiguration
	}
	s.sanitizersMu.Lock()
	defer s.sanitizersMu.Unlock()
	s.sanitizers[tag] = fn
	return nil
}

// SanitizeStruct sanitizes the struct fields based on 'sanitize' tags.
func (s *Sanitizer) SanitizeStruct(ptr any) error {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return nil
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil
	}

	return s.sanitizeValue(val)
}

// sanitizeValue recursively processes a reflect.Value for sanitization.
func (s *Sanitizer) sanitizeValue(val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}

		// Process nested structs (if pointer, dereference it)
		if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() && fieldVal.Elem().Kind() == reflect.Struct {
			if err := s.sanitizeValue(fieldVal.Elem()); err != nil {
				return err
			}
			continue
		}

		// Process nested structs (if direct struct value)
		if fieldVal.Kind() == reflect.Struct {
			if err := s.sanitizeValue(fieldVal); err != nil {
				return err
			}
			continue
		}

		sanitizeTag := fieldType.Tag.Get("sanitize")
		if sanitizeTag == "" {
			continue
		}

		// Skip sanitization for zero values if omitempty is specified
		if strings.Contains(sanitizeTag, "omitempty") && isZero(fieldVal) {
			continue
		}

		// Process sanitization rules
		rules := strings.Split(sanitizeTag, s.ruleSeparator)
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "omitempty" {
				continue
			}

			ruleName, params := s.parseRule(rule)

			s.sanitizersMu.RLock()
			sanitizerFunc, ok := s.sanitizers[ruleName]
			s.sanitizersMu.RUnlock()

			if !ok {
				continue // Skip unknown sanitizers
			}

			newValue := sanitizerFunc(fieldVal.Interface(), fieldType, params)
			if fieldVal.CanSet() {
				fieldVal.Set(reflect.ValueOf(newValue))
			}
		}
	}

	return nil
}

// parseRule splits a rule into its name and parameters.
func (s *Sanitizer) parseRule(rule string) (string, []string) {
	parts := strings.SplitN(rule, s.paramSeparator, 2)
	ruleName := strings.TrimSpace(parts[0])

	var params []string
	if len(parts) > 1 {
		params = strings.Split(parts[1], s.paramListSeparator)
		for i, p := range params {
			params[i] = strings.TrimSpace(p)
		}
	}

	return ruleName, params
}

// isZero checks if a value is the zero value for its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Array:
		return v.Len() == 0
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
