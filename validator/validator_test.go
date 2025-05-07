package validator_test

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/dmitrymomot/gokit/validator"
	"github.com/stretchr/testify/require"
)

func TestNewValidator_DefaultOptions(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	require.NotNil(t, v)
	require.NoError(t, v.ValidateStruct(struct{}{}))
	require.NoError(t, v.ValidateStruct(&struct{}{}))
}

func TestNewValidator_WithInvalidSeparators(t *testing.T) {
	t.Parallel()
	_, err := validator.New(validator.WithSeparators("", ":", ","))
	require.Error(t, err)
	require.Equal(t, validator.ErrInvalidSeparatorConfiguration, err)
}

func TestRegisterValidation_InvalidInputs(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	err = v.RegisterValidation("", nil)
	require.Error(t, err)
	require.Equal(t, validator.ErrInvalidValidatorConfiguration, err)
}

func TestRegisterValidation_CustomValidation(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		return nil
	}
	require.NoError(t, v.RegisterValidation("foo", dummy))
	type T struct {
		F string `validate:"foo"`
	}
	err = v.ValidateStruct(T{F: "bar"})
	require.NoError(t, err)
}

func TestValidateStruct_UnknownRule(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	type T struct {
		F string `validate:"unknown"`
	}
	err = v.ValidateStruct(T{F: "bar"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown rule 'unknown'")
}

func TestValidateStruct_CustomValidatorError(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		return errors.New("myerror")
	}
	require.NoError(t, v.RegisterValidation("bar", dummy))
	type T struct {
		X int `validate:"bar"`
	}
	err = v.ValidateStruct(T{X: 1})
	require.Error(t, err)
	require.Contains(t, err.Error(), "myerror")
}

func TestValidateStruct_NestedStruct(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		return errors.New("nested")
	}
	require.NoError(t, v.RegisterValidation("foo", dummy))
	type Inner struct {
		A any `validate:"foo"`
	}
	type Outer struct {
		InnerField Inner
	}
	err = v.ValidateStruct(Outer{InnerField: Inner{A: nil}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "nested")
	require.Contains(t, err.Error(), "A")
}

func TestValidateStruct_OmitemptySkipping(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	called := false
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		called = true
		return errors.New("should not be called")
	}
	require.NoError(t, v.RegisterValidation("test", dummy))
	type T struct {
		A string `validate:"omitempty;test"`
	}
	err = v.ValidateStruct(T{A: ""})
	require.NoError(t, err)
	require.False(t, called)
}

// Test that parseRule correctly splits parameters and passes label
func TestParseRule_ParamsAndLabel(t *testing.T) {
	t.Parallel()
	var gotParams []string
	var gotLabel string
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		gotParams = params
		gotLabel = label
		return nil
	}
	// Register custom validator using option
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type T struct {
		B string `validate:"foo:1,2,3" label:"MyField"`
	}
	err = v.ValidateStruct(T{B: "value"})
	require.NoError(t, err)
	require.Equal(t, []string{"1", "2", "3"}, gotParams)
	require.Equal(t, "MyField", gotLabel)
}

// Test omitempty skipping for nil pointer using isZero
func TestOmitemptySkipping_NilPointer(t *testing.T) {
	t.Parallel()
	called := false
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		called = true
		return nil
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type T struct {
		P *int `validate:"omitempty;foo" label:"Ptr"`
	}
	var p *int
	err = v.ValidateStruct(T{P: p})
	require.NoError(t, err)
	require.False(t, called)
}

// Validate non-struct input returns empty ValidationErrors and not IsValidationError
func TestValidateStruct_NonStruct(t *testing.T) {
	t.Parallel()
	v, err := validator.New()
	require.NoError(t, err)
	errTest := v.ValidateStruct(123)
	require.Error(t, errTest)
	require.False(t, validator.IsValidationError(errTest))
	require.Equal(t, "map[]", errTest.Error())
}

// Test custom error translator via WithErrorTranslator
func TestWithErrorTranslator(t *testing.T) {
	t.Parallel()
	// custom translator
	trans := func(key, label string, params ...string) string {
		return fmt.Sprintf("T:%s|%s", key, label)
	}
	// dummy validator emits translation
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		return errors.New(translator("foo", label))
	}
	v, err := validator.New(validator.WithErrorTranslator(trans), validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type X struct {
		A string `validate:"foo" label:"L"`
	}
	err2 := v.ValidateStruct(X{A: ""})
	require.Error(t, err2)
	require.Contains(t, err2.Error(), "T:foo|L")
}

// Test omitempty skipping for slices, maps, and bools via isZero
func TestOmitemptySkipping_MultiTypes(t *testing.T) {
	t.Parallel()
	called := false
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		called = true
		return errors.New("dummy")
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)

	type S struct {
		Arr []int          `validate:"omitempty;foo"`
		M   map[string]int `validate:"omitempty;foo"`
		B   bool           `validate:"omitempty;foo"`
	}
	// zero values skip
	err = v.ValidateStruct(S{Arr: nil, M: nil, B: false})
	require.NoError(t, err)
	require.False(t, called)
	// non-zero values call
	called = false
	err = v.ValidateStruct(S{Arr: []int{1}, M: map[string]int{"a": 1}, B: true})
	require.Error(t, err)
	require.True(t, called)
}

// Test regex rule parameter special-case in parseRule
func TestParseRule_RegexSpecial(t *testing.T) {
	t.Parallel()
	var gotParams []string
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		gotParams = params
		return nil
	}
	v, err := validator.New(validator.WithCustomValidator("regex", dummy))
	require.NoError(t, err)
	type T struct {
		X string `validate:"regex:^foo:bar,baz,qux" label:"X"`
	}
	err = v.ValidateStruct(T{X: "anything"})
	require.NoError(t, err)
	require.Equal(t, []string{"^foo:bar,baz,qux"}, gotParams)
}

// Test parseRule with no parameters
func TestParseRule_NoParams(t *testing.T) {
	t.Parallel()
	var gotParams []string
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		gotParams = params
		return nil
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type T struct {
		X string `validate:"foo" label:"L"`
	}
	err = v.ValidateStruct(T{X: "val"})
	require.NoError(t, err)
	require.Nil(t, gotParams)
}

// Test omitempty skipping for int type via isZero
func TestOmitemptySkipping_Int(t *testing.T) {
	t.Parallel()
	called := false
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		called = true
		return errors.New("dummy")
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type S struct {
		N int `validate:"omitempty;foo" label:"N"`
	}
	err = v.ValidateStruct(S{N: 0})
	require.NoError(t, err)
	require.False(t, called)
	called = false
	err = v.ValidateStruct(S{N: 5})
	require.Error(t, err)
	require.True(t, called)
}

// Test parseRule multiple paramSeparator in rule name and join
func TestParseRule_MultiParamSeparator(t *testing.T) {
	t.Parallel()
	var gotParams []string
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		gotParams = params
		return nil
	}
	// Explicitly set default separators to ensure a clean state for this test
	v, err := validator.New(
		validator.WithSeparators(";", ":", ","),
		validator.WithCustomValidator("foo", dummy),
	)
	require.NoError(t, err)
	type T struct {
		X string `validate:"foo:bar:baz"`
	}
	err = v.ValidateStruct(T{X: "anything"})
	require.NoError(t, err, "ValidateStruct returned an unexpected error")

	// Log gotParams before assertion
	t.Logf("TestParseRule_MultiParamSeparator: Rule string 'foo:bar:baz'")
	t.Logf("TestParseRule_MultiParamSeparator: Expected params: []string{\"bar:baz\"}")
	t.Logf("TestParseRule_MultiParamSeparator: Got params: %#v", gotParams)

	require.Equal(t, []string{"bar:baz"}, gotParams)
}

// Test custom separators option
func TestWithCustomSeparators(t *testing.T) {
	t.Parallel()
	called := false
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		called = true
		return errors.New("err")
	}
	v, err := validator.New(validator.WithSeparators("|", "#", ";"), validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type T struct {
		A string `validate:"omitempty|foo#1;2;3"`
	}
	// empty value should skip
	err = v.ValidateStruct(T{A: ""})
	require.NoError(t, err)
	require.False(t, called)
	// non-empty value triggers
	err = v.ValidateStruct(T{A: "val"})
	require.Error(t, err)
	require.True(t, called)
}

// Test omitempty skipping for float and uint types via isZero
func TestOmitemptySkipping_FloatUint(t *testing.T) {
	t.Parallel()
	called := false
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		called = true
		return errors.New("dummy")
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type S struct {
		F float64 `validate:"omitempty;foo"`
		U uint    `validate:"omitempty;foo"`
	}
	// zero values skip
	err = v.ValidateStruct(S{F: 0.0, U: 0})
	require.NoError(t, err)
	require.False(t, called)
	// non-zero values call
	called = false
	err = v.ValidateStruct(S{F: 1.23, U: 42})
	require.Error(t, err)
	require.True(t, called)
}

// Test ExtractValidationErrors and Values
func TestExtractValidationErrors_AndValues(t *testing.T) {
	t.Parallel()
	ve := validator.NewValidationError("F", "e")
	err := validator.ValidationErrors(ve)
	extracted := validator.ExtractValidationErrors(err)
	require.NotNil(t, extracted)
	require.Equal(t, url.Values(ve), extracted.Values())
	require.True(t, validator.IsValidationError(err))
}

// Test WithValidators loads only specified built-in validators
func TestWithValidators_Specific(t *testing.T) {
	t.Parallel()
	v, err := validator.New(validator.WithValidators("numeric"))
	require.NoError(t, err)
	type T struct {
		A string `validate:"numeric" label:"A"`
	}
	err = v.ValidateStruct(T{A: "123"})
	require.NoError(t, err)
	err = v.ValidateStruct(T{A: "abc"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "numeric")
	type U struct {
		B string `validate:"email"`
	}
	err = v.ValidateStruct(U{B: "bad"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown rule 'email'")
}

// Test WithAllValidators loads all built-in validators
func TestWithAllValidators(t *testing.T) {
	t.Parallel()
	v, err := validator.New(validator.WithAllValidators())
	require.NoError(t, err)
	type T struct {
		E string `validate:"email" label:"E"`
	}
	err = v.ValidateStruct(T{E: "bad"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "email")
}

// Test WithExcept excludes specified validators
func TestWithExcept(t *testing.T) {
	t.Parallel()
	v, err := validator.New(validator.WithAllValidators(), validator.WithExcept("email"))
	require.NoError(t, err)
	type T struct {
		E string `validate:"email"`
	}
	err = v.ValidateStruct(T{E: "bad"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown rule 'email'")
}

// Test WithCustomValidator invalid configuration via New option
func TestNewWithCustomValidator_Invalid(t *testing.T) {
	t.Parallel()
	_, err := validator.New(validator.WithCustomValidator("", nil))
	require.Error(t, err)
	require.Equal(t, validator.ErrInvalidValidatorConfiguration, err)
}

// Test label fallback when label tag is missing
func TestLabelFallback_DefaultLabel(t *testing.T) {
	t.Parallel()
	var gotLabel string
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		gotLabel = label
		return errors.New("err")
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type T struct {
		X string `validate:"foo"`
	}
	err = v.ValidateStruct(T{X: "val"})
	require.Error(t, err)
	require.Equal(t, "X", gotLabel)
}

// Test multiple errors accumulate for a single field
func TestMultipleErrorsPerField(t *testing.T) {
	t.Parallel()
	v, err := validator.New(validator.WithAllValidators())
	require.NoError(t, err)
	type T struct {
		N int `validate:"min:2;max:0" label:"N"`
	}
	err = v.ValidateStruct(T{N: 1})
	require.Error(t, err)
	ve := validator.ExtractValidationErrors(err)
	vals := ve.Values()
	errList := vals["N"]
	require.Len(t, errList, 2)
	require.Contains(t, errList, "validation.min")
	require.Contains(t, errList, "validation.max")
}

// Test nested struct prefix propagation to field names
func TestValidateStruct_NestedPrefixPropagation(t *testing.T) {
	t.Parallel()
	dummy := func(fieldValue any, fieldType reflect.StructField, params []string, label string, translator validator.ErrorTranslatorFunc) error {
		return errors.New("err")
	}
	v, err := validator.New(validator.WithCustomValidator("foo", dummy))
	require.NoError(t, err)
	type Inner struct {
		X string `validate:"foo" label:"InnerX"`
	}
	type Outer struct {
		InnerField Inner
	}
	err = v.ValidateStruct(Outer{InnerField: Inner{X: "val"}})
	require.Error(t, err)
	ve := validator.ExtractValidationErrors(err)
	vals := ve.Values()
	require.Len(t, vals, 1)
	require.Contains(t, vals, "InnerField.X")
}
