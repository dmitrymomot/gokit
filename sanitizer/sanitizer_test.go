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
	ReplaceField    string `sanitize:"replace:old,new"`
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
	MultipleRules   string `sanitize:"trim;lower;replace:hello,hi"`
	NoTag           string
	unexportedField string `sanitize:"trim"`
}

func TestSanitizeStruct(t *testing.T) {
	// Create a new sanitizer instance with default configuration
	s, err := sanitizer.New()
	require.NoError(t, err)

	t.Run("TrimField", func(t *testing.T) {
		input := &TestStruct{TrimField: "  hello  "}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello", input.TrimField)
	})

	t.Run("LowerField", func(t *testing.T) {
		input := &TestStruct{LowerField: "HELLO"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello", input.LowerField)
	})

	t.Run("UpperField", func(t *testing.T) {
		input := &TestStruct{UpperField: "hello"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "HELLO", input.UpperField)
	})

	t.Run("ReplaceField", func(t *testing.T) {
		input := &TestStruct{ReplaceField: "hello old world"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		// The replace rule is "replace:old,new", so "old" should be replaced with "new"
		assert.Equal(t, "hello new world", input.ReplaceField)
	})

	t.Run("StripHTMLField", func(t *testing.T) {
		input := &TestStruct{StripHTMLField: "<p>hello</p>"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello", input.StripHTMLField)
	})

	t.Run("EscapeField", func(t *testing.T) {
		input := &TestStruct{EscapeField: "<hello>&world"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "&lt;hello&gt;&amp;world", input.EscapeField)
	})

	t.Run("AlphanumField", func(t *testing.T) {
		input := &TestStruct{AlphanumField: "hello123!@#"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello123", input.AlphanumField)
	})

	t.Run("NumericField", func(t *testing.T) {
		input := &TestStruct{NumericField: "abc123def456"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "123456", input.NumericField)
	})

	t.Run("TruncateField", func(t *testing.T) {
		input := &TestStruct{TruncateField: "hello world"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello", input.TruncateField)
	})

	t.Run("NormalizeField", func(t *testing.T) {
		input := &TestStruct{NormalizeField: "hello\r\nworld"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello\nworld", input.NormalizeField)
	})

	t.Run("MultipleRules", func(t *testing.T) {
		input := &TestStruct{MultipleRules: "  HELLO World  "}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		// The rules are "trim;lower;replace:hello:hi"
		// 1. trim: "  HELLO World  " -> "HELLO World"
		// 2. lower: "HELLO World" -> "hello world"
		// 3. replace:hello:hi: "hello world" -> "hi world"
		assert.Equal(t, "hi world", input.MultipleRules)
	})

	t.Run("EmptyFields", func(t *testing.T) {
		input := &TestStruct{}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Empty(t, input.TrimField)
		assert.Empty(t, input.LowerField)
		assert.Empty(t, input.UpperField)
	})

	t.Run("UnexportedField", func(t *testing.T) {
		input := &TestStruct{unexportedField: "  hello  "}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "  hello  ", input.unexportedField)
	})

	t.Run("NoTag", func(t *testing.T) {
		input := &TestStruct{NoTag: "  hello  "}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "  hello  ", input.NoTag)
	})

	t.Run("Instance sanitizer behavior", func(t *testing.T) {
		// Create a new sanitizer instance with default configuration
		s, err := sanitizer.New()
		require.NoError(t, err)

		// Define a custom struct for this test
		type TestStructCustom struct {
			TrimField     string `sanitize:"trim"`
			LowerField    string `sanitize:"lower"`
			UpperField    string `sanitize:"upper"`
			MultipleRules string `sanitize:"trim;lower;replace:hello,hi"`
		}

		input := &TestStructCustom{
			TrimField:     "  hello  ",
			LowerField:    "HELLO",
			UpperField:    "hello",
			MultipleRules: "  hello world  ",
		}
		err = s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello", input.TrimField)
		assert.Equal(t, "hello", input.LowerField)
		assert.Equal(t, "HELLO", input.UpperField)
		// The rules are "trim;lower;replace:hello:hi"
		// 1. trim: "  hello world  " -> "hello world"
		// 2. lower: "hello world" -> "hello world" (no change)
		// 3. replace:hello:hi: "hello world" -> "hi world"
		assert.Equal(t, "hi world", input.MultipleRules)
	})
}

func TestRegisterSanitizer(t *testing.T) {
	type CustomStruct struct {
		Field string `sanitize:"custom"`
	}

	t.Run("Register and use custom sanitizer", func(t *testing.T) {
		// Create a new sanitizer instance for isolation
		s, err := sanitizer.New()
		require.NoError(t, err)

		err = s.RegisterSanitizer("custom", func(fieldValue any, fieldType reflect.StructField, params []string) any {
			if v, ok := fieldValue.(string); ok {
				return strings.ToUpper(v) + "!"
			}
			return fieldValue
		})
		require.NoError(t, err)

		input := &CustomStruct{Field: "hello"}
		err = s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "HELLO!", input.Field)
	})

	t.Run("Register nil sanitizer", func(t *testing.T) {
		// Create a new sanitizer instance for isolation
		s, err := sanitizer.New()
		require.NoError(t, err)

		err = s.RegisterSanitizer("nil", nil)
		require.Error(t, err) // Should now return an error instead of silently ignoring
		assert.ErrorIs(t, err, sanitizer.ErrInvalidSanitizerConfiguration)
	})

	t.Run("Register empty tag", func(t *testing.T) {
		// Create a new sanitizer instance for isolation
		s, err := sanitizer.New()
		require.NoError(t, err)

		err = s.RegisterSanitizer("", func(any, reflect.StructField, []string) any { return nil })
		require.Error(t, err)
		assert.ErrorIs(t, err, sanitizer.ErrInvalidSanitizerConfiguration)
	})

	t.Run("Register duplicate tag", func(t *testing.T) {
		// Create a new sanitizer instance for isolation
		s, err := sanitizer.New()
		require.NoError(t, err)

		err = s.RegisterSanitizer("dupe", func(any, reflect.StructField, []string) any { return nil })
		require.NoError(t, err)

		err = s.RegisterSanitizer("dupe", func(any, reflect.StructField, []string) any { return nil })
		require.NoError(t, err) // Should not return an error, just override
	})

	t.Run("DefaultSanitizer behavior", func(t *testing.T) {
		// Create a new sanitizer instance for isolation
		s, err := sanitizer.New()
		require.NoError(t, err)

		err = s.RegisterSanitizer("custom", func(fieldValue any, fieldType reflect.StructField, params []string) any {
			if v, ok := fieldValue.(string); ok {
				return strings.ToUpper(v) + "!"
			}
			return fieldValue
		})
		require.NoError(t, err)

		input := &CustomStruct{Field: "hello"}
		err = s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "HELLO!", input.Field)
	})
}

func TestInvalidInputs(t *testing.T) {
	s, err := sanitizer.New()
	require.NoError(t, err)

	t.Run("Nil input", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = s.SanitizeStruct(nil)
		})
	})

	t.Run("Non-pointer input", func(t *testing.T) {
		testStruct := TestStruct{}
		err := s.SanitizeStruct(testStruct)
		assert.NoError(t, err) // Should not return an error, just do nothing
	})

	t.Run("Non-struct pointer", func(t *testing.T) {
		var str string = "test"
		err := s.SanitizeStruct(&str)
		assert.NoError(t, err) // Should not return an error, just do nothing
	})
}

func TestSlugSanitizer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with special characters",
			input:    "Hello @#$%^&*()_+World!!!",
			expected: "hello-world",
		},
		{
			name:     "with diacritics",
			input:    "Café au Lait",
			expected: "cafe-au-lait",
		},
		{
			name:     "with multiple spaces",
			input:    "Hello    World",
			expected: "hello-world",
		},
		{
			name:     "with leading/trailing spaces",
			input:    "  Hello World  ",
			expected: "hello-world",
		},
		{
			name:     "with mixed case",
			input:    "HeLLo WoRLd",
			expected: "hello-world",
		},
	}

	s, err := sanitizer.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				Slug string `sanitize:"slug"`
			}
			input := &testStruct{Slug: tt.input}
			err := s.SanitizeStruct(input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, input.Slug)
		})
	}
}

func TestCaseConversions(t *testing.T) {
	// Create a new sanitizer instance for the tests
	s, err := sanitizer.New()
	require.NoError(t, err)

	t.Run("CamelCase", func(t *testing.T) {
		input := &TestStruct{CamelCaseField: "hello world"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "helloWorld", input.CamelCaseField)
	})

	t.Run("SnakeCase", func(t *testing.T) {
		input := &TestStruct{SnakeCaseField: "helloWorld"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello_world", input.SnakeCaseField)
	})

	t.Run("KebabCase", func(t *testing.T) {
		input := &TestStruct{KebabCaseField: "helloWorld"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "hello-world", input.KebabCaseField)
	})

	t.Run("UCFirst", func(t *testing.T) {
		input := &TestStruct{UCFirstField: "hello"}
		err := s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "Hello", input.UCFirstField)
	})

	t.Run("Instance sanitizer behavior", func(t *testing.T) {
		// Test with a new sanitizer instance
		s, err := sanitizer.New()
		require.NoError(t, err)

		input := &TestStruct{
			CamelCaseField: "hello world",
			SnakeCaseField: "helloWorld",
			KebabCaseField: "helloWorld",
			UCFirstField:   "hello",
		}
		err = s.SanitizeStruct(input)
		require.NoError(t, err)
		assert.Equal(t, "helloWorld", input.CamelCaseField)
		assert.Equal(t, "hello_world", input.SnakeCaseField)
		assert.Equal(t, "hello-world", input.KebabCaseField)
		assert.Equal(t, "Hello", input.UCFirstField)
	})
}
