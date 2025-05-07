package validator_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/validator"
	"github.com/stretchr/testify/require"
)

func TestBase64Validator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string // empty means expect no error
	}{
		{
			name:          "valid base64",
			structWithTag: struct{Data string `validate:"base64"`}{Data: "SGVsbG8gd29ybGQ="},
		},
		{
			name:          "invalid base64",
			structWithTag: struct{Data string `validate:"base64"`}{Data: "not_base64"},
			wantErrContains: "base64",
		},
		{
			name:          "empty string",
			structWithTag: struct{Data string `validate:"base64"`}{Data: ""},
			wantErrContains: "base64",
		},
		{
			name:          "type mismatch (int)",
			structWithTag: struct{Data int `validate:"base64"`}{Data: 123},
			wantErrContains: "type_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.NewValidator(nil).ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestJSONValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		structWithTag  any
		wantErrContains string // empty means expect no error
	}{
		{
			name:          "valid json object",
			structWithTag: struct{Data string `validate:"json"`}{Data: `{"key":"value"}`},
		},
		{
			name:          "valid json array",
			structWithTag: struct{Data string `validate:"json"`}{Data: `[1,2,3]`},
		},
		{
			name:          "invalid json",
			structWithTag: struct{Data string `validate:"json"`}{Data: `{key:value}`},
			wantErrContains: "json",
		},
		{
			name:          "empty string",
			structWithTag: struct{Data string `validate:"json"`}{Data: ""},
			wantErrContains: "json",
		},
		{
			name:          "type mismatch (bool)",
			structWithTag: struct{Data bool `validate:"json"`}{Data: true},
			wantErrContains: "type_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.NewValidator(nil).ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}

func TestSemverValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		structWithTag  any
		wantErrContains string // empty means expect no error
	}{
		{
			name:          "valid semver",
			structWithTag: struct{Version string `validate:"semver"`}{Version: "1.2.3"},
		},
		{
			name:          "valid semver with v",
			structWithTag: struct{Version string `validate:"semver"`}{Version: "v1.2.3"},
		},
		{
			name:          "valid semver with prerelease",
			structWithTag: struct{Version string `validate:"semver"`}{Version: "1.2.3-beta.1"},
		},
		{
			name:          "valid semver with build",
			structWithTag: struct{Version string `validate:"semver"`}{Version: "1.2.3+build.5"},
		},
		{
			name:          "invalid semver",
			structWithTag: struct{Version string `validate:"semver"`}{Version: "1.2"},
			wantErrContains: "semver",
		},
		{
			name:          "invalid semver letters",
			structWithTag: struct{Version string `validate:"semver"`}{Version: "abc"},
			wantErrContains: "semver",
		},
		{
			name:          "empty string",
			structWithTag: struct{Version string `validate:"semver"`}{Version: ""},
			wantErrContains: "semver",
		},
		{
			name:          "type mismatch (struct)",
			structWithTag: struct{Version struct{X int} `validate:"semver"`}{Version: struct{X int}{X: 1}},
			wantErrContains: "type_mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.NewValidator(nil).ValidateStruct(tt.structWithTag)
			if tt.wantErrContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			}
		})
	}
}
