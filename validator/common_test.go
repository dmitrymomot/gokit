package validator_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/validator"
	"github.com/stretchr/testify/require"
)

func TestRequiredValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "valid string",
			structWithTag: struct{ Field string `validate:"required"` }{Field: "value"},
		},
		{
			name:            "empty string",
			structWithTag:   struct{ Field string `validate:"required"` }{Field: ""},
			wantErrContains: "required",
		},
		{
			name:          "valid int",
			structWithTag: struct{ Field int `validate:"required"` }{Field: 10},
		},
		{
			name:            "zero int",
			structWithTag:   struct{ Field int `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid uint",
			structWithTag: struct{ Field uint `validate:"required"` }{Field: 10},
		},
		{
			name:            "zero uint",
			structWithTag:   struct{ Field uint `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid uint8",
			structWithTag: struct{ Field uint8 `validate:"required"` }{Field: 10},
		},
		{
			name:            "zero uint8",
			structWithTag:   struct{ Field uint8 `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid uint16",
			structWithTag: struct{ Field uint16 `validate:"required"` }{Field: 10},
		},
		{
			name:            "zero uint16",
			structWithTag:   struct{ Field uint16 `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid uint32",
			structWithTag: struct{ Field uint32 `validate:"required"` }{Field: 10},
		},
		{
			name:            "zero uint32",
			structWithTag:   struct{ Field uint32 `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid uint64",
			structWithTag: struct{ Field uint64 `validate:"required"` }{Field: 10},
		},
		{
			name:            "zero uint64",
			structWithTag:   struct{ Field uint64 `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid float32",
			structWithTag: struct{ Field float32 `validate:"required"` }{Field: 10.5},
		},
		{
			name:            "zero float32",
			structWithTag:   struct{ Field float32 `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid float64",
			structWithTag: struct{ Field float64 `validate:"required"` }{Field: 10.5},
		},
		{
			name:            "zero float64",
			structWithTag:   struct{ Field float64 `validate:"required"` }{Field: 0},
			wantErrContains: "required",
		},
		{
			name:          "valid bool",
			structWithTag: struct{ Field bool `validate:"required"` }{Field: true},
		},
		{
			name:            "false bool",
			structWithTag:   struct{ Field bool `validate:"required"` }{Field: false},
			wantErrContains: "required",
		},
		{
			name:          "valid slice",
			structWithTag: struct{ Field []string `validate:"required"` }{Field: []string{"item"}},
		},
		{
			name:            "empty slice",
			structWithTag:   struct{ Field []string `validate:"required"` }{Field: []string{}},
			wantErrContains: "required",
		},
		{
			name:            "nil slice",
			structWithTag:   struct{ Field []string `validate:"required"` }{Field: nil},
			wantErrContains: "required",
		},
		{
			name:          "valid map",
			structWithTag: struct{ Field map[string]string `validate:"required"` }{Field: map[string]string{"key": "value"}},
		},
		{
			name:            "empty map",
			structWithTag:   struct{ Field map[string]string `validate:"required"` }{Field: map[string]string{}},
			wantErrContains: "required",
		},
		{
			name:            "nil map",
			structWithTag:   struct{ Field map[string]string `validate:"required"` }{Field: nil},
			wantErrContains: "required",
		},
		{
			name:          "valid pointer",
			structWithTag: struct{ Field *string `validate:"required"` }{Field: func() *string { s := "value"; return &s }()},
		},
		{
			name:            "nil pointer",
			structWithTag:   struct{ Field *string `validate:"required"` }{Field: nil},
			wantErrContains: "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("required"))
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

func TestMaxValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string below max",
			structWithTag: struct{ Field string `validate:"max:10"` }{Field: "test"},
		},
		{
			name:          "string at max",
			structWithTag: struct{ Field string `validate:"max:4"` }{Field: "test"},
		},
		{
			name:            "string above max",
			structWithTag:   struct{ Field string `validate:"max:3"` }{Field: "test"},
			wantErrContains: "max",
		},
		{
			name:          "int below max",
			structWithTag: struct{ Field int `validate:"max:10"` }{Field: 5},
		},
		{
			name:          "int at max",
			structWithTag: struct{ Field int `validate:"max:5"` }{Field: 5},
		},
		{
			name:            "int above max",
			structWithTag:   struct{ Field int `validate:"max:4"` }{Field: 5},
			wantErrContains: "max",
		},
		{
			name:          "uint below max",
			structWithTag: struct{ Field uint `validate:"max:10"` }{Field: 5},
		},
		{
			name:          "uint at max",
			structWithTag: struct{ Field uint `validate:"max:5"` }{Field: 5},
		},
		{
			name:            "uint above max",
			structWithTag:   struct{ Field uint `validate:"max:4"` }{Field: 5},
			wantErrContains: "max",
		},
		{
			name:          "uint8 below max",
			structWithTag: struct{ Field uint8 `validate:"max:10"` }{Field: 5},
		},
		{
			name:          "uint8 at max",
			structWithTag: struct{ Field uint8 `validate:"max:5"` }{Field: 5},
		},
		{
			name:            "uint8 above max",
			structWithTag:   struct{ Field uint8 `validate:"max:4"` }{Field: 5},
			wantErrContains: "max",
		},
		{
			name:          "uint16 below max",
			structWithTag: struct{ Field uint16 `validate:"max:10"` }{Field: 5},
		},
		{
			name:          "uint16 at max",
			structWithTag: struct{ Field uint16 `validate:"max:5"` }{Field: 5},
		},
		{
			name:            "uint16 above max",
			structWithTag:   struct{ Field uint16 `validate:"max:4"` }{Field: 5},
			wantErrContains: "max",
		},
		{
			name:          "uint32 below max",
			structWithTag: struct{ Field uint32 `validate:"max:10"` }{Field: 5},
		},
		{
			name:          "uint32 at max",
			structWithTag: struct{ Field uint32 `validate:"max:5"` }{Field: 5},
		},
		{
			name:            "uint32 above max",
			structWithTag:   struct{ Field uint32 `validate:"max:4"` }{Field: 5},
			wantErrContains: "max",
		},
		{
			name:          "uint64 below max",
			structWithTag: struct{ Field uint64 `validate:"max:10"` }{Field: 5},
		},
		{
			name:          "uint64 at max",
			structWithTag: struct{ Field uint64 `validate:"max:5"` }{Field: 5},
		},
		{
			name:            "uint64 above max",
			structWithTag:   struct{ Field uint64 `validate:"max:4"` }{Field: 5},
			wantErrContains: "max",
		},
		{
			name:          "float32 below max",
			structWithTag: struct{ Field float32 `validate:"max:10.5"` }{Field: 5.5},
		},
		{
			name:          "float32 at max",
			structWithTag: struct{ Field float32 `validate:"max:5.5"` }{Field: 5.5},
		},
		{
			name:            "float32 above max",
			structWithTag:   struct{ Field float32 `validate:"max:5.4"` }{Field: 5.5},
			wantErrContains: "max",
		},
		{
			name:          "float below max",
			structWithTag: struct{ Field float64 `validate:"max:10.5"` }{Field: 5.5},
		},
		{
			name:          "float at max",
			structWithTag: struct{ Field float64 `validate:"max:5.5"` }{Field: 5.5},
		},
		{
			name:            "float above max",
			structWithTag:   struct{ Field float64 `validate:"max:5.4"` }{Field: 5.5},
			wantErrContains: "max",
		},
		{
			name:          "slice below max",
			structWithTag: struct{ Field []string `validate:"max:3"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice at max",
			structWithTag: struct{ Field []string `validate:"max:2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice above max",
			structWithTag:   struct{ Field []string `validate:"max:1"` }{Field: []string{"a", "b"}},
			wantErrContains: "max",
		},
		{
			name:          "map below max",
			structWithTag: struct{ Field map[string]string `validate:"max:3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map at max",
			structWithTag: struct{ Field map[string]string `validate:"max:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map above max",
			structWithTag:   struct{ Field map[string]string `validate:"max:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "max",
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"max:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"max"` }{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("max"))
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

func TestMinValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string above min",
			structWithTag: struct{ Field string `validate:"min:3"` }{Field: "test"},
		},
		{
			name:          "string at min",
			structWithTag: struct{ Field string `validate:"min:4"` }{Field: "test"},
		},
		{
			name:            "string below min",
			structWithTag:   struct{ Field string `validate:"min:5"` }{Field: "test"},
			wantErrContains: "min",
		},
		{
			name:          "int above min",
			structWithTag: struct{ Field int `validate:"min:3"` }{Field: 5},
		},
		{
			name:          "int at min",
			structWithTag: struct{ Field int `validate:"min:5"` }{Field: 5},
		},
		{
			name:            "int below min",
			structWithTag:   struct{ Field int `validate:"min:6"` }{Field: 5},
			wantErrContains: "min",
		},
		{
			name:          "uint above min",
			structWithTag: struct{ Field uint `validate:"min:3"` }{Field: 5},
		},
		{
			name:          "uint at min",
			structWithTag: struct{ Field uint `validate:"min:5"` }{Field: 5},
		},
		{
			name:            "uint below min",
			structWithTag:   struct{ Field uint `validate:"min:6"` }{Field: 5},
			wantErrContains: "min",
		},
		{
			name:          "uint8 above min",
			structWithTag: struct{ Field uint8 `validate:"min:3"` }{Field: 5},
		},
		{
			name:          "uint8 at min",
			structWithTag: struct{ Field uint8 `validate:"min:5"` }{Field: 5},
		},
		{
			name:            "uint8 below min",
			structWithTag:   struct{ Field uint8 `validate:"min:6"` }{Field: 5},
			wantErrContains: "min",
		},
		{
			name:          "uint16 above min",
			structWithTag: struct{ Field uint16 `validate:"min:3"` }{Field: 5},
		},
		{
			name:          "uint16 at min",
			structWithTag: struct{ Field uint16 `validate:"min:5"` }{Field: 5},
		},
		{
			name:            "uint16 below min",
			structWithTag:   struct{ Field uint16 `validate:"min:6"` }{Field: 5},
			wantErrContains: "min",
		},
		{
			name:          "uint32 above min",
			structWithTag: struct{ Field uint32 `validate:"min:3"` }{Field: 5},
		},
		{
			name:          "uint32 at min",
			structWithTag: struct{ Field uint32 `validate:"min:5"` }{Field: 5},
		},
		{
			name:            "uint32 below min",
			structWithTag:   struct{ Field uint32 `validate:"min:6"` }{Field: 5},
			wantErrContains: "min",
		},
		{
			name:          "uint64 above min",
			structWithTag: struct{ Field uint64 `validate:"min:3"` }{Field: 5},
		},
		{
			name:          "uint64 at min",
			structWithTag: struct{ Field uint64 `validate:"min:5"` }{Field: 5},
		},
		{
			name:            "uint64 below min",
			structWithTag:   struct{ Field uint64 `validate:"min:6"` }{Field: 5},
			wantErrContains: "min",
		},
		{
			name:          "float32 above min",
			structWithTag: struct{ Field float32 `validate:"min:3.5"` }{Field: 5.5},
		},
		{
			name:          "float32 at min",
			structWithTag: struct{ Field float32 `validate:"min:5.5"` }{Field: 5.5},
		},
		{
			name:            "float32 below min",
			structWithTag:   struct{ Field float32 `validate:"min:5.6"` }{Field: 5.5},
			wantErrContains: "min",
		},
		{
			name:          "float above min",
			structWithTag: struct{ Field float64 `validate:"min:3.5"` }{Field: 5.5},
		},
		{
			name:          "float at min",
			structWithTag: struct{ Field float64 `validate:"min:5.5"` }{Field: 5.5},
		},
		{
			name:            "float below min",
			structWithTag:   struct{ Field float64 `validate:"min:5.6"` }{Field: 5.5},
			wantErrContains: "min",
		},
		{
			name:          "slice above min",
			structWithTag: struct{ Field []string `validate:"min:1"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice at min",
			structWithTag: struct{ Field []string `validate:"min:2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice below min",
			structWithTag:   struct{ Field []string `validate:"min:3"` }{Field: []string{"a", "b"}},
			wantErrContains: "min",
		},
		{
			name:          "map above min",
			structWithTag: struct{ Field map[string]string `validate:"min:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map at min",
			structWithTag: struct{ Field map[string]string `validate:"min:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map below min",
			structWithTag:   struct{ Field map[string]string `validate:"min:3"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "min",
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"min:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"min"` }{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("min"))
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

func TestRangeValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string within range",
			structWithTag: struct{ Field string `validate:"range:2,10"` }{Field: "test"},
		},
		{
			name:          "string at min range",
			structWithTag: struct{ Field string `validate:"range:4,10"` }{Field: "test"},
		},
		{
			name:          "string at max range",
			structWithTag: struct{ Field string `validate:"range:2,4"` }{Field: "test"},
		},
		{
			name:            "string below range",
			structWithTag:   struct{ Field string `validate:"range:5,10"` }{Field: "test"},
			wantErrContains: "range",
		},
		{
			name:            "string above range",
			structWithTag:   struct{ Field string `validate:"range:2,3"` }{Field: "test"},
			wantErrContains: "range",
		},
		{
			name:          "int within range",
			structWithTag: struct{ Field int `validate:"range:3,10"` }{Field: 5},
		},
		{
			name:          "int at min range",
			structWithTag: struct{ Field int `validate:"range:5,10"` }{Field: 5},
		},
		{
			name:          "int at max range",
			structWithTag: struct{ Field int `validate:"range:1,5"` }{Field: 5},
		},
		{
			name:            "int below range",
			structWithTag:   struct{ Field int `validate:"range:6,10"` }{Field: 5},
			wantErrContains: "range",
		},
		{
			name:            "int above range",
			structWithTag:   struct{ Field int `validate:"range:1,4"` }{Field: 5},
			wantErrContains: "range",
		},
		{
			name:          "float within range",
			structWithTag: struct{ Field float64 `validate:"range:3.5,10.5"` }{Field: 5.5},
		},
		{
			name:          "float at min range",
			structWithTag: struct{ Field float64 `validate:"range:5.5,10.5"` }{Field: 5.5},
		},
		{
			name:          "float at max range",
			structWithTag: struct{ Field float64 `validate:"range:3.5,5.5"` }{Field: 5.5},
		},
		{
			name:            "float below range",
			structWithTag:   struct{ Field float64 `validate:"range:5.6,10.5"` }{Field: 5.5},
			wantErrContains: "range",
		},
		{
			name:            "float above range",
			structWithTag:   struct{ Field float64 `validate:"range:3.5,5.4"` }{Field: 5.5},
			wantErrContains: "range",
		},
		{
			name:          "slice within range",
			structWithTag: struct{ Field []string `validate:"range:1,3"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice at min range",
			structWithTag: struct{ Field []string `validate:"range:2,3"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice at max range",
			structWithTag: struct{ Field []string `validate:"range:1,2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice below range",
			structWithTag:   struct{ Field []string `validate:"range:3,4"` }{Field: []string{"a", "b"}},
			wantErrContains: "range",
		},
		{
			name:            "slice above range",
			structWithTag:   struct{ Field []string `validate:"range:0,1"` }{Field: []string{"a", "b"}},
			wantErrContains: "range",
		},
		{
			name:          "map within range",
			structWithTag: struct{ Field map[string]string `validate:"range:1,3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map at min range",
			structWithTag: struct{ Field map[string]string `validate:"range:2,3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map at max range",
			structWithTag: struct{ Field map[string]string `validate:"range:1,2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map below range",
			structWithTag:   struct{ Field map[string]string `validate:"range:3,4"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "range",
		},
		{
			name:            "map above range",
			structWithTag:   struct{ Field map[string]string `validate:"range:0,1"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "range",
		},
		{
			name:          "uint within range",
			structWithTag: struct{ Field uint `validate:"range:3,10"` }{Field: 5},
		},
		{
			name:          "uint at min range",
			structWithTag: struct{ Field uint `validate:"range:5,10"` }{Field: 5},
		},
		{
			name:          "uint at max range",
			structWithTag: struct{ Field uint `validate:"range:1,5"` }{Field: 5},
		},
		{
			name:            "uint below range",
			structWithTag:   struct{ Field uint `validate:"range:6,10"` }{Field: 5},
			wantErrContains: "range",
		},
		{
			name:            "uint above range",
			structWithTag:   struct{ Field uint `validate:"range:1,4"` }{Field: 5},
			wantErrContains: "range",
		},
		{
			name:          "invalid parameters",
			structWithTag: struct{ Field string `validate:"range:invalid,10"` }{Field: "test"},
		},
		{
			name:          "missing parameters",
			structWithTag: struct{ Field string `validate:"range:5"` }{Field: "test"},
		},
		{
			name:          "swapped min/max",
			structWithTag: struct{ Field int `validate:"range:10,5"` }{Field: 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("range"))
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

func TestLengthValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string with correct length",
			structWithTag: struct{ Field string `validate:"length:4"` }{Field: "test"},
		},
		{
			name:            "string with incorrect length",
			structWithTag:   struct{ Field string `validate:"length:5"` }{Field: "test"},
			wantErrContains: "length",
		},
		{
			name:          "slice with correct length",
			structWithTag: struct{ Field []string `validate:"length:2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice with incorrect length",
			structWithTag:   struct{ Field []string `validate:"length:3"` }{Field: []string{"a", "b"}},
			wantErrContains: "length",
		},
		{
			name:          "empty slice with zero length",
			structWithTag: struct{ Field []string `validate:"length:0"` }{Field: []string{}},
		},
		{
			name:          "nil slice with zero length",
			structWithTag: struct{ Field []string `validate:"length:0"` }{Field: nil},
		},
		{
			name:          "map with correct length",
			structWithTag: struct{ Field map[string]string `validate:"length:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map with incorrect length",
			structWithTag:   struct{ Field map[string]string `validate:"length:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "length",
		},
		{
			name:          "empty map with zero length",
			structWithTag: struct{ Field map[string]string `validate:"length:0"` }{Field: map[string]string{}},
		},
		{
			name:          "nil map with zero length",
			structWithTag: struct{ Field map[string]string `validate:"length:0"` }{Field: nil},
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"length:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"length"` }{Field: "test"},
		},
		{
			name:          "unsupported type",
			structWithTag: struct{ Field int `validate:"length:4"` }{Field: 1234},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("length"))
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

func TestBetweenValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string within range",
			structWithTag: struct{ Field string `validate:"between:2,10"` }{Field: "test"},
		},
		{
			name:          "string at min range",
			structWithTag: struct{ Field string `validate:"between:4,10"` }{Field: "test"},
		},
		{
			name:          "string at max range",
			structWithTag: struct{ Field string `validate:"between:2,4"` }{Field: "test"},
		},
		{
			name:            "string below range",
			structWithTag:   struct{ Field string `validate:"between:5,10"` }{Field: "test"},
			wantErrContains: "between",
		},
		{
			name:            "string above range",
			structWithTag:   struct{ Field string `validate:"between:2,3"` }{Field: "test"},
			wantErrContains: "between",
		},
		{
			name:          "int within range",
			structWithTag: struct{ Field int `validate:"between:3,10"` }{Field: 5},
		},
		{
			name:          "int at min range",
			structWithTag: struct{ Field int `validate:"between:5,10"` }{Field: 5},
		},
		{
			name:          "int at max range",
			structWithTag: struct{ Field int `validate:"between:1,5"` }{Field: 5},
		},
		{
			name:            "int below range",
			structWithTag:   struct{ Field int `validate:"between:6,10"` }{Field: 5},
			wantErrContains: "between",
		},
		{
			name:            "int above range",
			structWithTag:   struct{ Field int `validate:"between:1,4"` }{Field: 5},
			wantErrContains: "between",
		},
		{
			name:          "float within range",
			structWithTag: struct{ Field float64 `validate:"between:3.5,10.5"` }{Field: 5.5},
		},
		{
			name:          "float at min range",
			structWithTag: struct{ Field float64 `validate:"between:5.5,10.5"` }{Field: 5.5},
		},
		{
			name:          "float at max range",
			structWithTag: struct{ Field float64 `validate:"between:3.5,5.5"` }{Field: 5.5},
		},
		{
			name:            "float below range",
			structWithTag:   struct{ Field float64 `validate:"between:5.6,10.5"` }{Field: 5.5},
			wantErrContains: "between",
		},
		{
			name:            "float above range",
			structWithTag:   struct{ Field float64 `validate:"between:3.5,5.4"` }{Field: 5.5},
			wantErrContains: "between",
		},
		{
			name:          "slice within range",
			structWithTag: struct{ Field []string `validate:"between:1,3"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice at min range",
			structWithTag: struct{ Field []string `validate:"between:2,3"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice at max range",
			structWithTag: struct{ Field []string `validate:"between:1,2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice below range",
			structWithTag:   struct{ Field []string `validate:"between:3,4"` }{Field: []string{"a", "b"}},
			wantErrContains: "between",
		},
		{
			name:            "slice above range",
			structWithTag:   struct{ Field []string `validate:"between:0,1"` }{Field: []string{"a", "b"}},
			wantErrContains: "between",
		},
		{
			name:          "map within range",
			structWithTag: struct{ Field map[string]string `validate:"between:1,3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map at min range",
			structWithTag: struct{ Field map[string]string `validate:"between:2,3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map at max range",
			structWithTag: struct{ Field map[string]string `validate:"between:1,2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map below range",
			structWithTag:   struct{ Field map[string]string `validate:"between:3,4"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "between",
		},
		{
			name:            "map above range",
			structWithTag:   struct{ Field map[string]string `validate:"between:0,1"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "between",
		},
		{
			name:          "uint within range",
			structWithTag: struct{ Field uint `validate:"between:3,10"` }{Field: 5},
		},
		{
			name:          "uint at min range",
			structWithTag: struct{ Field uint `validate:"between:5,10"` }{Field: 5},
		},
		{
			name:          "uint at max range",
			structWithTag: struct{ Field uint `validate:"between:1,5"` }{Field: 5},
		},
		{
			name:            "uint below range",
			structWithTag:   struct{ Field uint `validate:"between:6,10"` }{Field: 5},
			wantErrContains: "between",
		},
		{
			name:            "uint above range",
			structWithTag:   struct{ Field uint `validate:"between:1,4"` }{Field: 5},
			wantErrContains: "between",
		},
		{
			name:          "uint32 within range",
			structWithTag: struct{ Field uint32 `validate:"between:3,10"` }{Field: 5},
		},
		{
			name:          "float32 within range",
			structWithTag: struct{ Field float32 `validate:"between:3.5,10.5"` }{Field: 5.5},
		},
		{
			name:          "invalid parameters",
			structWithTag: struct{ Field string `validate:"between:invalid,10"` }{Field: "test"},
		},
		{
			name:          "missing parameters",
			structWithTag: struct{ Field string `validate:"between:5"` }{Field: "test"},
		},
		{
			name:          "swapped min/max",
			structWithTag: struct{ Field int `validate:"between:10,5"` }{Field: 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("between"))
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

func TestEqualValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string equal",
			structWithTag: struct{ Field string `validate:"eq:test"` }{Field: "test"},
		},
		{
			name:            "string not equal",
			structWithTag:   struct{ Field string `validate:"eq:other"` }{Field: "test"},
			wantErrContains: "eq",
		},
		{
			name:          "int equal",
			structWithTag: struct{ Field int `validate:"eq:5"` }{Field: 5},
		},
		{
			name:            "int not equal",
			structWithTag:   struct{ Field int `validate:"eq:6"` }{Field: 5},
			wantErrContains: "eq",
		},
		{
			name:          "float equal",
			structWithTag: struct{ Field float64 `validate:"eq:5.5"` }{Field: 5.5},
		},
		{
			name:            "float not equal",
			structWithTag:   struct{ Field float64 `validate:"eq:5.6"` }{Field: 5.5},
			wantErrContains: "eq",
		},
		{
			name:          "uint equal",
			structWithTag: struct{ Field uint `validate:"eq:5"` }{Field: 5},
		},
		{
			name:            "uint not equal",
			structWithTag:   struct{ Field uint `validate:"eq:6"` }{Field: 5},
			wantErrContains: "eq",
		},
		{
			name:          "uint32 equal",
			structWithTag: struct{ Field uint32 `validate:"eq:5"` }{Field: 5},
		},
		{
			name:            "uint32 not equal",
			structWithTag:   struct{ Field uint32 `validate:"eq:6"` }{Field: 5},
			wantErrContains: "eq",
		},
		{
			name:          "float32 equal",
			structWithTag: struct{ Field float32 `validate:"eq:5.5"` }{Field: 5.5},
		},
		{
			name:            "float32 not equal",
			structWithTag:   struct{ Field float32 `validate:"eq:5.6"` }{Field: 5.5},
			wantErrContains: "eq",
		},
		{
			name:          "boolean equal",
			structWithTag: struct{ Field bool `validate:"eq:true"` }{Field: true},
		},
		{
			name:            "boolean not equal",
			structWithTag:   struct{ Field bool `validate:"eq:true"` }{Field: false},
			wantErrContains: "eq",
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"eq"` }{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("eq"))
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

func TestNotEqualValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string not equal",
			structWithTag: struct{ Field string `validate:"ne:other"` }{Field: "test"},
		},
		{
			name:            "string equal",
			structWithTag:   struct{ Field string `validate:"ne:test"` }{Field: "test"},
			wantErrContains: "ne",
		},
		{
			name:          "int not equal",
			structWithTag: struct{ Field int `validate:"ne:6"` }{Field: 5},
		},
		{
			name:            "int equal",
			structWithTag:   struct{ Field int `validate:"ne:5"` }{Field: 5},
			wantErrContains: "ne",
		},
		{
			name:          "float not equal",
			structWithTag: struct{ Field float64 `validate:"ne:5.6"` }{Field: 5.5},
		},
		{
			name:            "float equal",
			structWithTag:   struct{ Field float64 `validate:"ne:5.5"` }{Field: 5.5},
			wantErrContains: "ne",
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"ne"` }{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("ne"))
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

func TestLessThanValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string length less than",
			structWithTag: struct{ Field string `validate:"lt:5"` }{Field: "test"},
		},
		{
			name:            "string length equal",
			structWithTag:   struct{ Field string `validate:"lt:4"` }{Field: "test"},
			wantErrContains: "lt",
		},
		{
			name:            "string length greater than",
			structWithTag:   struct{ Field string `validate:"lt:3"` }{Field: "test"},
			wantErrContains: "lt",
		},
		{
			name:          "int less than",
			structWithTag: struct{ Field int `validate:"lt:6"` }{Field: 5},
		},
		{
			name:            "int equal",
			structWithTag:   struct{ Field int `validate:"lt:5"` }{Field: 5},
			wantErrContains: "lt",
		},
		{
			name:            "int greater than",
			structWithTag:   struct{ Field int `validate:"lt:4"` }{Field: 5},
			wantErrContains: "lt",
		},
		{
			name:          "float less than",
			structWithTag: struct{ Field float64 `validate:"lt:5.6"` }{Field: 5.5},
		},
		{
			name:            "float equal",
			structWithTag:   struct{ Field float64 `validate:"lt:5.5"` }{Field: 5.5},
			wantErrContains: "lt",
		},
		{
			name:            "float greater than",
			structWithTag:   struct{ Field float64 `validate:"lt:5.4"` }{Field: 5.5},
			wantErrContains: "lt",
		},
		{
			name:          "slice less than",
			structWithTag: struct{ Field []string `validate:"lt:3"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice equal",
			structWithTag:   struct{ Field []string `validate:"lt:2"` }{Field: []string{"a", "b"}},
			wantErrContains: "lt",
		},
		{
			name:            "slice greater than",
			structWithTag:   struct{ Field []string `validate:"lt:1"` }{Field: []string{"a", "b"}},
			wantErrContains: "lt",
		},
		{
			name:          "map less than",
			structWithTag: struct{ Field map[string]string `validate:"lt:3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map equal",
			structWithTag:   struct{ Field map[string]string `validate:"lt:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "lt",
		},
		{
			name:            "map greater than",
			structWithTag:   struct{ Field map[string]string `validate:"lt:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "lt",
		},
		{
			name:          "uint less than",
			structWithTag: struct{ Field uint `validate:"lt:6"` }{Field: 5},
		},
		{
			name:            "uint equal",
			structWithTag:   struct{ Field uint `validate:"lt:5"` }{Field: 5},
			wantErrContains: "lt",
		},
		{
			name:            "uint greater than",
			structWithTag:   struct{ Field uint `validate:"lt:4"` }{Field: 5},
			wantErrContains: "lt",
		},
		{
			name:          "uint32 less than",
			structWithTag: struct{ Field uint32 `validate:"lt:6"` }{Field: 5},
		},
		{
			name:          "float32 less than",
			structWithTag: struct{ Field float32 `validate:"lt:5.6"` }{Field: 5.5},
		},
		{
			name:            "float32 equal",
			structWithTag:   struct{ Field float32 `validate:"lt:5.5"` }{Field: 5.5},
			wantErrContains: "lt",
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"lt:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"lt"` }{Field: "test"},
		},
		{
			name:          "unsupported type",
			structWithTag: struct{ Field bool `validate:"lt:true"` }{Field: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("lt"))
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

func TestLessThanOrEqualValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string length less than",
			structWithTag: struct{ Field string `validate:"lte:5"` }{Field: "test"},
		},
		{
			name:          "string length equal",
			structWithTag: struct{ Field string `validate:"lte:4"` }{Field: "test"},
		},
		{
			name:            "string length greater than",
			structWithTag:   struct{ Field string `validate:"lte:3"` }{Field: "test"},
			wantErrContains: "lte",
		},
		{
			name:          "int less than",
			structWithTag: struct{ Field int `validate:"lte:6"` }{Field: 5},
		},
		{
			name:          "int equal",
			structWithTag: struct{ Field int `validate:"lte:5"` }{Field: 5},
		},
		{
			name:            "int greater than",
			structWithTag:   struct{ Field int `validate:"lte:4"` }{Field: 5},
			wantErrContains: "lte",
		},
		{
			name:          "float less than",
			structWithTag: struct{ Field float64 `validate:"lte:5.6"` }{Field: 5.5},
		},
		{
			name:          "float equal",
			structWithTag: struct{ Field float64 `validate:"lte:5.5"` }{Field: 5.5},
		},
		{
			name:            "float greater than",
			structWithTag:   struct{ Field float64 `validate:"lte:5.4"` }{Field: 5.5},
			wantErrContains: "lte",
		},
		{
			name:          "slice less than",
			structWithTag: struct{ Field []string `validate:"lte:3"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice equal",
			structWithTag: struct{ Field []string `validate:"lte:2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice greater than",
			structWithTag:   struct{ Field []string `validate:"lte:1"` }{Field: []string{"a", "b"}},
			wantErrContains: "lte",
		},
		{
			name:          "map less than",
			structWithTag: struct{ Field map[string]string `validate:"lte:3"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map equal",
			structWithTag: struct{ Field map[string]string `validate:"lte:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map greater than",
			structWithTag:   struct{ Field map[string]string `validate:"lte:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "lte",
		},
		{
			name:          "uint less than",
			structWithTag: struct{ Field uint `validate:"lte:6"` }{Field: 5},
		},
		{
			name:          "uint equal",
			structWithTag: struct{ Field uint `validate:"lte:5"` }{Field: 5},
		},
		{
			name:            "uint greater than",
			structWithTag:   struct{ Field uint `validate:"lte:4"` }{Field: 5},
			wantErrContains: "lte",
		},
		{
			name:          "uint32 less than",
			structWithTag: struct{ Field uint32 `validate:"lte:6"` }{Field: 5},
		},
		{
			name:          "uint32 equal",
			structWithTag: struct{ Field uint32 `validate:"lte:5"` }{Field: 5},
		},
		{
			name:          "float32 less than",
			structWithTag: struct{ Field float32 `validate:"lte:5.6"` }{Field: 5.5},
		},
		{
			name:          "float32 equal",
			structWithTag: struct{ Field float32 `validate:"lte:5.5"` }{Field: 5.5},
		},
		{
			name:            "float32 greater than",
			structWithTag:   struct{ Field float32 `validate:"lte:5.4"` }{Field: 5.5},
			wantErrContains: "lte",
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"lte:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"lte"` }{Field: "test"},
		},
		{
			name:          "unsupported type",
			structWithTag: struct{ Field bool `validate:"lte:true"` }{Field: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("lte"))
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

func TestGreaterThanValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string length greater than",
			structWithTag: struct{ Field string `validate:"gt:3"` }{Field: "test"},
		},
		{
			name:            "string length equal",
			structWithTag:   struct{ Field string `validate:"gt:4"` }{Field: "test"},
			wantErrContains: "gt",
		},
		{
			name:            "string length less than",
			structWithTag:   struct{ Field string `validate:"gt:5"` }{Field: "test"},
			wantErrContains: "gt",
		},
		{
			name:          "int greater than",
			structWithTag: struct{ Field int `validate:"gt:4"` }{Field: 5},
		},
		{
			name:            "int equal",
			structWithTag:   struct{ Field int `validate:"gt:5"` }{Field: 5},
			wantErrContains: "gt",
		},
		{
			name:            "int less than",
			structWithTag:   struct{ Field int `validate:"gt:6"` }{Field: 5},
			wantErrContains: "gt",
		},
		{
			name:          "float greater than",
			structWithTag: struct{ Field float64 `validate:"gt:5.4"` }{Field: 5.5},
		},
		{
			name:            "float equal",
			structWithTag:   struct{ Field float64 `validate:"gt:5.5"` }{Field: 5.5},
			wantErrContains: "gt",
		},
		{
			name:            "float less than",
			structWithTag:   struct{ Field float64 `validate:"gt:5.6"` }{Field: 5.5},
			wantErrContains: "gt",
		},
		{
			name:          "slice greater than",
			structWithTag: struct{ Field []string `validate:"gt:1"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice equal",
			structWithTag:   struct{ Field []string `validate:"gt:2"` }{Field: []string{"a", "b"}},
			wantErrContains: "gt",
		},
		{
			name:            "slice less than",
			structWithTag:   struct{ Field []string `validate:"gt:3"` }{Field: []string{"a", "b"}},
			wantErrContains: "gt",
		},
		{
			name:          "map greater than",
			structWithTag: struct{ Field map[string]string `validate:"gt:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map equal",
			structWithTag:   struct{ Field map[string]string `validate:"gt:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "gt",
		},
		{
			name:            "map less than",
			structWithTag:   struct{ Field map[string]string `validate:"gt:3"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "gt",
		},
		{
			name:          "uint greater than",
			structWithTag: struct{ Field uint `validate:"gt:4"` }{Field: 5},
		},
		{
			name:            "uint equal",
			structWithTag:   struct{ Field uint `validate:"gt:5"` }{Field: 5},
			wantErrContains: "gt",
		},
		{
			name:            "uint less than",
			structWithTag:   struct{ Field uint `validate:"gt:6"` }{Field: 5},
			wantErrContains: "gt",
		},
		{
			name:          "uint32 greater than",
			structWithTag: struct{ Field uint32 `validate:"gt:4"` }{Field: 5},
		},
		{
			name:          "float32 greater than",
			structWithTag: struct{ Field float32 `validate:"gt:5.4"` }{Field: 5.5},
		},
		{
			name:            "float32 equal",
			structWithTag:   struct{ Field float32 `validate:"gt:5.5"` }{Field: 5.5},
			wantErrContains: "gt",
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"gt:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"gt"` }{Field: "test"},
		},
		{
			name:          "unsupported type",
			structWithTag: struct{ Field bool `validate:"gt:true"` }{Field: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("gt"))
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

func TestGreaterThanOrEqualValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string length greater than",
			structWithTag: struct{ Field string `validate:"gte:3"` }{Field: "test"},
		},
		{
			name:          "string length equal",
			structWithTag: struct{ Field string `validate:"gte:4"` }{Field: "test"},
		},
		{
			name:            "string length less than",
			structWithTag:   struct{ Field string `validate:"gte:5"` }{Field: "test"},
			wantErrContains: "gte",
		},
		{
			name:          "int greater than",
			structWithTag: struct{ Field int `validate:"gte:4"` }{Field: 5},
		},
		{
			name:          "int equal",
			structWithTag: struct{ Field int `validate:"gte:5"` }{Field: 5},
		},
		{
			name:            "int less than",
			structWithTag:   struct{ Field int `validate:"gte:6"` }{Field: 5},
			wantErrContains: "gte",
		},
		{
			name:          "float greater than",
			structWithTag: struct{ Field float64 `validate:"gte:5.4"` }{Field: 5.5},
		},
		{
			name:          "float equal",
			structWithTag: struct{ Field float64 `validate:"gte:5.5"` }{Field: 5.5},
		},
		{
			name:            "float less than",
			structWithTag:   struct{ Field float64 `validate:"gte:5.6"` }{Field: 5.5},
			wantErrContains: "gte",
		},
		{
			name:          "slice greater than",
			structWithTag: struct{ Field []string `validate:"gte:1"` }{Field: []string{"a", "b"}},
		},
		{
			name:          "slice equal",
			structWithTag: struct{ Field []string `validate:"gte:2"` }{Field: []string{"a", "b"}},
		},
		{
			name:            "slice less than",
			structWithTag:   struct{ Field []string `validate:"gte:3"` }{Field: []string{"a", "b"}},
			wantErrContains: "gte",
		},
		{
			name:          "map greater than",
			structWithTag: struct{ Field map[string]string `validate:"gte:1"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:          "map equal",
			structWithTag: struct{ Field map[string]string `validate:"gte:2"` }{Field: map[string]string{"a": "1", "b": "2"}},
		},
		{
			name:            "map less than",
			structWithTag:   struct{ Field map[string]string `validate:"gte:3"` }{Field: map[string]string{"a": "1", "b": "2"}},
			wantErrContains: "gte",
		},
		{
			name:          "uint greater than",
			structWithTag: struct{ Field uint `validate:"gte:4"` }{Field: 5},
		},
		{
			name:          "uint equal",
			structWithTag: struct{ Field uint `validate:"gte:5"` }{Field: 5},
		},
		{
			name:            "uint less than",
			structWithTag:   struct{ Field uint `validate:"gte:6"` }{Field: 5},
			wantErrContains: "gte",
		},
		{
			name:          "uint32 greater than",
			structWithTag: struct{ Field uint32 `validate:"gte:4"` }{Field: 5},
		},
		{
			name:          "uint32 equal",
			structWithTag: struct{ Field uint32 `validate:"gte:5"` }{Field: 5},
		},
		{
			name:          "float32 greater than",
			structWithTag: struct{ Field float32 `validate:"gte:5.4"` }{Field: 5.5},
		},
		{
			name:          "float32 equal",
			structWithTag: struct{ Field float32 `validate:"gte:5.5"` }{Field: 5.5},
		},
		{
			name:            "float32 less than",
			structWithTag:   struct{ Field float32 `validate:"gte:5.6"` }{Field: 5.5},
			wantErrContains: "gte",
		},
		{
			name:          "invalid parameter",
			structWithTag: struct{ Field string `validate:"gte:invalid"` }{Field: "test"},
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"gte"` }{Field: "test"},
		},
		{
			name:          "unsupported type",
			structWithTag: struct{ Field bool `validate:"gte:true"` }{Field: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("gte"))
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

func TestInValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string in list",
			structWithTag: struct{ Field string `validate:"in:test,example,sample"` }{Field: "test"},
		},
		{
			name:            "string not in list",
			structWithTag:   struct{ Field string `validate:"in:example,sample"` }{Field: "test"},
			wantErrContains: "in",
		},
		{
			name:          "int in list",
			structWithTag: struct{ Field int `validate:"in:5,10,15"` }{Field: 5},
		},
		{
			name:            "int not in list",
			structWithTag:   struct{ Field int `validate:"in:10,15,20"` }{Field: 5},
			wantErrContains: "in",
		},
		{
			name:          "float in list",
			structWithTag: struct{ Field float64 `validate:"in:5.5,10.5,15.5"` }{Field: 5.5},
		},
		{
			name:            "float not in list",
			structWithTag:   struct{ Field float64 `validate:"in:10.5,15.5"` }{Field: 5.5},
			wantErrContains: "in",
		},
		{
			name:          "bool in list",
			structWithTag: struct{ Field bool `validate:"in:true,false"` }{Field: true},
		},
		{
			name:            "bool not in list",
			structWithTag:   struct{ Field bool `validate:"in:false"` }{Field: true},
			wantErrContains: "in",
		},
		{
			name:          "uint in list",
			structWithTag: struct{ Field uint `validate:"in:5,10,15"` }{Field: 5},
		},
		{
			name:            "uint not in list",
			structWithTag:   struct{ Field uint `validate:"in:10,15,20"` }{Field: 5},
			wantErrContains: "in",
		},
		{
			name:          "uint32 in list",
			structWithTag: struct{ Field uint32 `validate:"in:5,10,15"` }{Field: 5},
		},
		{
			name:          "float32 in list",
			structWithTag: struct{ Field float32 `validate:"in:5.5,10.5,15.5"` }{Field: 5.5},
		},
		{
			name:            "float32 not in list",
			structWithTag:   struct{ Field float32 `validate:"in:10.5,15.5"` }{Field: 5.5},
			wantErrContains: "in",
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"in"` }{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("in"))
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

func TestNotInValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "string not in list",
			structWithTag: struct{ Field string `validate:"notin:example,sample"` }{Field: "test"},
		},
		{
			name:            "string in list",
			structWithTag:   struct{ Field string `validate:"notin:test,example,sample"` }{Field: "test"},
			wantErrContains: "notin",
		},
		{
			name:          "int not in list",
			structWithTag: struct{ Field int `validate:"notin:10,15,20"` }{Field: 5},
		},
		{
			name:            "int in list",
			structWithTag:   struct{ Field int `validate:"notin:5,10,15"` }{Field: 5},
			wantErrContains: "notin",
		},
		{
			name:          "float not in list",
			structWithTag: struct{ Field float64 `validate:"notin:10.5,15.5"` }{Field: 5.5},
		},
		{
			name:            "float in list",
			structWithTag:   struct{ Field float64 `validate:"notin:5.5,10.5,15.5"` }{Field: 5.5},
			wantErrContains: "notin",
		},
		{
			name:          "bool not in list",
			structWithTag: struct{ Field bool `validate:"notin:false"` }{Field: true},
		},
		{
			name:            "bool in list",
			structWithTag:   struct{ Field bool `validate:"notin:true,false"` }{Field: true},
			wantErrContains: "notin",
		},
		{
			name:          "uint not in list",
			structWithTag: struct{ Field uint `validate:"notin:10,15,20"` }{Field: 5},
		},
		{
			name:            "uint in list",
			structWithTag:   struct{ Field uint `validate:"notin:5,10,15"` }{Field: 5},
			wantErrContains: "notin",
		},
		{
			name:          "uint32 not in list",
			structWithTag: struct{ Field uint32 `validate:"notin:10,15,20"` }{Field: 5},
		},
		{
			name:            "uint32 in list",
			structWithTag:   struct{ Field uint32 `validate:"notin:5,10,15"` }{Field: 5},
			wantErrContains: "notin",
		},
		{
			name:          "float32 not in list",
			structWithTag: struct{ Field float32 `validate:"notin:10.5,15.5"` }{Field: 5.5},
		},
		{
			name:            "float32 in list",
			structWithTag:   struct{ Field float32 `validate:"notin:5.5,10.5,15.5"` }{Field: 5.5},
			wantErrContains: "notin",
		},
		{
			name:          "missing parameter",
			structWithTag: struct{ Field string `validate:"notin"` }{Field: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("notin"))
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

func TestBooleanValidator(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		structWithTag   any
		wantErrContains string
	}{
		{
			name:          "true value",
			structWithTag: struct{ Field bool `validate:"boolean"` }{Field: true},
		},
		{
			name:          "false value",
			structWithTag: struct{ Field bool `validate:"boolean"` }{Field: false},
		},
		{
			name:            "string value",
			structWithTag:   struct{ Field string `validate:"boolean"` }{Field: "true"},
			wantErrContains: "boolean",
		},
		{
			name:            "int value",
			structWithTag:   struct{ Field int `validate:"boolean"` }{Field: 1},
			wantErrContains: "boolean",
		},
		{
			name:            "float value",
			structWithTag:   struct{ Field float64 `validate:"boolean"` }{Field: 1.0},
			wantErrContains: "boolean",
		},
		{
			name:            "nil pointer",
			structWithTag:   struct{ Field *bool `validate:"boolean"` }{Field: nil},
			wantErrContains: "boolean",
		},
		{
			name:            "pointer to bool",
			structWithTag:   struct{ Field *bool `validate:"boolean"` }{Field: func() *bool { b := true; return &b }()},
			wantErrContains: "boolean",
		},
		{
			name:            "string 'true'",
			structWithTag:   struct{ Field string `validate:"boolean"` }{Field: "true"},
			wantErrContains: "boolean",
		},
		{
			name:            "string 'false'",
			structWithTag:   struct{ Field string `validate:"boolean"` }{Field: "false"},
			wantErrContains: "boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := validator.New(validator.WithValidators("boolean"))
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