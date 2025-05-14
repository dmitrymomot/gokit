package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// requiredValidator checks that a value is not the zero value for its type.
func requiredValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	var valid bool

	switch value.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		valid = value.Len() > 0
	case reflect.Bool:
		valid = value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		valid = value.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		valid = value.Uint() != 0
	case reflect.Float32, reflect.Float64:
		valid = value.Float() != 0
	case reflect.Interface, reflect.Ptr:
		valid = !value.IsNil()
	default:
		zero := reflect.Zero(value.Type()).Interface()
		valid = value.Interface() != zero
	}

	if !valid {
		return errors.New(translator("validation.required", label, params...))
	}
	return nil
}

// maxValidator ensures numeric, string length, or collection length does not exceed a maximum.
func maxValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	maxStr := params[0]
	maxValue, err := strconv.ParseFloat(maxStr, 64)
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(value.Int()) > maxValue {
			return errors.New(translator("validation.max", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(value.Uint()) > maxValue {
			return errors.New(translator("validation.max", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() > maxValue {
			return errors.New(translator("validation.max", label, params...))
		}
	case reflect.String:
		if float64(len(value.String())) > maxValue {
			return errors.New(translator("validation.max", label, params...))
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if float64(value.Len()) > maxValue {
			return errors.New(translator("validation.max", label, params...))
		}
	}
	return nil
}

// minValidator ensures numeric, string length, or collection length is not below a minimum.
func minValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	minStr := params[0]
	minValue, err := strconv.ParseFloat(minStr, 64)
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.String:
		if float64(len(value.String())) < minValue {
			return errors.New(translator("validation.min", label, params...))
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if float64(value.Len()) < minValue {
			return errors.New(translator("validation.min", label, params...))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(value.Int()) < minValue {
			return errors.New(translator("validation.min", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(value.Uint()) < minValue {
			return errors.New(translator("validation.min", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() < minValue {
			return errors.New(translator("validation.min", label, params...))
		}
	}
	return nil
}

// rangeValidator checks that a value (numeric or length) falls within a specified range.
func rangeValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) < 2 {
		return nil
	}
	minStr, maxStr := params[0], params[1]
	minVal, err1 := strconv.ParseFloat(minStr, 64)
	maxVal, err2 := strconv.ParseFloat(maxStr, 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := float64(value.Int())
		if v < minVal || v > maxVal {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := float64(value.Uint())
		if v < minVal || v > maxVal {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		v := value.Float()
		if v < minVal || v > maxVal {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.String:
		l := float64(len(value.String()))
		if l < minVal || l > maxVal {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		l := float64(value.Len())
		if l < minVal || l > maxVal {
			return errors.New(translator("validation.range", label, params...))
		}
	}
	return nil
}

// lengthValidator ensures a string or collection length equals a specified value.
func lengthValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	expected, err := strconv.Atoi(params[0])
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	var length int
	switch value.Kind() {
	case reflect.String:
		length = len(value.String())
	case reflect.Slice, reflect.Array, reflect.Map:
		length = value.Len()
	default:
		return nil
	}
	if length != expected {
		return errors.New(translator("validation.length", label, params...))
	}
	return nil
}

// betweenValidator checks that a value or length is between two bounds inclusive.
func betweenValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) < 2 {
		return nil
	}
	minStr, maxStr := params[0], params[1]
	minVal, err1 := strconv.ParseFloat(minStr, 64)
	maxVal, err2 := strconv.ParseFloat(maxStr, 64)
	if err1 != nil || err2 != nil {
		return nil
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := float64(value.Int())
		if v < minVal || v > maxVal {
			return errors.New(translator("validation.between", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := float64(value.Uint())
		if v < minVal || v > maxVal {
			return errors.New(translator("validation.between", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		v := value.Float()
		if v < minVal || v > maxVal {
			return errors.New(translator("validation.between", label, params...))
		}
	case reflect.String:
		l := float64(len(value.String()))
		if l < minVal || l > maxVal {
			return errors.New(translator("validation.between", label, params...))
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		l := float64(value.Len())
		if l < minVal || l > maxVal {
			return errors.New(translator("validation.between", label, params...))
		}
	}
	return nil
}

// equalValidator checks if a value equals a specified parameter.
func equalValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	target := params[0]
	value := reflect.ValueOf(fieldValue)
	if str, ok := fieldValue.(string); ok {
		if str != target {
			return errors.New(translator("validation.eq", label, params...))
		}
		return nil
	}
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if strconv.FormatInt(value.Int(), 10) != target {
			return errors.New(translator("validation.eq", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if strconv.FormatUint(value.Uint(), 10) != target {
			return errors.New(translator("validation.eq", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		if fmt.Sprintf("%v", value.Float()) != target {
			return errors.New(translator("validation.eq", label, params...))
		}
	default:
		if fmt.Sprintf("%v", fieldValue) != target {
			return errors.New(translator("validation.eq", label, params...))
		}
	}
	return nil
}

// notEqualValidator checks if a value does not equal a specified parameter.
func notEqualValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	if fmt.Sprintf("%v", fieldValue) == params[0] {
		return errors.New(translator("validation.ne", label, params...))
	}
	return nil
}

// lessThanValidator checks that a numeric value or length is less than a specified parameter.
func lessThanValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	target, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	var v float64
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		v = value.Float()
	case reflect.String:
		v = float64(len(value.String()))
	case reflect.Slice, reflect.Array, reflect.Map:
		v = float64(value.Len())
	default:
		return nil
	}
	if v >= target {
		return errors.New(translator("validation.lt", label, params...))
	}
	return nil
}

// lessThanOrEqualValidator checks that a numeric value or length is less than or equal to a specified parameter.
func lessThanOrEqualValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	target, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	var v float64
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		v = value.Float()
	case reflect.String:
		v = float64(len(value.String()))
	case reflect.Slice, reflect.Array, reflect.Map:
		v = float64(value.Len())
	default:
		return nil
	}
	if v > target {
		return errors.New(translator("validation.lte", label, params...))
	}
	return nil
}

// greaterThanValidator checks that a numeric value or length is greater than a specified parameter.
func greaterThanValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	target, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	var v float64
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		v = value.Float()
	case reflect.String:
		v = float64(len(value.String()))
	case reflect.Slice, reflect.Array, reflect.Map:
		v = float64(value.Len())
	default:
		return nil
	}
	if v <= target {
		return errors.New(translator("validation.gt", label, params...))
	}
	return nil
}

// greaterThanOrEqualValidator checks that a numeric value or length is greater than or equal to a specified parameter.
func greaterThanOrEqualValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	target, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	var v float64
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v = float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		v = value.Float()
	case reflect.String:
		v = float64(len(value.String()))
	case reflect.Slice, reflect.Array, reflect.Map:
		v = float64(value.Len())
	default:
		return nil
	}
	if v < target {
		return errors.New(translator("validation.gte", label, params...))
	}
	return nil
}

// inValidator checks if a field value is in the specified list.
func inValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	valueStr := fmt.Sprintf("%v", fieldValue)
	for _, param := range params {
		if param == valueStr {
			return nil
		}
	}
	return errors.New(translator("validation.in", label, params...))
}

// notInValidator checks if a field value is not in the specified list.
func notInValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	valueStr := fmt.Sprintf("%v", fieldValue)
	for _, param := range params {
		if param == valueStr {
			return errors.New(translator("validation.notin", label, params...))
		}
	}
	return nil
}

// booleanValidator checks if a field is a boolean.
func booleanValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if _, ok := fieldValue.(bool); !ok {
		return errors.New(translator("validation.boolean", label, params...))
	}
	return nil
}
