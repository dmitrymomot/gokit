package binder

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Common time layouts to try when parsing time strings
var timeLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// Bind binds the request data to the provided struct based on content type
func Bind(r *http.Request, v any) error {
	if r == nil {
		return ErrInvalidRequest
	}

	contentType := r.Header.Get("Content-Type")
	contentType = strings.TrimSpace(strings.Split(contentType, ";")[0])

	// Use content-type based binding
	switch contentType {
	case "application/json":
		return BindJSON(r, v)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		return BindForm(r, v)
	default:
		// For GET requests with no content type, use query parameter binding
		if r.Method == http.MethodGet {
			return BindQuery(r, v)
		}
		return ErrInvalidContentType
	}
}

// BindJSON binds JSON request body to a struct
func BindJSON(r *http.Request, v any) error {
	if r == nil {
		return ErrInvalidRequest
	}

	if r.Body == nil {
		return ErrEmptyBody
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errors.Join(ErrInvalidJSON, err)
	}

	if len(body) == 0 {
		return ErrEmptyBody
	}

	if err := json.Unmarshal(body, v); err != nil {
		return errors.Join(ErrInvalidJSON, err)
	}

	return nil
}

// BindForm binds form data to a struct
func BindForm(r *http.Request, v any) error {
	if r == nil {
		return ErrInvalidRequest
	}

	// Check content type to handle multipart forms appropriately
	contentType := r.Header.Get("Content-Type")
	isMultipart := strings.Contains(contentType, "multipart/form-data")

	if isMultipart {
		// For multipart forms, use ParseMultipartForm
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
			return errors.Join(ErrInvalidFormData, err)
		}
		// Use r.Form which contains both URL query parameters and form data
		return bindValues(r.Form, v)
	}

	// For regular forms, use ParseForm
	if err := r.ParseForm(); err != nil {
		return errors.Join(ErrInvalidFormData, err)
	}

	return bindValues(r.Form, v)
}

// BindQuery binds query parameters to a struct
func BindQuery(r *http.Request, v any) error {
	if r == nil {
		return ErrInvalidRequest
	}

	return bindValues(r.URL.Query(), v)
}

// bindValues binds url.Values to a struct
func bindValues(values url.Values, v any) error {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return fmt.Errorf("%w: non-pointer value", ErrUnsupportedType)
	}

	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("%w: non-struct value", ErrUnsupportedType)
	}

	for i := range rt.NumField() {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		name := getFieldName(field)
		if name == "-" || name == "" {
			continue
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			// Skip time.Time as it needs special handling
			if field.Type == reflect.TypeOf(time.Time{}) {
				if vals, ok := values[name]; ok && len(vals) > 0 {
					if err := setTimeField(fieldValue, vals[0]); err != nil {
						return err
					}
				}
				continue
			}

			// For embedded or nested structs, bind with field name as prefix
			if err := bindNestedStruct(values, fieldValue, name); err != nil {
				return err
			}
			continue
		}

		vals, ok := values[name]
		if !ok || len(vals) == 0 {
			continue
		}

		if err := setFieldValue(fieldValue, vals); err != nil {
			return err
		}
	}

	return nil
}

// bindNestedStruct binds values to a nested struct with the given prefix
func bindNestedStruct(values url.Values, structValue reflect.Value, prefix string) error {
	if structValue.Kind() != reflect.Struct {
		return nil
	}

	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		name := getFieldName(field)
		if name == "-" || name == "" {
			continue
		}

		// Construct the prefixed field name
		prefixedName := prefix + "." + name

		// Handle nested structs recursively
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			if err := bindNestedStruct(values, fieldValue, prefixedName); err != nil {
				return err
			}
			continue
		}

		vals, ok := values[prefixedName]
		if !ok || len(vals) == 0 {
			continue
		}

		if err := setFieldValue(fieldValue, vals); err != nil {
			return err
		}
	}

	return nil
}

// getFieldName returns the form field name for a struct field
func getFieldName(field reflect.StructField) string {
	// Check for form tag first, then json tag
	tag := field.Tag.Get("form")
	if tag == "" {
		tag = field.Tag.Get("json")
	}

	// Extract the name from the tag (before any options)
	if tag != "" {
		parts := strings.Split(tag, ",")
		return parts[0]
	}

	// Default to field name
	return field.Name
}

// setTimeField sets a time.Time field from a string value
func setTimeField(field reflect.Value, timeStr string) error {
	if timeStr == "" {
		return nil
	}

	// Try parsing with various time layouts
	for _, layout := range timeLayouts {
		if t, err := time.Parse(layout, timeStr); err == nil {
			field.Set(reflect.ValueOf(t))
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrUnsupportedTimeFormat, timeStr)
}

// setFieldValue sets the value of a struct field based on form values
func setFieldValue(fieldValue reflect.Value, values []string) error {
	if len(values) == 0 {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(values[0])

	case reflect.Bool:
		val, err := strconv.ParseBool(values[0])
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		fieldValue.SetBool(val)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(values[0])
			if err != nil {
				return errors.Join(ErrInvalidFormData, err)
			}
			fieldValue.SetInt(int64(duration))
		} else {
			val, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return errors.Join(ErrInvalidFormData, err)
			}
			fieldValue.SetInt(val)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(values[0], 10, 64)
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		fieldValue.SetUint(val)

	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		fieldValue.SetFloat(val)

	case reflect.Slice:
		elemType := fieldValue.Type().Elem()
		slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

		for i, val := range values {
			elem := reflect.New(elemType).Elem()
			if err := setSingleValue(elem, val); err != nil {
				return err
			}
			slice.Index(i).Set(elem)
		}

		fieldValue.Set(slice)

	case reflect.Map:
		// Only support maps with string keys
		if fieldValue.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("%w: non-string map key", ErrInvalidMapKey)
		}

		// Create a new map if it's nil
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.MakeMap(fieldValue.Type()))
		}

		// Set map values
		elemType := fieldValue.Type().Elem()

		// For maps, we expect param names in the form of map[key]
		// and we need to extract the key and set the value
		for k, v := range extractMapValues(values[0]) {
			elem := reflect.New(elemType).Elem()
			if err := setSingleValue(elem, v); err != nil {
				return err
			}
			fieldValue.SetMapIndex(reflect.ValueOf(k), elem)
		}

	default:
		// Try to handle custom types via TextUnmarshaler interface
		if reflect.PointerTo(fieldValue.Type()).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) ||
			reflect.PointerTo(fieldValue.Type()).Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
			// Create a new pointer to the value
			ptr := reflect.New(fieldValue.Type())
			ptr.Elem().Set(fieldValue)

			// Try json.Unmarshaler first
			if unmarshaler, ok := ptr.Interface().(json.Unmarshaler); ok {
				if err := unmarshaler.UnmarshalJSON([]byte(`"` + values[0] + `"`)); err != nil {
					return fmt.Errorf("%w: %v", ErrInvalidFormData, err)
				}
				fieldValue.Set(ptr.Elem())
				return nil
			}

			// Try encoding.TextUnmarshaler
			if unmarshaler, ok := ptr.Interface().(encoding.TextUnmarshaler); ok {
				if err := unmarshaler.UnmarshalText([]byte(values[0])); err != nil {
					return fmt.Errorf("%w: %v", ErrInvalidFormData, err)
				}
				fieldValue.Set(ptr.Elem())
				return nil
			}
		}

		return fmt.Errorf("%w: %T", ErrUnsupportedType, fieldValue.Type())
	}

	return nil
}

// extractMapValues extracts map values from a string in the format key1=value1,key2=value2
func extractMapValues(str string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(str, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		}
	}

	return result
}

// setSingleValue sets a single value to a reflect.Value
func setSingleValue(value reflect.Value, strVal string) error {
	switch value.Kind() {
	case reflect.String:
		value.SetString(strVal)
	case reflect.Bool:
		val, err := strconv.ParseBool(strVal)
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		value.SetBool(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(strVal)
			if err != nil {
				return fmt.Errorf("%w: %v", ErrInvalidFormData, err)
			}
			value.SetInt(int64(duration))
		} else {
			val, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return errors.Join(ErrInvalidFormData, err)
			}
			value.SetInt(val)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		value.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		value.SetFloat(val)
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return setTimeField(value, strVal)
		}
		return fmt.Errorf("%w: %T", ErrUnsupportedType, value.Type())
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedType, value.Type())
	}

	return nil
}
