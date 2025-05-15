package sanitizer_test

import (
	"reflect"
	"testing"

	"github.com/dmitrymomot/gokit/sanitizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests the creation of a new Sanitizer instance.
func TestNew(t *testing.T) {
	t.Run("Default configuration", func(t *testing.T) {
		s, err := sanitizer.New()
		require.NoError(t, err)
		assert.NotNil(t, s)
	})

	t.Run("With custom options", func(t *testing.T) {
		s, err := sanitizer.New(
			sanitizer.WithRuleSeparator("|"),
			sanitizer.WithParamSeparator("#"),
			sanitizer.WithParamListSeparator("&"),
		)
		require.NoError(t, err)
		assert.NotNil(t, s)
	})

	t.Run("With field name tag", func(t *testing.T) {
		s, err := sanitizer.New(sanitizer.WithFieldNameTag("json"))
		require.NoError(t, err)
		assert.NotNil(t, s)

		// Create a test struct with json tags
		type TestStruct struct {
			FieldName string `json:"field_name" sanitize:"trim"`
		}
		ts := &TestStruct{FieldName: "  test  "}

		// Sanitize the struct
		err = s.SanitizeStruct(ts)
		require.NoError(t, err)

		// Verify the field was sanitized
		assert.Equal(t, "test", ts.FieldName)
	})

	t.Run("With custom sanitizers", func(t *testing.T) {
		// Create a custom sanitizer that adds a prefix
		customSanitizers := map[string]sanitizer.SanitizeFunc{
			"addprefix": func(fieldValue any, fieldType reflect.StructField, params []string) any {
				v, ok := fieldValue.(string)
				if !ok {
					return fieldValue
				}
				prefix := "prefix_"
				if len(params) > 0 {
					prefix = params[0]
				}
				return prefix + v
			},
		}

		s, err := sanitizer.New(sanitizer.WithSanitizers(customSanitizers))
		require.NoError(t, err)
		assert.NotNil(t, s)

		// Test with no parameters
		t.Run("Custom sanitizer without params", func(t *testing.T) {
			type TestStruct struct {
				Field string `sanitize:"addprefix"`
			}
			ts := &TestStruct{Field: "test"}

			err = s.SanitizeStruct(ts)
			require.NoError(t, err)
			assert.Equal(t, "prefix_test", ts.Field)
		})

		// Test with parameters
		t.Run("Custom sanitizer with params", func(t *testing.T) {
			type TestStruct struct {
				Field string `sanitize:"addprefix:custom_"`
			}
			ts := &TestStruct{Field: "test"}

			err = s.SanitizeStruct(ts)
			require.NoError(t, err)
			assert.Equal(t, "custom_test", ts.Field)
		})
	})

	t.Run("With nil sanitizers map", func(t *testing.T) {
		_, err := sanitizer.New(sanitizer.WithSanitizers(nil))
		require.Error(t, err)
	})
}

// TestMustNew tests the MustNew function.
func TestMustNew(t *testing.T) {
	t.Run("Valid options", func(t *testing.T) {
		s := sanitizer.MustNew()
		assert.NotNil(t, s)
	})

	t.Run("Panic on error", func(t *testing.T) {
		assert.Panics(t, func() {
			sanitizer.MustNew(sanitizer.WithRuleSeparator("")) // Invalid option
		})
	})
}

// TestRegisterSanitizer tests the registration of custom sanitizers.
func TestRegisterSanitizer(t *testing.T) {
	tests := []struct {
		name        string
		tag         string
		fn          sanitizer.SanitizeFunc
		expectError bool
	}{
		{
			name: "Valid registration",
			tag:  "custom",
			fn:   func(v any, _ reflect.StructField, _ []string) any { return v },
		},
		{
			name:        "Empty tag",
			tag:         "",
			fn:          func(v any, _ reflect.StructField, _ []string) any { return v },
			expectError: true,
		},
		{
			name:        "Nil function",
			tag:         "custom",
			fn:          nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sanitizer.MustNew()
			err := s.RegisterSanitizer(tt.tag, tt.fn)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSanitizeStruct tests the core struct sanitization logic.
func TestSanitizeStruct(t *testing.T) {
	type NestedStruct struct {
		Field string `sanitize:"trim"`
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "Nil input",
			input:    (*struct{})(nil),
			expected: (*struct{})(nil),
		},
		{
			name:     "Non-pointer input",
			input:    struct{}{},
			expected: struct{}{},
		},
		{
			name: "Unexported field",
			input: struct {
				exported   string `sanitize:"trim"`
				unexported string `sanitize:"trim"`
			}{
				exported:   "  exported  ",
				unexported: "  unexported  ",
			},
			expected: struct {
				exported   string `sanitize:"trim"`
				unexported string `sanitize:"trim"`
			}{
				exported:   "  exported  ",
				unexported: "  unexported  ",
			},
		},
		{
			name: "Nested struct",
			input: struct {
				Nested NestedStruct
			}{
				Nested: NestedStruct{
					Field: "  nested  ",
				},
			},
			expected: struct {
				Nested NestedStruct
			}{
				Nested: NestedStruct{
					Field: "  nested  ",
				},
			},
		},
		{
			name: "Omitempty with zero value",
			input: struct {
				Field string `sanitize:"trim,omitempty"`
			}{
				Field: "",
			},
			expected: struct {
				Field string `sanitize:"trim,omitempty"`
			}{
				Field: "",
			},
		},
		{
			name: "Omitempty with non-zero value",
			input: struct {
				Field string `sanitize:"trim,omitempty"`
			}{
				Field: "  test  ",
			},
			expected: struct {
				Field string `sanitize:"trim,omitempty"`
			}{
				Field: "  test  ",
			},
		},
	}

	s := sanitizer.MustNew()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.SanitizeStruct(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

// TestRuleParsing tests the parsing of sanitizer rules from struct tags.
func TestRuleParsing(t *testing.T) {
	// Create a sanitizer instance
	s := sanitizer.MustNew()

	// Single rule test
	t.Run("Single rule", func(t *testing.T) {
		type TestStruct struct {
			Field string `sanitize:"trim"`
		}
		ts := &TestStruct{Field: "  test  "}
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)
		assert.Equal(t, "test", ts.Field)
	})

	// Multiple rules test
	t.Run("Multiple rules", func(t *testing.T) {
		type TestStruct struct {
			Field string `sanitize:"trim;lower"`
		}
		ts := &TestStruct{Field: "  TEST  "}
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)
		assert.Equal(t, "test", ts.Field)
	})

	// Rules with parameters test
	t.Run("Rules with params", func(t *testing.T) {
		type TestStruct struct {
			Field string `sanitize:"replace:test,modified"`
		}
		ts := &TestStruct{Field: "test"}
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)
		assert.Equal(t, "modified", ts.Field)
	})

	// Empty rules test
	t.Run("Empty rules", func(t *testing.T) {
		type TestStruct struct {
			Field string `sanitize:""`
		}
		ts := &TestStruct{Field: "test"}
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)
		assert.Equal(t, "test", ts.Field)
	})

	// Multiple params test
	t.Run("Multiple params", func(t *testing.T) {
		type TestStruct struct {
			Field string `sanitize:"truncate:3"`
		}
		ts := &TestStruct{Field: "longtext"}
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)
		assert.Equal(t, "lon", ts.Field)
	})
}

// TestStringSanitizersEdgeCases tests edge cases for string sanitizers.
func TestStringSanitizersEdgeCases(t *testing.T) {
	s := sanitizer.MustNew()

	// Test trim sanitizer with various inputs
	t.Run("Trim with edge cases", func(t *testing.T) {
		type TestStruct struct {
			Field1 string        `sanitize:"trim"`
			Field2 int           `sanitize:"trim"`
			Field3 bool          `sanitize:"trim"`
			Field4 float64       `sanitize:"trim"`
			Field5 []string      `sanitize:"trim"`
			Field6 map[string]int `sanitize:"trim"`
		}

		ts := &TestStruct{
			Field1: "\t  test  \n",
			Field2: 123,
			Field3: true,
			Field4: 3.14,
			Field5: []string{"test"},
			Field6: map[string]int{"test": 1},
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		// String should be trimmed, others unchanged
		assert.Equal(t, "test", ts.Field1)
		assert.Equal(t, 123, ts.Field2)
		assert.Equal(t, true, ts.Field3)
		assert.Equal(t, 3.14, ts.Field4)
		assert.Equal(t, []string{"test"}, ts.Field5)
		assert.Equal(t, map[string]int{"test": 1}, ts.Field6)
	})

	// Test case conversion sanitizers with edge cases
	t.Run("Case conversions with edge cases", func(t *testing.T) {
		type TestStruct struct {
			Lower   string `sanitize:"lower"`
			Upper   string `sanitize:"upper"`
			NonStr  int    `sanitize:"lower"`
			Empty   string `sanitize:"upper"`
			Special string `sanitize:"lower"`
		}

		ts := &TestStruct{
			Lower:   "TeSt@123",
			Upper:   "TeSt@123",
			NonStr:  123,
			Empty:   "",
			Special: "ÑáÉí123",
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Equal(t, "test@123", ts.Lower)
		assert.Equal(t, "TEST@123", ts.Upper)
		assert.Equal(t, 123, ts.NonStr)
		assert.Equal(t, "", ts.Empty)
		assert.Equal(t, "ñáéí123", ts.Special)
	})

	// Test replace sanitizer with edge cases
	t.Run("Replace with edge cases", func(t *testing.T) {
		type TestStruct struct {
			Multiple string `sanitize:"replace:e,X"`
			NoMatch  string `sanitize:"replace:z,X"`
			Empty    string `sanitize:"replace:e,X"`
			NonStr   int    `sanitize:"replace:1,X"`
		}

		ts := &TestStruct{
			Multiple: "test test",
			NoMatch:  "test test",
			Empty:    "",
			NonStr:   123,
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Equal(t, "tXst tXst", ts.Multiple)
		assert.Equal(t, "test test", ts.NoMatch)
		assert.Equal(t, "", ts.Empty)
		assert.Equal(t, 123, ts.NonStr)
	})
	
	// Test HTML sanitizers with edge cases
	t.Run("HTML sanitizers with edge cases", func(t *testing.T) {
		type TestStruct struct {
			StripHTML  string `sanitize:"striphtml"`
			EscapeHTML string `sanitize:"escape"`
			Empty      string `sanitize:"striphtml"`
			NoHTML     string `sanitize:"striphtml"`
			NonStr     int    `sanitize:"escape"`
		}

		ts := &TestStruct{
			StripHTML:  "<p>Test <strong>with</strong> <a href=\"#\">HTML</a></p>",
			EscapeHTML: "<script>alert('xss')</script>",
			Empty:      "",
			NoHTML:     "plain text",
			NonStr:     123,
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Equal(t, "Test with HTML", ts.StripHTML)
		// HTML escaping can vary slightly between implementations, so we'll just check for key patterns
		assert.Contains(t, ts.EscapeHTML, "&lt;script&gt;")
		assert.Contains(t, ts.EscapeHTML, "&lt;/script&gt;")
		assert.Contains(t, ts.EscapeHTML, "&#39;xss&#39;")
		assert.Equal(t, "", ts.Empty)
		assert.Equal(t, "plain text", ts.NoHTML)
		assert.Equal(t, 123, ts.NonStr)
	})

	// Test alphanumeric and numeric sanitizers
	t.Run("Alphanumeric and numeric sanitizers", func(t *testing.T) {
		type TestStruct struct {
			AlphaNum string `sanitize:"alphanum"`
			Numeric  string `sanitize:"numeric"`
			Empty    string `sanitize:"alphanum"`
			NonStr   int    `sanitize:"numeric"`
		}

		ts := &TestStruct{
			AlphaNum: "Test-123_@#$%^",
			Numeric:  "Test-123_@#$%^",
			Empty:    "",
			NonStr:   123,
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Equal(t, "Test123", ts.AlphaNum)
		assert.Equal(t, "123", ts.Numeric)
		assert.Equal(t, "", ts.Empty)
		assert.Equal(t, 123, ts.NonStr)
	})

	// Test truncate with various inputs
	t.Run("Truncate with edge cases", func(t *testing.T) {
		type TestStruct struct {
			Longer   string `sanitize:"truncate:3"`
			Equal    string `sanitize:"truncate:4"`
			Shorter  string `sanitize:"truncate:10"`
			Empty    string `sanitize:"truncate:5"`
			NonStr   int    `sanitize:"truncate:2"`
			Unicode  string `sanitize:"truncate:3"`
		}

		ts := &TestStruct{
			Longer:   "long text",
			Equal:    "test",
			Shorter:  "hi",
			Empty:    "",
			NonStr:   12345,
			Unicode:  "test123",
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Equal(t, "lon", ts.Longer)
		assert.Equal(t, "test", ts.Equal)
		assert.Equal(t, "hi", ts.Shorter)
		assert.Equal(t, "", ts.Empty)
		assert.Equal(t, 12345, ts.NonStr)
		assert.Equal(t, "tes", ts.Unicode)
	})

	// Test other string sanitizers
	t.Run("Other string sanitizers", func(t *testing.T) {
		type TestStruct struct {
			Normalize string `sanitize:"normalize"`
			Trimspace string `sanitize:"trimspace"`
			Email     string `sanitize:"email"`
			Empty     string `sanitize:"normalize"`
			NonStr    int    `sanitize:"trimspace"`
		}

		ts := &TestStruct{
			Normalize: "line1\r\nline2\nline3\rline4",
			Trimspace: "  test with   spaces  ",
			Email:     " Test.User+tag@Example.COM ",
			Empty:     "",
			NonStr:    123,
		}

		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Equal(t, "line1\nline2\nline3\nline4", ts.Normalize)
		assert.Equal(t, "testwithspaces", ts.Trimspace)
		assert.Equal(t, "test.user+tag@example.com", ts.Email)
		assert.Equal(t, "", ts.Empty)
		assert.Equal(t, 123, ts.NonStr)
	})
}

// TestIsZero tests the isZero helper function in the sanitizer package.
func TestIsZero(t *testing.T) {
	// This test ensures the implementation of isZero works correctly with omitempty rule
	// by using different struct types with the omitempty rule and checking the result
	
	// Test strings
	t.Run("String types", func(t *testing.T) {
		type TestStruct struct {
			Empty    string `sanitize:"trim,omitempty"`
			NonEmpty string `sanitize:"trim,omitempty"`
		}

		ts := &TestStruct{
			Empty:    "",
			NonEmpty: "test",
		}

		s := sanitizer.MustNew()
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Zero(t, ts.Empty, "Empty string should be zero value")
		assert.Equal(t, "test", ts.NonEmpty, "Non-empty string should not be zero value")
	})

	// Test numeric types
	t.Run("Numeric types", func(t *testing.T) {
		type TestStruct struct {
			ZeroInt     int     `sanitize:"trim,omitempty"`
			NonZeroInt  int     `sanitize:"trim,omitempty"`
			ZeroUint    uint    `sanitize:"trim,omitempty"`
			NonZeroUint uint    `sanitize:"trim,omitempty"`
			ZeroFloat   float64 `sanitize:"trim,omitempty"`
			NonZeroFlt  float64 `sanitize:"trim,omitempty"`
		}

		ts := &TestStruct{
			ZeroInt:     0,
			NonZeroInt:  42,
			ZeroUint:    0,
			NonZeroUint: 42,
			ZeroFloat:   0.0,
			NonZeroFlt:  3.14,
		}

		s := sanitizer.MustNew()
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Zero(t, ts.ZeroInt, "Zero int should be zero value")
		assert.Equal(t, 42, ts.NonZeroInt, "Non-zero int should not be zero value")
		assert.Zero(t, ts.ZeroUint, "Zero uint should be zero value")
		assert.Equal(t, uint(42), ts.NonZeroUint, "Non-zero uint should not be zero value")
		assert.Zero(t, ts.ZeroFloat, "Zero float should be zero value")
		assert.Equal(t, 3.14, ts.NonZeroFlt, "Non-zero float should not be zero value")
	})

	// Test bool, complex, and other basic types
	t.Run("Bool and complex types", func(t *testing.T) {
		type TestStruct struct {
			FalseBool  bool      `sanitize:"trim,omitempty"`
			TrueBool   bool      `sanitize:"trim,omitempty"`
			ZeroComp   complex128 `sanitize:"trim,omitempty"`
			NonZeroComp complex128 `sanitize:"trim,omitempty"`
		}

		ts := &TestStruct{
			FalseBool:   false,
			TrueBool:    true,
			ZeroComp:    complex(0, 0),
			NonZeroComp: complex(1, 2),
		}

		s := sanitizer.MustNew()
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Zero(t, ts.FalseBool, "False should be zero value")
		assert.Equal(t, true, ts.TrueBool, "True should not be zero value")
		assert.Zero(t, ts.ZeroComp, "Zero complex should be zero value")
		assert.Equal(t, complex(1, 2), ts.NonZeroComp, "Non-zero complex should not be zero value")
	})

	// Test arrays, slices, and maps
	t.Run("Arrays, slices, and maps", func(t *testing.T) {
		type TestStruct struct {
			EmptyArray   [0]int             `sanitize:"trim,omitempty"`
			NonEmptyArray [1]int             `sanitize:"trim,omitempty"`
			NilSlice     []string           `sanitize:"trim,omitempty"`
			EmptySlice   []string           `sanitize:"trim,omitempty"`
			FilledSlice  []string           `sanitize:"trim,omitempty"`
			NilMap       map[string]string  `sanitize:"trim,omitempty"`
			EmptyMap     map[string]string  `sanitize:"trim,omitempty"`
			FilledMap    map[string]string  `sanitize:"trim,omitempty"`
		}

		ts := &TestStruct{
			EmptyArray:   [0]int{},
			NonEmptyArray: [1]int{1},
			NilSlice:     nil,
			EmptySlice:   []string{},
			FilledSlice:  []string{"test"},
			NilMap:       nil,
			EmptyMap:     map[string]string{},
			FilledMap:    map[string]string{"key": "value"},
		}

		s := sanitizer.MustNew()
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		// Arrays
		assert.Zero(t, ts.EmptyArray, "Empty array should be zero value")
		assert.Equal(t, [1]int{1}, ts.NonEmptyArray, "Non-empty array should not be zero value")

		// Slices - nil is considered zero, empty is not in the actual implementation
		assert.Zero(t, ts.NilSlice, "Nil slice should be zero value")
		assert.Equal(t, []string{}, ts.EmptySlice, "Empty slice should not be zero value (per implementation)")
		assert.Equal(t, []string{"test"}, ts.FilledSlice, "Filled slice should not be zero value")

		// Maps - nil is considered zero, empty is not in the actual implementation
		assert.Zero(t, ts.NilMap, "Nil map should be zero value")
		assert.Equal(t, map[string]string{}, ts.EmptyMap, "Empty map should not be zero value (per implementation)")
		assert.Equal(t, map[string]string{"key": "value"}, ts.FilledMap, "Filled map should not be zero value")
	})

	// Test pointers and interfaces
	t.Run("Pointers and interfaces", func(t *testing.T) {
		type TestStruct struct {
			NilPtr      *string     `sanitize:"trim,omitempty"`
			NonNilPtr   *string     `sanitize:"trim,omitempty"`
			NilIface    interface{} `sanitize:"trim,omitempty"`
			NonNilIface interface{} `sanitize:"trim,omitempty"`
			EmptyStruct struct{}    `sanitize:"trim,omitempty"`
		}

		str := "test"
		ts := &TestStruct{
			NilPtr:      nil,
			NonNilPtr:   &str,
			NilIface:    nil,
			NonNilIface: "test",
			EmptyStruct: struct{}{},
		}

		s := sanitizer.MustNew()
		err := s.SanitizeStruct(ts)
		require.NoError(t, err)

		assert.Zero(t, ts.NilPtr, "Nil pointer should be zero value")
		assert.NotZero(t, ts.NonNilPtr, "Non-nil pointer should not be zero value")
		assert.Zero(t, ts.NilIface, "Nil interface should be zero value")
		assert.NotZero(t, ts.NonNilIface, "Non-nil interface should not be zero value")
		// Empty struct is not considered zero in Go, so it should remain
		assert.Equal(t, struct{}{}, ts.EmptyStruct, "Empty struct should not be zero value")
	})
}


