package validator_test

import (
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/validator"
	"github.com/stretchr/testify/require"
)

func TestDateValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name: "valid date (YYYY-MM-DD)",
			structWithTag: struct{ Date string `validate:"date:2006-01-02"` }{Date: "2021-12-31"},
		},
		{
			name: "invalid date (bad format)",
			structWithTag: struct{ Date string `validate:"date:2006-01-02"` }{Date: "31-12-2021"},
			wantErrContains: "date",
		},
		{
			name: "empty string (omitempty first)",
			structWithTag: struct{ Date string `validate:"omitempty;date:2006-01-02"` }{Date: ""},
		},
		{
			name: "empty string (omitempty not first)",
			structWithTag: struct{ Date string `validate:"date:2006-01-02,omitempty"` }{Date: ""},
			wantErrContains: "date",
		},
		{
			name: "type mismatch (int)",
			structWithTag: struct{ Date int `validate:"date:2006-01-02"` }{Date: 123},
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

func TestPastDateValidator(t *testing.T) {
	t.Parallel()
	now := time.Now()
	past := now.AddDate(-1, 0, 0).Format("2006-01-02")
	future := now.AddDate(1, 0, 0).Format("2006-01-02")
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name: "valid past date",
			structWithTag: struct{ Date string `validate:"pastdate:2006-01-02"` }{Date: past},
		},
		{
			name: "invalid (future date)",
			structWithTag: struct{ Date string `validate:"pastdate:2006-01-02"` }{Date: future},
			wantErrContains: "pastdate",
		},
		{
			name: "invalid format",
			structWithTag: struct{ Date string `validate:"pastdate:2006-01-02"` }{Date: "notadate"},
			wantErrContains: "pastdate",
		},
		{
			name: "empty string (omitempty first)",
			structWithTag: struct{ Date string `validate:"omitempty;pastdate:2006-01-02"` }{Date: ""},
		},
		{
			name: "empty string (omitempty not first)",
			structWithTag: struct{ Date string `validate:"pastdate:2006-01-02,omitempty"` }{Date: ""},
			wantErrContains: "pastdate",
		},
		{
			name: "type mismatch (bool)",
			structWithTag: struct{ Date bool `validate:"pastdate:2006-01-02"` }{Date: true},
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

func TestFutureDateValidator(t *testing.T) {
	t.Parallel()
	now := time.Now()
	past := now.AddDate(-1, 0, 0).Format("2006-01-02")
	future := now.AddDate(1, 0, 0).Format("2006-01-02")
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name: "valid future date",
			structWithTag: struct{ Date string `validate:"futuredate:2006-01-02"` }{Date: future},
		},
		{
			name: "invalid (past date)",
			structWithTag: struct{ Date string `validate:"futuredate:2006-01-02"` }{Date: past},
			wantErrContains: "futuredate",
		},
		{
			name: "invalid format",
			structWithTag: struct{ Date string `validate:"futuredate:2006-01-02"` }{Date: "notadate"},
			wantErrContains: "futuredate",
		},
		{
			name: "empty string (omitempty first)",
			structWithTag: struct{ Date string `validate:"omitempty;futuredate:2006-01-02"` }{Date: ""},
		},
		{
			name: "empty string (omitempty not first)",
			structWithTag: struct{ Date string `validate:"futuredate:2006-01-02,omitempty"` }{Date: ""},
			wantErrContains: "futuredate",
		},
		{
			name: "type mismatch (struct)",
			structWithTag: struct{ Date struct{} `validate:"futuredate:2006-01-02"` }{Date: struct{}{}},
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

func TestWorkdayValidator(t *testing.T) {
	t.Parallel()
	monday := time.Date(2023, 5, 8, 0, 0, 0, 0, time.UTC).Format("2006-01-02") // Monday
	saturday := time.Date(2023, 5, 13, 0, 0, 0, 0, time.UTC).Format("2006-01-02") // Saturday
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name: "valid workday (Monday)",
			structWithTag: struct{ Date string `validate:"workday:2006-01-02"` }{Date: monday},
		},
		{
			name: "invalid (Saturday)",
			structWithTag: struct{ Date string `validate:"workday:2006-01-02"` }{Date: saturday},
			wantErrContains: "workday",
		},
		{
			name: "invalid format",
			structWithTag: struct{ Date string `validate:"workday:2006-01-02"` }{Date: "notadate"},
			wantErrContains: "workday",
		},
		{
			name: "empty string (omitempty first)",
			structWithTag: struct{ Date string `validate:"omitempty;workday:2006-01-02"` }{Date: ""},
		},
		{
			name: "empty string (omitempty not first)",
			structWithTag: struct{ Date string `validate:"workday:2006-01-02,omitempty"` }{Date: ""},
			wantErrContains: "workday",
		},
		{
			name: "type mismatch (float64)",
			structWithTag: struct{ Date float64 `validate:"workday:2006-01-02"` }{Date: 1.23},
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

func TestWeekendValidator(t *testing.T) {
	t.Parallel()
	sunday := time.Date(2023, 5, 14, 0, 0, 0, 0, time.UTC).Format("2006-01-02") // Sunday
	wednesday := time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02") // Wednesday
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name: "valid weekend (Sunday)",
			structWithTag: struct{ Date string `validate:"weekend:2006-01-02"` }{Date: sunday},
		},
		{
			name: "invalid (Wednesday)",
			structWithTag: struct{ Date string `validate:"weekend:2006-01-02"` }{Date: wednesday},
			wantErrContains: "weekend",
		},
		{
			name: "invalid format",
			structWithTag: struct{ Date string `validate:"weekend:2006-01-02"` }{Date: "notadate"},
			wantErrContains: "weekend",
		},
		{
			name: "empty string (omitempty first)",
			structWithTag: struct{ Date string `validate:"omitempty;weekend:2006-01-02"` }{Date: ""},
		},
		{
			name: "empty string (omitempty not first)",
			structWithTag: struct{ Date string `validate:"weekend:2006-01-02,omitempty"` }{Date: ""},
			wantErrContains: "weekend",
		},
		{
			name: "type mismatch (slice)",
			structWithTag: struct{ Date []string `validate:"weekend:2006-01-02"` }{Date: []string{"2023-05-14"}},
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
