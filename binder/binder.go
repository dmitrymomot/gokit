package binder

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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
		return ErrUnsupportedType
	}

	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()

	if rt.Kind() != reflect.Struct {
		return ErrUnsupportedType
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		name := getFieldName(field)
		if name == "-" || name == "" {
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

	default:
		return ErrUnsupportedType
	}
	
	return nil
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
		val, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return errors.Join(ErrInvalidFormData, err)
		}
		value.SetInt(val)
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
	default:
		return ErrUnsupportedType
	}
	
	return nil
}
