package validator_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/validator"
	"github.com/stretchr/testify/require"
)

// TestCreditCardValidator tests the creditCardValidator function.
func TestCreditCardValidator(t *testing.T) {
	type args struct {
		Value string
	}
	tests := []struct {
		name            string
		args            args
		structWithTag   interface{}
		wantErr         bool
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid credit card visa",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "4242424242424242"},
		},
		{
			name: "valid credit card mastercard with spaces",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "5454 5454 5454 5454"},
		},
		{
			name: "valid credit card amex",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "371449635398431"},
		},
		// Invalid cases
		{
			name: "invalid credit card - luhn check fail",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "49927398717"},
			wantErr:         true,
			wantErrContains: "validation.creditcard",
		},
		{
			name: "invalid credit card - too short",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "123456789012"},
			wantErr:         true,
			wantErrContains: "validation.creditcard",
		},
		{
			name: "invalid credit card - too long",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "12345678901234567890"},
			wantErr:         true,
			wantErrContains: "validation.creditcard",
		},
		{
			name: "invalid credit card - contains non-numeric",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: "4992739871A"},
			wantErr:         true,
			wantErrContains: "validation.creditcard",
		},
		{
			name: "empty string",
			structWithTag: struct {
				CreditCard string `validate:"creditcard"`
			}{CreditCard: ""},
			wantErr:         true,
			wantErrContains: "validation.creditcard",
		},
		// Type mismatch
		{
			name: "type mismatch int",
			structWithTag: struct {
				CreditCard int `validate:"creditcard"`
			}{CreditCard: 123},
			// This validator returns nil if the type is not string
		},
	}

	v, err := validator.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.structWithTag)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPasswordValidator tests the passwordValidator function.
func TestPasswordValidator(t *testing.T) {
	type args struct {
		Value string
	}
	tests := []struct {
		name            string
		args            args
		structWithTag   interface{}
		wantErr         bool
		wantErrContains string
	}{
		// Valid cases
		{
			name: "valid password",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "Password123!"},
		},
		{
			name: "valid password with spaces (trimmed)",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "  Password123!  "},
		},
		// Invalid cases - length
		{
			name: "invalid password - too short",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "Pass1!"},
			wantErr:         true,
			wantErrContains: "validation.password",
		},
		// Invalid cases - missing components
		{
			name: "invalid password - no uppercase",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "password123!"},
			wantErr:         true,
			wantErrContains: "validation.password",
		},
		{
			name: "invalid password - no lowercase",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "PASSWORD123!"},
			wantErr:         true,
			wantErrContains: "validation.password",
		},
		{
			name: "invalid password - no number",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "Password!"},
			wantErr:         true,
			wantErrContains: "validation.password",
		},
		{
			name: "invalid password - no special character",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: "Password123"},
			wantErr:         true,
			wantErrContains: "validation.password",
		},
		{
			name: "empty string",
			structWithTag: struct {
				Password string `validate:"password"`
			}{Password: ""},
			wantErr:         true,
			wantErrContains: "validation.password", // Empty string does not meet length or complexity
		},
		// Type mismatch
		{
			name: "type mismatch int",
			structWithTag: struct {
				Password int `validate:"password"`
			}{Password: 12345678},
			// passwordValidator expects a string; non-string types result in behavior dependent on reflect.Value.String()
			// For an int, this would likely be the string representation of the int, which would then fail validation.
			wantErr:         true,
			wantErrContains: "validation.password",
		},
	}

	v, err := validator.New()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.structWithTag)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
