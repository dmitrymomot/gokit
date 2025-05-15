package sanitizer_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/sanitizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CaseTestStruct struct {
	CapitalizeField string `sanitize:"capitalize"`
	CamelCaseField  string `sanitize:"camelcase"`
	PascalCaseField string `sanitize:"pascalcase"`
	SnakeCaseField  string `sanitize:"snakecase"`
	KebabCaseField  string `sanitize:"kebabcase"`
	UCFirstField    string `sanitize:"ucfirst"`
	LCFirstField    string `sanitize:"lcfirst"`
}

func TestCaseConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    *CaseTestStruct
		expected *CaseTestStruct
	}{
	{
		name: "Capitalize field",
		input: &CaseTestStruct{
			CapitalizeField: "hello world",
		},
		expected: &CaseTestStruct{
			CapitalizeField: "Hello world",
		},
	},
	{
		name: "CamelCase field",
		input: &CaseTestStruct{
			CamelCaseField: "hello_world_example",
		},
		expected: &CaseTestStruct{
			CamelCaseField: "helloWorldExample",
		},
	},
	{
		name: "PascalCase field",
		input: &CaseTestStruct{
			PascalCaseField: "hello_world_example",
		},
		expected: &CaseTestStruct{
			PascalCaseField: "HelloWorldExample",
		},
	},
	{
		name: "SnakeCase field",
		input: &CaseTestStruct{
			SnakeCaseField: "HelloWorldExample",
		},
		expected: &CaseTestStruct{
			SnakeCaseField: "hello_world_example",
		},
	},
	{
		name: "KebabCase field",
		input: &CaseTestStruct{
			KebabCaseField: "HelloWorldExample",
		},
		expected: &CaseTestStruct{
			KebabCaseField: "hello-world-example",
		},
	},
	{
		name: "UCFirst field",
		input: &CaseTestStruct{
			UCFirstField: "hello world",
		},
		expected: &CaseTestStruct{
			UCFirstField: "Hello world",
		},
	},
	{
		name: "LCFirst field",
		input: &CaseTestStruct{
			LCFirstField: "Hello World",
		},
		expected: &CaseTestStruct{
			LCFirstField: "hello World",
		},
	},
	{
		name: "Empty strings",
		input: &CaseTestStruct{
			CapitalizeField: "",
			CamelCaseField:  "",
			PascalCaseField: "",
			SnakeCaseField:  "",
			KebabCaseField:  "",
			UCFirstField:    "",
			LCFirstField:    "",
		},
		expected: &CaseTestStruct{
			CapitalizeField: "",
			CamelCaseField:  "",
			PascalCaseField: "",
			SnakeCaseField:  "",
			KebabCaseField:  "",
			UCFirstField:    "",
			LCFirstField:    "",
		},
	},
	{
		name: "Single character strings",
		input: &CaseTestStruct{
			CapitalizeField: "a",
			UCFirstField:    "A",
			LCFirstField:    "A",
		},
		expected: &CaseTestStruct{
			CapitalizeField: "A",
			UCFirstField:    "A",
			LCFirstField:    "a",
		},
	},
	}

	s, err := sanitizer.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SanitizeStruct(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestCaseConversionsWithMultipleRules(t *testing.T) {
	type TestStruct struct {
		Field1 string `sanitize:"trim;lower;snakecase"`
		Field2 string `sanitize:"trim;upper;camelcase"`
	}

	tests := []struct {
		name     string
		input    *TestStruct
		expected *TestStruct
	}{
	{
		name: "Multiple rules with case conversion",
		input: &TestStruct{
			Field1: "  Hello World Example  ",
			Field2: "  HELLO_WORLD_EXAMPLE  ",
		},
		expected: &TestStruct{
			Field1: "hello_world_example",
			Field2: "helloWorldExample",
		},
	},
	}

	s, err := sanitizer.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SanitizeStruct(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}
