package sanitizer

import (
	"reflect"
	"strings"
)

// SanitizeFunc defines the signature of a sanitization function.
type SanitizeFunc func(fieldValue any, fieldType reflect.StructField, params []string) any

// RegisterSanitizer registers a custom sanitization function.
func RegisterSanitizer(tag string, fn SanitizeFunc) {
	if tag == "" || fn == nil {
		return
	}
	sanitizersMutex.Lock()
	defer sanitizersMutex.Unlock()
	if sanitizers == nil {
		sanitizers = make(map[string]SanitizeFunc)
	}
	sanitizers[tag] = fn
}

// SanitizeStruct sanitizes the struct fields based on 'sanitize' tags.
func SanitizeStruct(s any) {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)
		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}
		sanitizeTag := fieldType.Tag.Get("sanitize")
		if sanitizeTag == "" {
			continue
		}
		rules := strings.Split(sanitizeTag, ",")
		for _, rule := range rules {
			ruleName, params := parseSanitizerRule(rule)
			sanitizersMutex.RLock()
			sanitizerFunc, ok := sanitizers[ruleName]
			sanitizersMutex.RUnlock()
			if ok {
				newValue := sanitizerFunc(fieldVal.Interface(), fieldType, params)
				if fieldVal.CanSet() {
					fieldVal.Set(reflect.ValueOf(newValue))
				}
			}
		}
	}
}

// parseSanitizerRule splits a rule into its name and parameters.
func parseSanitizerRule(rule string) (string, []string) {
	parts := strings.SplitN(rule, ":", 2)
	ruleName := parts[0]
	var params []string
	if len(parts) > 1 {
		params = strings.Split(parts[1], ":")
	}
	return ruleName, params
}
