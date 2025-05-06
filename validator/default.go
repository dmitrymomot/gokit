package validator

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	// validators holds the registered validation functions.
	validators = map[string]ValidationFunc{
		"required":      requiredValidator,
		"max":           maxValidator,
		"min":           minValidator,
		"range":         rangeValidator,
		"email":         emailValidator,
		"regex":         regexValidator,
		"numeric":       numericValidator,
		"alpha":         alphaValidator,
		"alphanum":      alphanumValidator,
		"alphaspace":    alphaSpaceValidator,
		"aspace":        alphaSpaceValidator,        // Short alias for alphaspace
		"alphaspacenum": alphaSpaceNumValidator,
		"aspacenum":     alphaSpaceNumValidator,     // Short alias for alphaspacenum
		"startswith":    startsWithValidator,
		"starts":        startsWithValidator,        // Short alias for startswith
		"endswith":      endsWithValidator,
		"ends":          endsWithValidator,          // Short alias for endswith
		"contains":      containsValidator,
		"notcontains":   notContainsValidator,
		"notcont":       notContainsValidator,       // Short alias for notcontains
		"ascii":         asciiValidator,
		"positive":      positiveValidator,
		"pos":           positiveValidator,          // Short alias for positive
		"negative":      negativeValidator,
		"neg":           negativeValidator,          // Short alias for negative
		"even":          evenValidator,
		"odd":           oddValidator,
		"multiple":      multipleValidator,
		"mult":          multipleValidator,          // Short alias for multiple
		"base64":        base64Validator,
		"b64":           base64Validator,            // Short alias for base64
		"json":          jsonValidator,
		"semver":        semverValidator,
		"extension":     extensionValidator,
		"ext":           extensionValidator,         // Short alias for extension
		"url":           urlValidator,
		"ip":            ipValidator,
		"ipv4":          ipv4Validator,
		"ipv6":          ipv6Validator,
		"domain":        domainValidator,
		"mac":           macValidator,
		"port":          portValidator,
		"date":          dateValidator,
		"pastdate":      pastDateValidator,
		"past":          pastDateValidator,          // Short alias for pastdate
		"futuredate":    futureDateValidator,
		"future":        futureDateValidator,        // Short alias for futuredate
		"workday":       workdayValidator,
		"wday":          workdayValidator,           // Short alias for workday
		"weekend":       weekendValidator,
		"wend":          weekendValidator,           // Short alias for weekend
		"in":            inValidator,
		"notin":         notInValidator,
		"length":        lengthValidator,
		"len":           lengthValidator,            // Short alias for length
		"between":       betweenValidator,
		"btw":           betweenValidator,           // Short alias for between
		"boolean":       booleanValidator,
		"bool":          booleanValidator,           // Short alias for boolean
		"uuid":          uuidValidator,
		"creditcard":    creditCardValidator,
		"cc":            creditCardValidator,        // Short alias for creditcard
		"eq":            equalValidator,
		"ne":            notEqualValidator,
		"lt":            lessThanValidator,
		"lte":           lessThanOrEqualValidator,
		"gt":            greaterThanValidator,
		"gte":           greaterThanOrEqualValidator,
		"realemail":     realEmailValidator,
		"remail":        realEmailValidator,         // Short alias for realemail
		"password":      passwordValidator,
		"pwd":           passwordValidator,          // Short alias for password
		"phone":         phoneValidator,
		"username":      usernameValidator,
		"uname":         usernameValidator,          // Short alias for username
		"slug":          slugValidator,
		"hexcolor":      hexcolorValidator,
		"hcolor":        hexcolorValidator,          // Short alias for hexcolor
		"fullname":      fullnameValidator,
		"fname":         fullnameValidator,          // Short alias for fullname
		"name":          nameValidator,
	}
	// validatorsMutex is used to synchronize access to the validators map.
	validatorsMutex sync.RWMutex
)

func requiredValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	valid := true

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

func maxValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No max value specified
	}
	maxStr := params[0]
	maxValue, err := strconv.ParseFloat(maxStr, 64)
	if err != nil {
		return nil // Invalid max value specified
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
	default:
		// For other types, do nothing
	}
	return nil
}

func minValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No min value specified
	}
	minStr := params[0]
	minValue, err := strconv.ParseFloat(minStr, 64)
	if err != nil {
		return nil // Invalid min value specified
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
	default:
		// For other types, do nothing
	}
	return nil
}

// rangeValidator checks if a field is within the parameter values.
// It returns an error if the field value is not within the parameter values.
// It returns nil if the field value is within the parameter values.
// It supports string, slice, array, and map types.
// It supports int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, and float64 types.
// It supports time.Time type.
// It supports time.Duration type.
func rangeValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) != 2 {
		return nil
	}

	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min, err1 := strconv.ParseInt(params[0], 10, 64)
		max, err2 := strconv.ParseInt(params[1], 10, 64)
		if err1 != nil || err2 != nil {
			return nil
		}
		if value.Int() < min || value.Int() > max {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		min, err1 := strconv.ParseUint(params[0], 10, 64)
		max, err2 := strconv.ParseUint(params[1], 10, 64)
		if err1 != nil || err2 != nil {
			return nil
		}
		if value.Uint() < min || value.Uint() > max {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		min, err1 := strconv.ParseFloat(params[0], 64)
		max, err2 := strconv.ParseFloat(params[1], 64)
		if err1 != nil || err2 != nil {
			return nil
		}
		if value.Float() < min || value.Float() > max {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.String:
		min, err1 := strconv.Atoi(params[0])
		max, err2 := strconv.Atoi(params[1])
		if err1 != nil || err2 != nil {
			return nil
		}
		if len(value.String()) < min || len(value.String()) > max {
			return errors.New(translator("validation.range", label, params...))
		}
	case reflect.Struct:
		if fieldType.Type.String() == "time.Time" {
			min, err := time.Parse("2006-01-02", params[0])
			if err != nil {
				return nil
			}
			max, err := time.Parse("2006-01-02", params[1])
			if err != nil {
				return nil
			}
			if value.Interface().(time.Time).Before(min) || value.Interface().(time.Time).After(max) {
				return errors.New(translator("validation.range", label, params...))
			}
		}
	case reflect.Interface:
		if fieldType.Type.String() == "time.Duration" {
			min, err := time.ParseDuration(params[0])
			if err != nil {
				return nil
			}
			max, err := time.ParseDuration(params[1])
			if err != nil {
				return nil
			}
			if value.Interface().(time.Duration) < min || value.Interface().(time.Duration) > max {
				return errors.New(translator("validation.range", label, params...))
			}
		}
	default:
		return nil
	}
	return nil
}

// emailRegex is a regular expression for validating email addresses.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// emailValidator validates an email address.
func emailValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.email", label, params...))
	}
	if len(value) == 0 {
		return nil // Empty email is allowed, use required to enforce non-empty
	}
	if !emailRegex.MatchString(value) {
		return errors.New(translator("validation.email", label, params...))
	}
	return nil
}

// realEmailValidator validates an email address using a regular expression and DNS lookup.
// It is more strict than emailValidator.
// It checks if the domain has an MX record.
// It also checks if the domain is in the public suffix list.
// It does not check if the domain is reserved or has an A record.
func realEmailValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return errors.New(translator("validation.realemail", label, params...))
	}
	if len(value) == 0 {
		return nil // Empty email is allowed, use required to enforce non-empty
	}
	if !emailRegex.MatchString(value) {
		return errors.New(translator("validation.realemail", label, params...))
	}
	parts := strings.Split(value, "@")
	domain := parts[1]
	if _, err := net.LookupMX(domain); err != nil {
		return errors.New(translator("validation.realemail", label, params...))
	}
	return nil
}

// regexValidator validates a field against a regular expression.
func regexValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No pattern specified
	}
	pattern := params[0]
	value, ok := fieldValue.(string)
	if !ok {
		return nil // Not a string, skip
	}
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		return errors.New(translator("validation.regex", label, params...))
	}
	return nil
}

// numericValidator checks if a field is numeric.
func numericValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	switch v := fieldValue.(type) {
	case string:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return errors.New(translator("validation.numeric", label, params...))
		}
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		// Valid numeric types
	default:
		return errors.New(translator("validation.numeric", label, params...))
	}
	return nil
}

// alphaValidator checks if a field contains only alphabetic characters.
func alphaValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	if matched := regexp.MustCompile(`^[A-Za-z]+$`).MatchString(value); !matched {
		return errors.New(translator("validation.alpha", label, params...))
	}
	return nil
}

// alphanumValidator checks if a field contains only alphanumeric characters.
func alphanumValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	if matched := regexp.MustCompile(`^[A-Za-z0-9]+$`).MatchString(value); !matched {
		return errors.New(translator("validation.alphanum", label, params...))
	}
	return nil
}

// alphaSpaceValidator checks if a field contains only alphabetic characters and spaces.
func alphaSpaceValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	if matched := regexp.MustCompile(`^[A-Za-z ]+$`).MatchString(value); !matched {
		return errors.New(translator("validation.alphaspace", label, params...))
	}
	return nil
}

// alphaSpaceNumValidator checks if a field contains only alphanumeric characters and spaces.
func alphaSpaceNumValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	if matched := regexp.MustCompile(`^[A-Za-z0-9 ]+$`).MatchString(value); !matched {
		return errors.New(translator("validation.alphaspacenum", label, params...))
	}
	return nil
}

// urlValidator checks if a field is a valid URL.
func urlValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	_, err := url.ParseRequestURI(value)
	if err != nil {
		return errors.New(translator("validation.url", label, params...))
	}
	return nil
}

// ipValidator checks if a field is a valid IP address.
func ipValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	if net.ParseIP(value) == nil {
		return errors.New(translator("validation.ip", label, params...))
	}
	return nil
}

// dateValidator checks if a field is a valid date with the specified format.
func dateValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	format := params[0]
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	if _, err := time.Parse(format, value); err != nil {
		return errors.New(translator("validation.date", label, params...))
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

// lengthValidator checks if a field length is equal to the specified value.
func lengthValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	expectedLen, err := strconv.Atoi(params[0])
	if err != nil {
		return nil
	}
	value := reflect.ValueOf(fieldValue)
	var length int
	switch value.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map:
		length = value.Len()
	default:
		return nil
	}
	if length != expectedLen {
		return errors.New(translator("validation.length", label, params...))
	}
	return nil
}

// betweenValidator checks if a field value is between min and max.
func betweenValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) != 2 {
		return nil
	}
	min, err1 := strconv.ParseFloat(params[0], 64)
	max, err2 := strconv.ParseFloat(params[1], 64)
	if err1 != nil || err2 != nil {
		return nil
	}

	value := reflect.ValueOf(fieldValue)
	var val float64

	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val = float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val = float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		val = value.Float()
	case reflect.String:
		val = float64(len(value.String()))
	default:
		return nil
	}

	if val < min || val > max {
		return errors.New(translator("validation.between", label, params...))
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

// uuidValidator checks if a field is a valid UUID.
func uuidValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	uuidRegex := `[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}`
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	if matched := regexp.MustCompile("^" + uuidRegex + "$").MatchString(value); !matched {
		return errors.New(translator("validation.uuid", label, params...))
	}
	return nil
}

// creditCardValidator checks if a field is a valid credit card number using Luhn algorithm.
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
	for i := len(number) - 1; i > -1; i-- {
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

// equalValidator checks if a field is equal to parameter value.
func equalValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	value := fmt.Sprintf("%v", fieldValue)
	if value != params[0] {
		return errors.New(translator("validation.eq", label, params...))
	}
	return nil
}

// notEqualValidator checks if a field is not equal to parameter value.
// It returns an error if the field value is equal to the parameter value.
// It returns nil if the field value is not equal to the parameter value.
func notEqualValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	value := fmt.Sprintf("%v", fieldValue)
	if value == params[0] {
		return errors.New(translator("validation.ne", label, params...))
	}
	return nil
}

// lessThanValidator checks if a field is less than the parameter value.
// It returns an error if the field value is greater than or equal to the parameter value.
// It returns nil if the field value is less than the parameter value.
func lessThanValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	value, err := strconv.ParseFloat(fmt.Sprintf("%v", fieldValue), 64)
	if err != nil {
		return nil
	}
	threshold, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	if value >= threshold {
		return errors.New(translator("validation.lt", label, params...))
	}
	return nil
}

// lessThanOrEqualValidator checks if a field is less than or equal to the parameter value.
// It returns an error if the field value is greater than the parameter value.
// It returns nil if the field value is less than or equal to the parameter value.
func lessThanOrEqualValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	value, err := strconv.ParseFloat(fmt.Sprintf("%v", fieldValue), 64)
	if err != nil {
		return nil
	}
	threshold, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	if value > threshold {
		return errors.New(translator("validation.lte", label, params...))
	}
	return nil
}

// greaterThanValidator checks if a field is greater than the parameter value.
// It returns an error if the field value is less than or equal to the parameter value.
// It returns nil if the field value is greater than the parameter value.
func greaterThanValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	value, err := strconv.ParseFloat(fmt.Sprintf("%v", fieldValue), 64)
	if err != nil {
		return nil
	}
	threshold, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	if value <= threshold {
		return errors.New(translator("validation.gt", label, params...))
	}
	return nil
}

// greaterThanOrEqualValidator checks if a field is greater than or equal to the parameter value.
// It returns an error if the field value is less than the parameter value.
// It returns nil if the field value is greater than or equal to the parameter value.
func greaterThanOrEqualValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil
	}
	value, err := strconv.ParseFloat(fmt.Sprintf("%v", fieldValue), 64)
	if err != nil {
		return nil
	}
	threshold, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return nil
	}
	if value < threshold {
		return errors.New(translator("validation.gte", label, params...))
	}
	return nil
}

// passwordValidator checks if a field is a valid password.
// It returns an error if the field value is not a valid password.
// It returns nil if the field value is a valid password.
// Password must be at least 8 characters long and contain at least
// one uppercase letter, one lowercase letter, one number, and one special character.
func passwordValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	password := value.String()
	strings.TrimSpace(password)

	if len(password) < 8 {
		return errors.New(translator("validation.password", label, params...))
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

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

// phoneRegex is the regular expression for a valid phone number.
// Phone number must be a valid international phone number.
var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// phoneValidator checks if a field is a valid phone number in international format.
func phoneValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	if !phoneRegex.MatchString(value.String()) {
		return errors.New(translator("validation.phone", label, params...))
	}
	return nil
}

// usernameRegex is the regular expression for a valid username.
// Username must be at least 3 characters long and contain only alphanumeric characters and underscores.
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,}$`)

// usernameValidator checks if a field is a valid username.
// It returns an error if the field value is not a valid username.
// It returns nil if the field value is a valid username.
// Username must be at least 3 characters long and contain only alphanumeric characters and underscores.
func usernameValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	if len(value.String()) < 3 || !usernameRegex.MatchString(value.String()) {
		return errors.New(translator("validation.username", label, params...))
	}
	return nil
}

// slugValidator checks if a field is a valid slug.
// It returns an error if the field value is not a valid slug.
// It returns nil if the field value is a valid slug.
// Slug must be at least 3 characters long and contain only alphanumeric characters, underscores, and hyphens.
// Slug must not start or end with an underscore or hyphen.
// Slug must not contain consecutive underscores or hyphens.
// Slug must not contain spaces.
// Slug must not contain uppercase letters.
// Slug must not contain special characters.
// Slug must not be a reserved word.
func slugValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	slug := value.String()
	if len(slug) < 3 {
		return errors.New(translator("validation.slug", label, params...))
	}
	if slug[0] == '_' || slug[0] == '-' || slug[len(slug)-1] == '_' || slug[len(slug)-1] == '-' {
		return errors.New(translator("validation.slug", label, params...))
	}
	if strings.Contains(slug, "__") || strings.Contains(slug, "--") {
		return errors.New(translator("validation.slug", label, params...))
	}
	if strings.Contains(slug, " ") {
		return errors.New(translator("validation.slug", label, params...))
	}
	if strings.ToUpper(slug) != slug {
		return errors.New(translator("validation.slug", label, params...))
	}
	if !usernameRegex.MatchString(slug) {
		return errors.New(translator("validation.slug", label, params...))
	}
	if len(params) > 0 {
		reservedWords := strings.Split(params[0], ",")
		for _, reserved := range reservedWords {
			if slug == strings.TrimSpace(reserved) {
				return errors.New(translator("validation.slug", label, params...))
			}
		}
	}
	return nil
}

// hexcolorValidator checks if a field is a valid hexadecimal color code.
// It returns an error if the field value is not a valid hexadecimal color code.
// It returns nil if the field value is a valid hexadecimal color code.
// Hexadecimal color code must be 7 characters long and start with a hash symbol.
func hexcolorValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	hexcolor := value.String()
	if len(hexcolor) != 7 || hexcolor[0] != '#' {
		return errors.New(translator("validation.hexcolor", label, params...))
	}
	return nil
}

// fullnameValidator checks if a field is a valid full name.
// It returns an error if the field value is not a valid full name.
// It returns nil if the field value is a valid full name.
// Full name must be at least 3 characters long and contain only letters, spaces, and hyphens.
// Full name must not contain consecutive spaces or hyphens.
// Full name must not start or end with a space or hyphen.
// Full name must not contain special characters.
// Full name must not contain numbers.
func fullnameValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	fullname := value.String()
	if len(fullname) < 3 {
		return errors.New(translator("validation.fullname", label, params...))
	}
	if strings.Contains(fullname, "  ") || strings.Contains(fullname, "--") {
		return errors.New(translator("validation.fullname", label, params...))
	}
	if fullname[0] == ' ' || fullname[0] == '-' || fullname[len(fullname)-1] == ' ' || fullname[len(fullname)-1] == '-' {
		return errors.New(translator("validation.fullname", label, params...))
	}
	if strings.ContainsAny(fullname, "1234567890") {
		return errors.New(translator("validation.fullname", label, params...))
	}
	return nil
}

// nameRegex is the regular expression for a valid name.
// Name must be at least 2 characters long and contain only letters and spaces.
var nameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z ]*[a-zA-Z]$`)

// nameValidator checks if a field is a valid name.
// It returns an error if the field value is not a valid name.
// It returns nil if the field value is a valid name.
// Name must be at least 2 characters long and contain only letters and spaces.
func nameValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	name := value.String()

	// Check minimum length of 2 characters
	if len(name) < 2 {
		return errors.New(translator("validation.name", label, params...))
	}

	// Use regex to validate the name format
	if !nameRegex.MatchString(name) {
		return errors.New(translator("validation.name", label, params...))
	}

	return nil
}

// pastDateValidator checks if a date is in the past.
// It returns an error if the field value is not a date in the past.
// It returns nil if the field value is a date in the past.
func pastDateValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	var (
		t   time.Time
		err error
		ok  bool
	)

	// Check if the field value is a time.Time directly
	if t, ok = fieldValue.(time.Time); !ok {
		// If not, try to parse it as a string
		if s, ok := fieldValue.(string); ok && s != "" {
			// If format is provided in params
			if len(params) > 0 {
				t, err = time.Parse(params[0], s)
			} else {
				// Try standard formats
				t, err = time.Parse(time.RFC3339, s)
				if err != nil {
					t, err = time.Parse("2006-01-02", s)
				}
			}
			if err != nil {
				return errors.New(translator("validation.date", label, params...))
			}
		} else {
			return nil // Not a time.Time or string, ignore
		}
	}

	// Check if the date is in the past
	if !t.Before(time.Now()) {
		return errors.New(translator("validation.pastdate", label, params...))
	}

	return nil
}

// futureDateValidator checks if a date is in the future.
// It returns an error if the field value is not a date in the future.
// It returns nil if the field value is a date in the future.
func futureDateValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	var (
		t   time.Time
		err error
		ok  bool
	)

	// Check if the field value is a time.Time directly
	if t, ok = fieldValue.(time.Time); !ok {
		// If not, try to parse it as a string
		if s, ok := fieldValue.(string); ok && s != "" {
			// If format is provided in params
			if len(params) > 0 {
				t, err = time.Parse(params[0], s)
			} else {
				// Try standard formats
				t, err = time.Parse(time.RFC3339, s)
				if err != nil {
					t, err = time.Parse("2006-01-02", s)
				}
			}
			if err != nil {
				return errors.New(translator("validation.date", label, params...))
			}
		} else {
			return nil // Not a time.Time or string, ignore
		}
	}

	// Check if the date is in the future
	if !t.After(time.Now()) {
		return errors.New(translator("validation.futuredate", label, params...))
	}

	return nil
}

// workdayValidator checks if a date falls on a workday (Monday-Friday).
// It returns an error if the field value is not a workday.
// It returns nil if the field value is a workday.
func workdayValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	var (
		t   time.Time
		err error
		ok  bool
	)

	// Check if the field value is a time.Time directly
	if t, ok = fieldValue.(time.Time); !ok {
		// If not, try to parse it as a string
		if s, ok := fieldValue.(string); ok && s != "" {
			// If format is provided in params
			if len(params) > 0 {
				t, err = time.Parse(params[0], s)
			} else {
				// Try standard formats
				t, err = time.Parse(time.RFC3339, s)
				if err != nil {
					t, err = time.Parse("2006-01-02", s)
				}
			}
			if err != nil {
				return errors.New(translator("validation.date", label, params...))
			}
		} else {
			return nil // Not a time.Time or string, ignore
		}
	}

	// Check if the date is a workday (Monday-Friday)
	if weekday := t.Weekday(); weekday < 1 || weekday > 5 {
		return errors.New(translator("validation.workday", label, params...))
	}

	return nil
}

// weekendValidator checks if a date falls on a weekend (Saturday-Sunday).
// It returns an error if the field value is not a weekend day.
// It returns nil if the field value is a weekend day.
func weekendValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	var (
		t   time.Time
		err error
		ok  bool
	)

	// Check if the field value is a time.Time directly
	if t, ok = fieldValue.(time.Time); !ok {
		// If not, try to parse it as a string
		if s, ok := fieldValue.(string); ok && s != "" {
			// If format is provided in params
			if len(params) > 0 {
				t, err = time.Parse(params[0], s)
			} else {
				// Try standard formats
				t, err = time.Parse(time.RFC3339, s)
				if err != nil {
					t, err = time.Parse("2006-01-02", s)
				}
			}
			if err != nil {
				return errors.New(translator("validation.date", label, params...))
			}
		} else {
			return nil // Not a time.Time or string, ignore
		}
	}

	// Check if the date is a weekend day (Saturday-Sunday)
	if weekday := t.Weekday(); weekday != 0 && weekday != 6 {
		return errors.New(translator("validation.weekend", label, params...))
	}

	return nil
}

// ipv4Validator checks if a field is a valid IPv4 address.
// It returns an error if the field value is not a valid IPv4 address.
// It returns nil if the field value is a valid IPv4 address.
func ipv4Validator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() == nil {
		return errors.New(translator("validation.ipv4", label, params...))
	}
	
	return nil
}

// ipv6Validator checks if a field is a valid IPv6 address.
// It returns an error if the field value is not a valid IPv6 address.
// It returns nil if the field value is a valid IPv6 address.
func ipv6Validator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	ip := net.ParseIP(value)
	if ip == nil || ip.To4() != nil {
		return errors.New(translator("validation.ipv6", label, params...))
	}
	
	return nil
}

// domainRegex is a regular expression for validating domain names.
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]$`)

// domainValidator checks if a field is a valid domain name.
// It returns an error if the field value is not a valid domain name.
// It returns nil if the field value is a valid domain name.
func domainValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	if !domainRegex.MatchString(value) {
		return errors.New(translator("validation.domain", label, params...))
	}
	
	return nil
}

// macRegex is a regular expression for validating MAC addresses.
var macRegex = regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)

// macValidator checks if a field is a valid MAC address.
// It returns an error if the field value is not a valid MAC address.
// It returns nil if the field value is a valid MAC address.
func macValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	if !macRegex.MatchString(value) {
		return errors.New(translator("validation.mac", label, params...))
	}
	
	return nil
}

// portValidator checks if a field is a valid port number (0-65535).
// It returns an error if the field value is not a valid port number.
// It returns nil if the field value is a valid port number.
func portValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	var port int64
	
	value := reflect.ValueOf(fieldValue)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		port = value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		port = int64(value.Uint())
	case reflect.Float32, reflect.Float64:
		port = int64(value.Float())
	case reflect.String:
		var err error
		port, err = strconv.ParseInt(value.String(), 10, 64)
		if err != nil {
			return errors.New(translator("validation.port", label, params...))
		}
	default:
		return nil // Not a numeric type, ignore
	}
	
	if port < 0 || port > 65535 {
		return errors.New(translator("validation.port", label, params...))
	}
	
	return nil
}

// startsWithValidator checks if a field starts with a specific prefix.
// It returns an error if the field value does not start with the specified prefix.
// It returns nil if the field value starts with the specified prefix.
// The prefix is provided as the first parameter.
func startsWithValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No prefix specified
	}
	
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	
	prefix := params[0]
	if !strings.HasPrefix(value, prefix) {
		return errors.New(translator("validation.startswith", label, params...))
	}
	
	return nil
}

// endsWithValidator checks if a field ends with a specific suffix.
// It returns an error if the field value does not end with the specified suffix.
// It returns nil if the field value ends with the specified suffix.
// The suffix is provided as the first parameter.
func endsWithValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No suffix specified
	}
	
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	
	suffix := params[0]
	if !strings.HasSuffix(value, suffix) {
		return errors.New(translator("validation.endswith", label, params...))
	}
	
	return nil
}

// containsValidator checks if a field contains a specific substring.
// It returns an error if the field value does not contain the specified substring.
// It returns nil if the field value contains the specified substring.
// The substring is provided as the first parameter.
func containsValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No substring specified
	}
	
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	
	substring := params[0]
	if !strings.Contains(value, substring) {
		return errors.New(translator("validation.contains", label, params...))
	}
	
	return nil
}

// notContainsValidator checks if a field does not contain a specific substring.
// It returns an error if the field value contains the specified substring.
// It returns nil if the field value does not contain the specified substring.
// The substring is provided as the first parameter.
func notContainsValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No substring specified
	}
	
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	
	substring := params[0]
	if strings.Contains(value, substring) {
		return errors.New(translator("validation.notcontains", label, params...))
	}
	
	return nil
}

// asciiRegex is a regular expression for validating ASCII characters.
var asciiRegex = regexp.MustCompile(`^[\x00-\x7F]*$`)

// asciiValidator checks if a field contains only ASCII characters.
// It returns an error if the field value contains non-ASCII characters.
// It returns nil if the field value contains only ASCII characters.
func asciiValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok {
		return nil
	}
	
	if !asciiRegex.MatchString(value) {
		return errors.New(translator("validation.ascii", label, params...))
	}
	
	return nil
}

// positiveValidator checks if a field is a positive number.
// It returns an error if the field value is not a positive number.
// It returns nil if the field value is a positive number.
func positiveValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() <= 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Uint() == 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		if value.Float() <= 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	case reflect.String:
		num, err := strconv.ParseFloat(value.String(), 64)
		if err != nil || num <= 0 {
			return errors.New(translator("validation.positive", label, params...))
		}
	default:
		return nil // Not a numeric type, ignore
	}
	
	return nil
}

// negativeValidator checks if a field is a negative number.
// It returns an error if the field value is not a negative number.
// It returns nil if the field value is a negative number.
func negativeValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() >= 0 {
			return errors.New(translator("validation.negative", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// Unsigned integers can't be negative
		return errors.New(translator("validation.negative", label, params...))
	case reflect.Float32, reflect.Float64:
		if value.Float() >= 0 {
			return errors.New(translator("validation.negative", label, params...))
		}
	case reflect.String:
		num, err := strconv.ParseFloat(value.String(), 64)
		if err != nil || num >= 0 {
			return errors.New(translator("validation.negative", label, params...))
		}
	default:
		return nil // Not a numeric type, ignore
	}
	
	return nil
}

// evenValidator checks if a field is an even number.
// It returns an error if the field value is not an even number.
// It returns nil if the field value is an even number.
func evenValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int()%2 != 0 {
			return errors.New(translator("validation.even", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Uint()%2 != 0 {
			return errors.New(translator("validation.even", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		// For floating-point, check if it's an integer and even
		f := value.Float()
		if f != float64(int64(f)) || int64(f)%2 != 0 {
			return errors.New(translator("validation.even", label, params...))
		}
	case reflect.String:
		num, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil || num%2 != 0 {
			return errors.New(translator("validation.even", label, params...))
		}
	default:
		return nil // Not a numeric type, ignore
	}
	
	return nil
}

// oddValidator checks if a field is an odd number.
// It returns an error if the field value is not an odd number.
// It returns nil if the field value is an odd number.
func oddValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value := reflect.ValueOf(fieldValue)
	
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int()%2 == 0 {
			return errors.New(translator("validation.odd", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Uint()%2 == 0 {
			return errors.New(translator("validation.odd", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		// For floating-point, check if it's an integer and odd
		f := value.Float()
		if f != float64(int64(f)) || int64(f)%2 == 0 {
			return errors.New(translator("validation.odd", label, params...))
		}
	case reflect.String:
		num, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil || num%2 == 0 {
			return errors.New(translator("validation.odd", label, params...))
		}
	default:
		return nil // Not a numeric type, ignore
	}
	
	return nil
}

// multipleValidator checks if a field is a multiple of a specified number.
// It returns an error if the field value is not a multiple of the specified number.
// It returns nil if the field value is a multiple of the specified number.
// The divisor is provided as the first parameter.
func multipleValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No divisor specified
	}
	
	divisor, err := strconv.ParseInt(params[0], 10, 64)
	if err != nil || divisor == 0 {
		return nil // Invalid divisor
	}
	
	value := reflect.ValueOf(fieldValue)
	
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int()%divisor != 0 {
			return errors.New(translator("validation.multiple", label, params...))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Uint()%uint64(divisor) != 0 {
			return errors.New(translator("validation.multiple", label, params...))
		}
	case reflect.Float32, reflect.Float64:
		// For floating-point, check if it's an integer multiple
		f := value.Float()
		if f != float64(int64(f)) || int64(f)%divisor != 0 {
			return errors.New(translator("validation.multiple", label, params...))
		}
	case reflect.String:
		num, err := strconv.ParseInt(value.String(), 10, 64)
		if err != nil || num%divisor != 0 {
			return errors.New(translator("validation.multiple", label, params...))
		}
	default:
		return nil // Not a numeric type, ignore
	}
	
	return nil
}

// base64Validator checks if a field is a valid Base64 encoded string.
// It returns an error if the field value is not a valid Base64 encoded string.
// It returns nil if the field value is a valid Base64 encoded string.
func base64Validator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	// Check if the string has valid Base64 length (multiple of 4)
	if len(value)%4 != 0 {
		return errors.New(translator("validation.base64", label, params...))
	}
	
	// Try to decode the string using standard encoding
	_, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		// Try URL encoding as fallback
		_, err = base64.URLEncoding.DecodeString(value)
		if err != nil {
			return errors.New(translator("validation.base64", label, params...))
		}
	}
	
	return nil
}

// jsonValidator checks if a field is a valid JSON string.
// It returns an error if the field value is not a valid JSON string.
// It returns nil if the field value is a valid JSON string.
func jsonValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	var js interface{}
	err := json.Unmarshal([]byte(value), &js)
	if err != nil {
		return errors.New(translator("validation.json", label, params...))
	}
	
	return nil
}

// semverRegex is a regular expression for validating semantic version strings.
var semverRegex = regexp.MustCompile(`^v?([0-9]+)\.([0-9]+)\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

// semverValidator checks if a field is a valid semantic version string.
// It returns an error if the field value is not a valid semantic version string.
// It returns nil if the field value is a valid semantic version string.
func semverValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	if !semverRegex.MatchString(value) {
		return errors.New(translator("validation.semver", label, params...))
	}
	
	return nil
}

// extensionValidator checks if a file has one of the allowed extensions.
// It returns an error if the field value does not have one of the allowed extensions.
// It returns nil if the field value has one of the allowed extensions.
// The allowed extensions are provided as parameters.
func extensionValidator(fieldValue any, fieldType reflect.StructField, params []string, label string, translator ErrorTranslatorFunc) error {
	if len(params) == 0 {
		return nil // No extensions specified
	}
	
	value, ok := fieldValue.(string)
	if !ok || value == "" {
		return nil
	}
	
	// Extract the extension (everything after the last dot)
	ext := strings.ToLower(filepath.Ext(value))
	if ext == "" {
		return errors.New(translator("validation.extension", label, params...))
	}
	
	// Remove the leading dot
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	
	// Check if the extension is in the allowed list
	for _, allowedExt := range params {
		// Remove leading dot from allowed extension if present
		if len(allowedExt) > 0 && allowedExt[0] == '.' {
			allowedExt = allowedExt[1:]
		}
		
		if strings.EqualFold(ext, allowedExt) {
			return nil
		}
	}
	
	return errors.New(translator("validation.extension", label, params...))
}
