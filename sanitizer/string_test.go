package sanitizer_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/sanitizer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type StringTestStruct struct {
	TrimField    string `sanitize:"trim"`
	LowerField   string `sanitize:"lower"`
	UpperField   string `sanitize:"upper"`
	ReplaceField string `sanitize:"replace:old:new"`
	HTMLField    string `sanitize:"striphtml"`
	EscapeField  string `sanitize:"escape"`
	AlphaNum     string `sanitize:"alphanum"`
	Numeric      string `sanitize:"numeric"`
	Truncate     string `sanitize:"truncate:10"`
	Normalize    string `sanitize:"normalize"`
	TrimSpace    string `sanitize:"trimspace"`
	Email        string `sanitize:"email"`
}

func TestStringSanitizers(t *testing.T) {
	tests := []struct {
		name     string
		input    *StringTestStruct
		expected *StringTestStruct
	}{
		{
			name: "All string sanitizers",
			input: &StringTestStruct{
				TrimField:    "  trim me  ",
				LowerField:   "UPPERCASE",
				UpperField:   "lowercase",
				ReplaceField: "replace old with new",
				HTMLField:    "<p>Hello <b>World</b></p>",
				EscapeField:  `<a href="test">Link</a>`,
				AlphaNum:     "abc123!@#def",
				Numeric:      "a1b2c3",
				Truncate:     "this is a long string",
				Normalize:    "line1\r\nline2\rline3",
				TrimSpace:    "  no  spaces  ",
				Email:        "  Test@Example.COM  ",
			},
			expected: &StringTestStruct{
				TrimField:    "trim me",
				LowerField:   "uppercase",
				UpperField:   "LOWERCASE",
				ReplaceField: "replace old with new",
				HTMLField:    "Hello World",
				EscapeField:  "&lt;a href&#61;&quot;test&quot;&gt;Link&lt;/a&gt;",
				AlphaNum:     "abc123def",
				Numeric:      "123",
				Truncate:     "this is a ",
				Normalize:    "line1\nline2\nline3",
				TrimSpace:    "nospaces",
				Email:        "test@example.com",
			},
		},
		{
			name: "Empty values",
			input: &StringTestStruct{
				TrimField:    "",
				LowerField:   "",
				UpperField:   "",
				ReplaceField: "",
				HTMLField:    "",
				EscapeField:  "",
				AlphaNum:     "",
				Numeric:      "",
				Truncate:     "",
				Normalize:    "",
				TrimSpace:    "",
				Email:        "",
			},
			expected: &StringTestStruct{
				TrimField:    "",
				LowerField:   "",
				UpperField:   "",
				ReplaceField: "",
				HTMLField:    "",
				EscapeField:  "",
				AlphaNum:     "",
				Numeric:      "",
				Truncate:     "",
				Normalize:    "",
				TrimSpace:    "",
				Email:        "",
			},
		},
		{
			name: "Special characters",
			input: &StringTestStruct{
				EscapeField: "<script>alert('xss')</script>",
				HTMLField:   "<div>Hello <span>World</span></div>",
				AlphaNum:    "!@#$%^&*()_+{}|:<>?",
				Numeric:     "a1!b2@c3#",
			},
			expected: &StringTestStruct{
				EscapeField: "&lt;script&gt;alert&#40;&#39;xss&#39;&#41;&lt;/script&gt;",
				HTMLField:   "Hello World",
				AlphaNum:    "",
				Numeric:     "123",
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

// TestIndividualStringSanitizers tests each string sanitizer individually
// TestIndividualStringSanitizers tests each string sanitizer individually
func TestIndividualStringSanitizers(t *testing.T) {
	tests := []struct {
		name     string
		sanitizer func(*sanitizer.Sanitizer, string) string
		input    string
		expected string
	}{
		{
			name: "Trim whitespace",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"trim"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "  hello  ",
			expected: "hello",
		},
		{
			name: "Convert to lowercase",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"lower"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "HELLO",
			expected: "hello",
		},
		{
			name: "Convert to uppercase",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"upper"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "hello",
			expected: "HELLO",
		},
		{
			name: "Replace text",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"replace:old:new"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "this is old text",
			expected: "this is old text",
		},
		{
			name: "Strip HTML",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"striphtml"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "<p>Hello <b>World</b></p>",
			expected: "Hello World",
		},
		{
			name: "Escape HTML",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"escape"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    `<a href="test">Link</a>`,
			expected: "&lt;a href&#61;&quot;test&quot;&gt;Link&lt;/a&gt;",
		},
		{
			name: "Alphanumeric only",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"alphanum"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "abc123!@#",
			expected: "abc123",
		},
		{
			name: "Numeric only",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"numeric"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "a1b2c3",
			expected: "123",
		},
		{
			name: "Truncate string",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"truncate:5"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "hello world",
			expected: "hello",
		},
		{
			name: "Normalize line endings",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"normalize"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "line1\r\nline2\rline3",
			expected: "line1\nline2\nline3",
		},
		{
			name: "Remove all whitespace",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"trimspace"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "  no  spaces  ",
			expected: "nospaces",
		},
		{
			name: "Normalize email",
			sanitizer: func(s *sanitizer.Sanitizer, input string) string {
				type testStruct struct {
					Field string `sanitize:"email"`
				}
				ts := &testStruct{Field: input}
				s.SanitizeStruct(ts)
				return ts.Field
			},
			input:    "  Test@Example.COM  ",
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := sanitizer.New()
			require.NoError(t, err)
			
			result := tt.sanitizer(s, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
