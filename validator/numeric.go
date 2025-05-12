package validator

import (
	"errors"
	"reflect"
	"strconv"
)

// numericValidator ensures a value is numeric (string or built-in number types).
func numericValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	switch v := fieldValue.(type) {
	case string:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return errors.New(translator("validation.numeric", label, params...))
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		// valid numeric types
	default:
		return errors.New(translator("validation.numeric", label, params...))
	}
	return nil
}

// positiveValidator ensures a numeric value is greater than zero.
func positiveValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val := reflect.ValueOf(fieldValue)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() <= 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Uint() == 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() <= 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	default:
		return errors.New(translator("validation.positive", label, params...))
	}
	return nil
}

// negativeValidator ensures a numeric value is less than zero.
func negativeValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val := reflect.ValueOf(fieldValue)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int() >= 0 {
			return errors.New(translator("validation.negative", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// unsigned cannot be negative; any non-zero value fails
		if val.Uint() > 0 {
			return errors.New(translator("validation.negative", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		if val.Float() >= 0 {
			return errors.New(translator("validation.negative", label, params...))
		}
	default:
		return errors.New(translator("validation.negative", label, params...))
	}
	return nil
}

// evenValidator ensures an integer is even.
func evenValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val := reflect.ValueOf(fieldValue)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int()%2 != 0 {
			return errors.New(translator("validation.even", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Uint()%2 != 0 {
			return errors.New(translator("validation.even", label, params...))
		}
	default:
		return errors.New(translator("validation.even", label, params...))
	}
	return nil
}

// oddValidator ensures an integer is odd.
func oddValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val := reflect.ValueOf(fieldValue)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int()%2 == 0 {
			return errors.New(translator("validation.odd", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val.Uint()%2 == 0 {
			return errors.New(translator("validation.odd", label, params...))
		}
	default:
		return errors.New(translator("validation.odd", label, params...))
	}
	return nil
}

// multipleValidator ensures a numeric value is a multiple of a given divisor.
func multipleValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	divisor, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil || divisor == 0 {
		return nil
	}
	val := reflect.ValueOf(fieldValue)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val.Int()%divisor != 0 {
			return errors.New(translator("validation.multiple", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if int64(val.Uint())%divisor != 0 {
			return errors.New(translator("validation.multiple", label, params...))
		}
	default:
		return nil
	}
	return nil
}
