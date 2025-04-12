package sanitizer_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dmitrymomot/gokit/sanitizer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	TrimField       string `sanitize:"trim"`
	LowerField      string `sanitize:"lower"`
	UpperField      string `sanitize:"upper"`
	ReplaceField    string `sanitize:"replace:old:new"`
	StripHTMLField  string `sanitize:"striphtml"`
	EscapeField     string `sanitize:"escape"`
	AlphanumField   string `sanitize:"alphanum"`
	NumericField    string `sanitize:"numeric"`
	TruncateField   string `sanitize:"truncate:5"`
	NormalizeField  string `sanitize:"normalize"`
	CapitalizeField string `sanitize:"capitalize"`
	CamelCaseField  string `sanitize:"camelcase"`
	SnakeCaseField  string `sanitize:"snakecase"`
	KebabCaseField  string `sanitize:"kebabcase"`
	UCFirstField    string `sanitize:"ucfirst"`
	MultipleRules   string `sanitize:"trim,lower,replace:hello:hi"`
	NoTag           string
	unexportedField string `sanitize:"trim"`
}

func TestSanitizeStruct(t *testing.T) {
	t.Run("TrimField", func(t *testing.T) {
		input := &TestStruct{TrimField: "  hello  "}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello", input.TrimField)
	})

	t.Run("LowerField", func(t *testing.T) {
		input := &TestStruct{LowerField: "HELLO"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello", input.LowerField)
	})

	t.Run("UpperField", func(t *testing.T) {
		input := &TestStruct{UpperField: "hello"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "HELLO", input.UpperField)
	})

	t.Run("ReplaceField", func(t *testing.T) {
		input := &TestStruct{ReplaceField: "hello old world"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello new world", input.ReplaceField)
	})

	t.Run("StripHTMLField", func(t *testing.T) {
		input := &TestStruct{StripHTMLField: "<p>hello</p>"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello", input.StripHTMLField)
	})

	t.Run("EscapeField", func(t *testing.T) {
		input := &TestStruct{EscapeField: "<hello>&world"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "&lt;hello&gt;&amp;world", input.EscapeField)
	})

	t.Run("AlphanumField", func(t *testing.T) {
		input := &TestStruct{AlphanumField: "hello123!@#"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello123", input.AlphanumField)
	})

	t.Run("NumericField", func(t *testing.T) {
		input := &TestStruct{NumericField: "abc123def456"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "123456", input.NumericField)
	})

	t.Run("TruncateField", func(t *testing.T) {
		input := &TestStruct{TruncateField: "hello world"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello", input.TruncateField)
	})

	t.Run("NormalizeField", func(t *testing.T) {
		input := &TestStruct{NormalizeField: "hello\r\nworld"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello\nworld", input.NormalizeField)
	})

	t.Run("MultipleRules", func(t *testing.T) {
		input := &TestStruct{MultipleRules: "  HELLO World  "}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hi world", input.MultipleRules)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		input := &TestStruct{}
		sanitizer.SanitizeStruct(input)
		assert.Empty(t, input.TrimField)
		assert.Empty(t, input.LowerField)
		assert.Empty(t, input.UpperField)
	})

	t.Run("UnexportedField", func(t *testing.T) {
		input := &TestStruct{unexportedField: "  hello  "}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "  hello  ", input.unexportedField)
	})

	t.Run("NoTag", func(t *testing.T) {
		input := &TestStruct{NoTag: "  hello  "}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "  hello  ", input.NoTag)
	})
}

func TestRegisterSanitizer(t *testing.T) {
	type CustomStruct struct {
		Field string `sanitize:"custom"`
	}

	// Reset before all test cases
	sanitizer.ResetSanitizers()

	t.Run("Register and use custom sanitizer", func(t *testing.T) {
		sanitizer.ResetSanitizers() // Reset before each test case
		sanitizer.RegisterSanitizer("custom", func(fieldValue any, fieldType reflect.StructField, params []string) any {
			if v, ok := fieldValue.(string); ok {
				return strings.ToUpper(v) + "!"
			}
			return fieldValue
		})

		input := &CustomStruct{Field: "hello"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "HELLO!", input.Field)
	})

	t.Run("Register nil sanitizer", func(t *testing.T) {
		sanitizer.ResetSanitizers() // Reset before each test case
		sanitizer.RegisterSanitizer("nil", nil)
		input := &CustomStruct{Field: "hello"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello", input.Field)
	})

	t.Run("Register empty tag", func(t *testing.T) {
		sanitizer.ResetSanitizers() // Reset before each test case
		sanitizer.RegisterSanitizer("", func(fieldValue any, fieldType reflect.StructField, params []string) any {
			return fieldValue
		})
		input := &CustomStruct{Field: "hello"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello", input.Field)
	})
}

func TestInvalidInputs(t *testing.T) {
	sanitizer.ResetSanitizers() // Reset before each test case

	t.Run("Nil pointer", func(t *testing.T) {
		require.NotPanics(t, func() {
			sanitizer.SanitizeStruct(nil)
		})
	})

	t.Run("Non-pointer", func(t *testing.T) {
		require.NotPanics(t, func() {
			sanitizer.SanitizeStruct(TestStruct{})
		})
	})

	t.Run("Pointer to non-struct", func(t *testing.T) {
		str := "hello"
		require.NotPanics(t, func() {
			sanitizer.SanitizeStruct(&str)
		})
	})
}

func TestCaseConversions(t *testing.T) {
	t.Run("CamelCase", func(t *testing.T) {
		input := &TestStruct{CamelCaseField: "hello world"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "helloWorld", input.CamelCaseField)
	})

	t.Run("SnakeCase", func(t *testing.T) {
		input := &TestStruct{SnakeCaseField: "helloWorld"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello_world", input.SnakeCaseField)
	})

	t.Run("KebabCase", func(t *testing.T) {
		input := &TestStruct{KebabCaseField: "helloWorld"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "hello-world", input.KebabCaseField)
	})

	t.Run("UCFirst", func(t *testing.T) {
		input := &TestStruct{UCFirstField: "hello"}
		sanitizer.SanitizeStruct(input)
		assert.Equal(t, "Hello", input.UCFirstField)
	})
}
