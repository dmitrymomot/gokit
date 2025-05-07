package validator

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// creditCardValidator checks if a field is a valid credit card number using the Luhn algorithm.
func creditCardValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	number, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	number = strings.ReplaceAll(number, " ", "")
	if len(number) < 13 || len(number) > 19 {
		return errors.New(translator("validation.creditcard", label, params...))
	}
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return errors.New(translator("validation.creditcard", label, params...))
		}
		if alt {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}
		sum += n
		alt = !alt
	}
	if sum%10 != 0 {
		return errors.New(translator("validation.creditcard", label, params...))
	}
	return nil
}

// passwordValidator checks if a field is a strong password.
// Password must be at least 8 characters long and contain at least
// one uppercase letter, one lowercase letter, one number, and one special character.
func passwordValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	val := reflect.ValueOf(fieldValue)
	password := val.String()
	password = strings.TrimSpace(password)

	if len(password) < 8 {
		return errors.New(translator("validation.password", label, params...))
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New(translator("validation.password", label, params...))
	}

	return nil
}
