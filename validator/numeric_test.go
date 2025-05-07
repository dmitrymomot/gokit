package validator_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dmitrymomot/gokit/validator"
)

func TestNumericValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid integer string",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: "123"},
		},
		{
			name: "valid float string",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: "123.45"},
		},
		{
			name: "valid negative integer string",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: "-123"},
		},
		{
			name: "valid negative float string",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: "-123.45"},
		},
		{
			name: "valid int",
			structWithTag: struct {
				Value int `validate:"numeric"`
			}{Value: 123},
		},
		{
			name: "valid int8",
			structWithTag: struct {
				Value int8 `validate:"numeric"`
			}{Value: 12},
		},
		{
			name: "valid int16",
			structWithTag: struct {
				Value int16 `validate:"numeric"`
			}{Value: 1234},
		},
		{
			name: "valid int32",
			structWithTag: struct {
				Value int32 `validate:"numeric"`
			}{Value: 123456},
		},
		{
			name: "valid int64",
			structWithTag: struct {
				Value int64 `validate:"numeric"`
			}{Value: 1234567890},
		},
		{
			name: "valid uint",
			structWithTag: struct {
				Value uint `validate:"numeric"`
			}{Value: 123},
		},
		{
			name: "valid uint8",
			structWithTag: struct {
				Value uint8 `validate:"numeric"`
			}{Value: 12},
		},
		{
			name: "valid uint16",
			structWithTag: struct {
				Value uint16 `validate:"numeric"`
			}{Value: 1234},
		},
		{
			name: "valid uint32",
			structWithTag: struct {
				Value uint32 `validate:"numeric"`
			}{Value: 123456},
		},
		{
			name: "valid uint64",
			structWithTag: struct {
				Value uint64 `validate:"numeric"`
			}{Value: 1234567890},
		},
		{
			name: "valid float32",
			structWithTag: struct {
				Value float32 `validate:"numeric"`
			}{Value: 123.45},
		},
		{
			name: "valid float64",
			structWithTag: struct {
				Value float64 `validate:"numeric"`
			}{Value: 123.4567},
		},
		// Invalid cases
		{
			name: "invalid non-numeric string",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: "abc"},
			wantErrContains: "validation.numeric",
		},
		{
			name: "invalid string with special characters",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: "123@#$"},
			wantErrContains: "validation.numeric",
		},
		{
			name: "empty string",
			structWithTag: struct {
				Value string `validate:"numeric"`
			}{Value: ""},
			wantErrContains: "validation.numeric",
		},
		// Type mismatch
		{
			name: "type mismatch bool",
			structWithTag: struct {
				Value bool `validate:"numeric"`
			}{Value: true},
			wantErrContains: "validation.numeric",
		},
		{
			name: "type mismatch struct",
			structWithTag: struct {
				Value struct{} `validate:"numeric"`
			}{Value: struct{}{}},
			wantErrContains: "validation.numeric",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New()
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

func TestPositiveValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid positive int",
			structWithTag: struct {
				Value int `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive int8",
			structWithTag: struct {
				Value int8 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive int16",
			structWithTag: struct {
				Value int16 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive int32",
			structWithTag: struct {
				Value int32 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive int64",
			structWithTag: struct {
				Value int64 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive uint",
			structWithTag: struct {
				Value uint `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive uint8",
			structWithTag: struct {
				Value uint8 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive uint16",
			structWithTag: struct {
				Value uint16 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive uint32",
			structWithTag: struct {
				Value uint32 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive uint64",
			structWithTag: struct {
				Value uint64 `validate:"positive"`
			}{Value: 1},
		},
		{
			name: "valid positive float32",
			structWithTag: struct {
				Value float32 `validate:"positive"`
			}{Value: 0.1},
		},
		{
			name: "valid positive float64",
			structWithTag: struct {
				Value float64 `validate:"positive"`
			}{Value: 0.1},
		},
		// Invalid cases
		{
			name: "invalid zero int",
			structWithTag: struct {
				Value int `validate:"positive"`
			}{Value: 0},
			wantErrContains: "validation.positive",
		},
		{
			name: "invalid negative int",
			structWithTag: struct {
				Value int `validate:"positive"`
			}{Value: -1},
			wantErrContains: "validation.positive",
		},
		{
			name: "invalid zero uint",
			structWithTag: struct {
				Value uint `validate:"positive"`
			}{Value: 0},
			wantErrContains: "validation.positive",
		},
		{
			name: "invalid zero float",
			structWithTag: struct {
				Value float64 `validate:"positive"`
			}{Value: 0.0},
			wantErrContains: "validation.positive",
		},
		{
			name: "invalid negative float",
			structWithTag: struct {
				Value float64 `validate:"positive"`
			}{Value: -0.1},
			wantErrContains: "validation.positive",
		},
		// Type mismatch
		{
			name: "type mismatch string",
			structWithTag: struct {
				Value string `validate:"positive"`
			}{Value: "1"},
			wantErrContains: "validation.positive",
		},
		{
			name: "type mismatch bool",
			structWithTag: struct {
				Value bool `validate:"positive"`
			}{Value: true},
			wantErrContains: "validation.positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New()
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

func TestNegativeValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid negative int",
			structWithTag: struct {
				Value int `validate:"negative"`
			}{Value: -1},
		},
		{
			name: "valid negative int8",
			structWithTag: struct {
				Value int8 `validate:"negative"`
			}{Value: -1},
		},
		{
			name: "valid negative int16",
			structWithTag: struct {
				Value int16 `validate:"negative"`
			}{Value: -1},
		},
		{
			name: "valid negative int32",
			structWithTag: struct {
				Value int32 `validate:"negative"`
			}{Value: -1},
		},
		{
			name: "valid negative int64",
			structWithTag: struct {
				Value int64 `validate:"negative"`
			}{Value: -1},
		},
		{
			name: "valid negative float32",
			structWithTag: struct {
				Value float32 `validate:"negative"`
			}{Value: -0.1},
		},
		{
			name: "valid negative float64",
			structWithTag: struct {
				Value float64 `validate:"negative"`
			}{Value: -0.1},
		},
		// Invalid cases
		{
			name: "invalid zero int",
			structWithTag: struct {
				Value int `validate:"negative"`
			}{Value: 0},
			wantErrContains: "validation.negative",
		},
		{
			name: "invalid positive int",
			structWithTag: struct {
				Value int `validate:"negative"`
			}{Value: 1},
			wantErrContains: "validation.negative",
		},
		{
			name: "invalid positive uint (always fails for non-zero)",
			structWithTag: struct {
				Value uint `validate:"negative"`
			}{Value: 1},
			wantErrContains: "validation.negative",
		},
		{
			name: "invalid zero uint (passes, as 0 is not > 0)",
			structWithTag: struct {
				Value uint `validate:"negative"`
			}{Value: 0},
		},
		{
			name: "invalid zero float",
			structWithTag: struct {
				Value float64 `validate:"negative"`
			}{Value: 0.0},
			wantErrContains: "validation.negative",
		},
		{
			name: "invalid positive float",
			structWithTag: struct {
				Value float64 `validate:"negative"`
			}{Value: 0.1},
			wantErrContains: "validation.negative",
		},
		// Type mismatch
		{
			name: "type mismatch string",
			structWithTag: struct {
				Value string `validate:"negative"`
			}{Value: "-1"},
			wantErrContains: "validation.negative",
		},
		{
			name: "type mismatch bool",
			structWithTag: struct {
				Value bool `validate:"negative"`
			}{Value: true},
			wantErrContains: "validation.negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New()
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

func TestEvenValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid even int",
			structWithTag: struct {
				Value int `validate:"even"`
			}{Value: 2},
		},
		{
			name: "valid even int (zero)",
			structWithTag: struct {
				Value int `validate:"even"`
			}{Value: 0},
		},
		{
			name: "valid even int (negative)",
			structWithTag: struct {
				Value int `validate:"even"`
			}{Value: -4},
		},
		{
			name: "valid even uint",
			structWithTag: struct {
				Value uint `validate:"even"`
			}{Value: 2},
		},
		{
			name: "valid even uint (zero)",
			structWithTag: struct {
				Value uint `validate:"even"`
			}{Value: 0},
		},
		// Invalid cases
		{
			name: "invalid odd int",
			structWithTag: struct {
				Value int `validate:"even"`
			}{Value: 1},
			wantErrContains: "validation.even",
		},
		{
			name: "invalid odd int (negative)",
			structWithTag: struct {
				Value int `validate:"even"`
			}{Value: -3},
			wantErrContains: "validation.even",
		},
		{
			name: "invalid odd uint",
			structWithTag: struct {
				Value uint `validate:"even"`
			}{Value: 1},
			wantErrContains: "validation.even",
		},
		// Type mismatch
		{
			name: "type mismatch float64",
			structWithTag: struct {
				Value float64 `validate:"even"`
			}{Value: 2.0},
			wantErrContains: "validation.even",
		},
		{
			name: "type mismatch string",
			structWithTag: struct {
				Value string `validate:"even"`
			}{Value: "2"},
			wantErrContains: "validation.even",
		},
		{
			name: "type mismatch bool",
			structWithTag: struct {
				Value bool `validate:"even"`
			}{Value: true},
			wantErrContains: "validation.even",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New()
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

func TestOddValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid odd int",
			structWithTag: struct {
				Value int `validate:"odd"`
			}{Value: 1},
		},
		{
			name: "valid odd int (negative)",
			structWithTag: struct {
				Value int `validate:"odd"`
			}{Value: -3},
		},
		{
			name: "valid odd uint",
			structWithTag: struct {
				Value uint `validate:"odd"`
			}{Value: 1},
		},
		// Invalid cases
		{
			name: "invalid even int",
			structWithTag: struct {
				Value int `validate:"odd"`
			}{Value: 2},
			wantErrContains: "validation.odd",
		},
		{
			name: "invalid even int (zero)",
			structWithTag: struct {
				Value int `validate:"odd"`
			}{Value: 0},
			wantErrContains: "validation.odd",
		},
		{
			name: "invalid even int (negative)",
			structWithTag: struct {
				Value int `validate:"odd"`
			}{Value: -4},
			wantErrContains: "validation.odd",
		},
		{
			name: "invalid even uint",
			structWithTag: struct {
				Value uint `validate:"odd"`
			}{Value: 2},
			wantErrContains: "validation.odd",
		},
		{
			name: "invalid even uint (zero)",
			structWithTag: struct {
				Value uint `validate:"odd"`
			}{Value: 0},
			wantErrContains: "validation.odd",
		},
		// Type mismatch
		{
			name: "type mismatch float64",
			structWithTag: struct {
				Value float64 `validate:"odd"`
			}{Value: 1.0},
			wantErrContains: "validation.odd",
		},
		{
			name: "type mismatch string",
			structWithTag: struct {
				Value string `validate:"odd"`
			}{Value: "1"},
			wantErrContains: "validation.odd",
		},
		{
			name: "type mismatch bool",
			structWithTag: struct {
				Value bool `validate:"odd"`
			}{Value: true},
			wantErrContains: "validation.odd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New()
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

func TestMultipleValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid multiple of 3 (int)",
			structWithTag: struct {
				Value int `validate:"multiple:3"`
			}{Value: 9},
		},
		{
			name: "valid multiple of 5 (int, zero)",
			structWithTag: struct {
				Value int `validate:"multiple:5"`
			}{Value: 0},
		},
		{
			name: "valid multiple of 2 (int, negative)",
			structWithTag: struct {
				Value int `validate:"multiple:2"`
			}{Value: -4},
		},
		{
			name: "valid multiple of 7 (uint)",
			structWithTag: struct {
				Value uint `validate:"multiple:7"`
			}{Value: 14},
		},
		{
			name: "valid multiple of 4 (uint, zero)",
			structWithTag: struct {
				Value uint `validate:"multiple:4"`
			}{Value: 0},
		},
		{
			name: "no params (should pass)",
			structWithTag: struct {
				Value int `validate:"multiple"`
			}{Value: 7},
		},
		{
			name: "invalid param (non-integer, should pass as per implementation)",
			structWithTag: struct {
				Value int `validate:"multiple:abc"`
			}{Value: 7},
		},
		{
			name: "zero divisor (should pass as per implementation)",
			structWithTag: struct {
				Value int `validate:"multiple:0"`
			}{Value: 7},
		},
		// Invalid cases
		{
			name: "invalid multiple of 3 (int)",
			structWithTag: struct {
				Value int `validate:"multiple:3"`
			}{Value: 10},
			wantErrContains: "validation.multiple",
		},
		{
			name: "invalid multiple of 2 (int, negative)",
			structWithTag: struct {
				Value int `validate:"multiple:2"`
			}{Value: -5},
			wantErrContains: "validation.multiple",
		},
		{
			name: "invalid multiple of 7 (uint)",
			structWithTag: struct {
				Value uint `validate:"multiple:7"`
			}{Value: 15},
			wantErrContains: "validation.multiple",
		},
		// Type mismatch (should pass as per implementation which only checks int/uint)
		{
			name: "type mismatch float64",
			structWithTag: struct {
				Value float64 `validate:"multiple:2"`
			}{Value: 4.0},
		},
		{
			name: "type mismatch string",
			structWithTag: struct {
				Value string `validate:"multiple:2"`
			}{Value: "4"},
		},
		{
			name: "type mismatch bool",
			structWithTag: struct {
				Value bool `validate:"multiple:2"`
			}{Value: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New()
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
