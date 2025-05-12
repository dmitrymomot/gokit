package validator_test

import (
	"strings"
	"testing"

	"github.com/dmitrymomot/gokit/validator"
	"github.com/stretchr/testify/require"
)

func TestRegexValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid regex pattern",
			structWithTag: struct{ Field string `validate:"regex:^[a-z]+$"` }{Field: "abc"},
		},
		{
			name:            "invalid value",
			structWithTag:   struct{ Field string `validate:"regex:^[a-z]+$"` }{Field: "123"},
			wantErrContains: "regex",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Field string `validate:"regex:^[a-z]+$"` }{Field: ""},
			wantErrContains: "regex",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"regex:^[a-z]+$"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
		{
			name:            "missing pattern",
			structWithTag:   struct{ Field string `validate:"regex"` }{Field: "abc"},
			wantErrContains: "missing_regex_pattern",
		},
		{
			name:            "invalid regex pattern",
			structWithTag:   struct{ Field string `validate:"regex:["` }{Field: "abc"},
			wantErrContains: "invalid_regex_pattern",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("regex"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			// In case DNS lookup fails even for valid domains (offline or DNS issues)
			if err != nil && tt.wantErrContains == "" {
				if strings.Contains(err.Error(), "lookup") {
					t.Skipf("DNS lookup failed, test may be running offline: %v", err)
				}
				require.NoError(t, err)
			} else if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestEmailValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid email",
			structWithTag: struct{ Email string `validate:"email"` }{Email: "test@dmomot.com"},
		},
		{
			name:            "invalid email",
			structWithTag:   struct{ Email string `validate:"email"` }{Email: "not-an-email"},
			wantErrContains: "email",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Email string `validate:"email"` }{Email: ""},
			wantErrContains: "email",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Email int `validate:"email"` }{Email: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("email"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestRealEmailValidator(t *testing.T) {
	// Skip this test in short mode - MX record check requires network access
	if testing.Short() {
		t.Skip("skipping TestRealEmailValidator in short mode (requires network access)")
	}
	
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid email with common domain",
			structWithTag: struct{ Email string `validate:"realemail"` }{Email: "test@dmomot.com"},
		},
		{
			name:          "valid email with alternate address",
			structWithTag: struct{ Email string `validate:"realemail"` }{Email: "info@dmomot.com"},
		},
		{
			name:          "valid email with correct domain",
			structWithTag: struct{ Email string `validate:"realemail"` }{Email: "contact@dmomot.com"},
		},
		{
			name:            "invalid email format",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "not-an-email"},
			wantErrContains: "realemail",
		},
		{
			name:            "missing @ symbol",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "user.example.com"},
			wantErrContains: "realemail",
		},
		{
			name:            "url instead of email",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "http://example.com"},
			wantErrContains: "realemail",
		},
		{
			name:            "missing domain part",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "user@"},
			wantErrContains: "realemail",
		},
		{
			name:            "domain without TLD",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "user@localhost"},
			wantErrContains: "realemail",
		},
		{
			name:            "TLD too short",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "user@example.a"},
			wantErrContains: "realemail",
		},
		{
			name:            "TLD with numbers",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: "user@example.c0m"},
			wantErrContains: "realemail",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Email string `validate:"realemail"` }{Email: ""},
			wantErrContains: "realemail",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Email int `validate:"realemail"` }{Email: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("realemail"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestFullnameValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid full name",
			structWithTag: struct{ Name string `validate:"fullname"` }{Name: "John Doe"},
		},
		{
			name:            "too short",
			structWithTag:   struct{ Name string `validate:"fullname"` }{Name: "Jo"},
			wantErrContains: "fullname",
		},
		{
			name:            "contains numbers",
			structWithTag:   struct{ Name string `validate:"fullname"` }{Name: "John Doe123"},
			wantErrContains: "fullname",
		},
		{
			name:            "starts with space",
			structWithTag:   struct{ Name string `validate:"fullname"` }{Name: " John Doe"},
			wantErrContains: "fullname",
		},
		{
			name:            "ends with space",
			structWithTag:   struct{ Name string `validate:"fullname"` }{Name: "John Doe "},
			wantErrContains: "fullname",
		},
		{
			name:            "double spaces",
			structWithTag:   struct{ Name string `validate:"fullname"` }{Name: "John  Doe"},
			wantErrContains: "fullname",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Name string `validate:"fullname"` }{Name: ""},
			wantErrContains: "fullname",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Name int `validate:"fullname"` }{Name: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("fullname"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestNameValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid name",
			structWithTag: struct{ Name string `validate:"name"` }{Name: "John"},
		},
		{
			name:          "valid name with space",
			structWithTag: struct{ Name string `validate:"name"` }{Name: "John Doe"},
		},
		{
			name:            "too short",
			structWithTag:   struct{ Name string `validate:"name"` }{Name: "J"},
			wantErrContains: "name",
		},
		{
			name:            "starts with number",
			structWithTag:   struct{ Name string `validate:"name"` }{Name: "1John"},
			wantErrContains: "name",
		},
		{
			name:            "contains number",
			structWithTag:   struct{ Name string `validate:"name"` }{Name: "John2Doe"},
			wantErrContains: "name",
		},
		{
			name:            "starts with space",
			structWithTag:   struct{ Name string `validate:"name"` }{Name: " John"},
			wantErrContains: "name",
		},
		{
			name:            "ends with space",
			structWithTag:   struct{ Name string `validate:"name"` }{Name: "John "},
			wantErrContains: "name",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Name string `validate:"name"` }{Name: ""},
			wantErrContains: "name",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Name int `validate:"name"` }{Name: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("name"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestAlphaValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid alpha string",
			structWithTag: struct{ Field string `validate:"alpha"` }{Field: "abcDEF"},
		},
		{
			name:            "contains number",
			structWithTag:   struct{ Field string `validate:"alpha"` }{Field: "abc123"},
			wantErrContains: "alpha",
		},
		{
			name:            "contains space",
			structWithTag:   struct{ Field string `validate:"alpha"` }{Field: "abc def"},
			wantErrContains: "alpha",
		},
		{
			name:            "contains special character",
			structWithTag:   struct{ Field string `validate:"alpha"` }{Field: "abc@def"},
			wantErrContains: "alpha",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Field string `validate:"alpha"` }{Field: ""},
			wantErrContains: "alpha",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"alpha"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("alpha"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestAlphanumValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid alphanumeric string",
			structWithTag: struct{ Field string `validate:"alphanum"` }{Field: "abc123DEF456"},
		},
		{
			name:            "contains space",
			structWithTag:   struct{ Field string `validate:"alphanum"` }{Field: "abc 123"},
			wantErrContains: "alphanum",
		},
		{
			name:            "contains special character",
			structWithTag:   struct{ Field string `validate:"alphanum"` }{Field: "abc@123"},
			wantErrContains: "alphanum",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Field string `validate:"alphanum"` }{Field: ""},
			wantErrContains: "alphanum",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"alphanum"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("alphanum"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestAlphaSpaceValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid alpha with space",
			structWithTag: struct{ Field string `validate:"alphaspace"` }{Field: "abc DEF"},
		},
		{
			name:            "contains number",
			structWithTag:   struct{ Field string `validate:"alphaspace"` }{Field: "abc 123"},
			wantErrContains: "alphaspace",
		},
		{
			name:            "contains special character",
			structWithTag:   struct{ Field string `validate:"alphaspace"` }{Field: "abc @def"},
			wantErrContains: "alphaspace",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Field string `validate:"alphaspace"` }{Field: ""},
			wantErrContains: "alphaspace",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"alphaspace"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("alphaspace"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestAlphaSpaceNumValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid alphanumeric with spaces",
			structWithTag: struct{ Field string `validate:"alphaspacenum"` }{Field: "abc 123 DEF 456"},
		},
		{
			name:            "contains special character",
			structWithTag:   struct{ Field string `validate:"alphaspacenum"` }{Field: "abc 123 @def"},
			wantErrContains: "alphaspacenum",
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Field string `validate:"alphaspacenum"` }{Field: ""},
			wantErrContains: "alphaspacenum",
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"alphaspacenum"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("alphaspacenum"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestStartsWithValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid starts with",
			structWithTag: struct{ Field string `validate:"startswith:abc"` }{Field: "abcdef"},
		},
		{
			name:            "doesn't start with prefix",
			structWithTag:   struct{ Field string `validate:"startswith:abc"` }{Field: "defabc"},
			wantErrContains: "startswith",
		},
		{
			name:          "empty string with empty prefix",
			structWithTag: struct{ Field string `validate:"startswith:"` }{Field: ""},
		},
		{
			name:          "non-empty string with empty prefix",
			structWithTag: struct{ Field string `validate:"startswith:"` }{Field: "abc"},
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"startswith:abc"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("startswith"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestEndsWithValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid ends with",
			structWithTag: struct{ Field string `validate:"endswith:def"` }{Field: "abcdef"},
		},
		{
			name:            "doesn't end with suffix",
			structWithTag:   struct{ Field string `validate:"endswith:abc"` }{Field: "defabc123"},
			wantErrContains: "endswith",
		},
		{
			name:          "empty string with empty suffix",
			structWithTag: struct{ Field string `validate:"endswith:"` }{Field: ""},
		},
		{
			name:          "non-empty string with empty suffix",
			structWithTag: struct{ Field string `validate:"endswith:"` }{Field: "abc"},
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"endswith:def"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("endswith"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestContainsValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid contains",
			structWithTag: struct{ Field string `validate:"contains:bcd"` }{Field: "abcdef"},
		},
		{
			name:            "doesn't contain substring",
			structWithTag:   struct{ Field string `validate:"contains:xyz"` }{Field: "abcdef"},
			wantErrContains: "contains",
		},
		{
			name:          "empty string with empty contains",
			structWithTag: struct{ Field string `validate:"contains:"` }{Field: ""},
		},
		{
			name:          "non-empty string with empty contains",
			structWithTag: struct{ Field string `validate:"contains:"` }{Field: "abc"},
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"contains:bcd"` }{Field: 12345},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("contains"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestNotContainsValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid not contains",
			structWithTag: struct{ Field string `validate:"notcontains:xyz"` }{Field: "abcdef"},
		},
		{
			name:            "contains substring",
			structWithTag:   struct{ Field string `validate:"notcontains:bcd"` }{Field: "abcdef"},
			wantErrContains: "notcontains",
		},
		{
			name:          "empty string with empty notcontains",
			structWithTag: struct{ Field string `validate:"notcontains:"` }{Field: ""},
		},
		{
			name:          "non-empty string with empty notcontains",
			structWithTag: struct{ Field string `validate:"notcontains:"` }{Field: "abc"},
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"notcontains:bcd"` }{Field: 12345},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("notcontains"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestAsciiValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid ascii string",
			structWithTag: struct{ Field string `validate:"ascii"` }{Field: "abcDEF123!@#"},
		},
		{
			name:            "contains non-ascii characters",
			structWithTag:   struct{ Field string `validate:"ascii"` }{Field: "abc🍎def"},
			wantErrContains: "ascii",
		},
		{
			name:          "empty string",
			structWithTag: struct{ Field string `validate:"ascii"` }{Field: ""},
		},
		{
			name:            "type mismatch (int)",
			structWithTag:   struct{ Field int `validate:"ascii"` }{Field: 123},
			wantErrContains: "type_mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("ascii"))
			require.NoError(t, err)
			
			err = v.ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}