package validator

import (
	"errors"
	"reflect"
	"time"
)

// dateValidator checks if a field is a valid date with the specified format.
func dateValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	format := params[0]
	val, ok := fieldValue.(string)
	if !ok || val == "" {
		return nil
	}
	if _, err := time.Parse(format, val); err != nil {
		return errors.New(translator("validation.date", label, params...))
	}
	return nil
}

// pastDateValidator ensures a parsed date is before the current time.
func pastDateValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	format := params[0]
	val, ok := fieldValue.(string)
	if !ok || val == "" {
		return nil
	}
	t, err := time.Parse(format, val)
	if err != nil {
		return errors.New(translator("validation.pastdate", label, params...))
	}
	if !t.Before(time.Now()) {
		return errors.New(translator("validation.pastdate", label, params...))
	}
	return nil
}

// futureDateValidator ensures a parsed date is after the current time.
func futureDateValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	format := params[0]
	val, ok := fieldValue.(string)
	if !ok || val == "" {
		return nil
	}
	t, err := time.Parse(format, val)
	if err != nil {
		return errors.New(translator("validation.futuredate", label, params...))
	}
	if !t.After(time.Now()) {
		return errors.New(translator("validation.futuredate", label, params...))
	}
	return nil
}

// workdayValidator checks if a date falls on a weekday (Monday to Friday).
func workdayValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	format := params[0]
	val, ok := fieldValue.(string)
	if !ok || val == "" {
		return nil
	}
	t, err := time.Parse(format, val)
	if err != nil {
		return errors.New(translator("validation.workday", label, params...))
	}
	wd := t.Weekday()
	if wd == time.Saturday || wd == time.Sunday {
		return errors.New(translator("validation.workday", label, params...))
	}
	return nil
}

// weekendValidator checks if a date falls on a weekend (Saturday or Sunday).
func weekendValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	format := params[0]
	val, ok := fieldValue.(string)
	if !ok || val == "" {
		return nil
	}
	t, err := time.Parse(format, val)
	if err != nil {
		return errors.New(translator("validation.weekend", label, params...))
	}
	wd := t.Weekday()
	if wd != time.Saturday && wd != time.Sunday {
		return errors.New(translator("validation.weekend", label, params...))
	}
	return nil
}
