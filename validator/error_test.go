package validator_test

import (
	"reflect"
	"testing"

	"github.com/dmitrymomot/gokit/validator"
)

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected validator.ValidationErrors
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: validator.ValidationErrors{},
		},
		{
			name:     "single pair",
			args:     []string{"field1", "error1"},
			expected: validator.ValidationErrors{"field1": []string{"error1"}},
		},
		{
			name: "multiple pairs",
			args: []string{"field1", "error1", "field2", "error2"},
			expected: validator.ValidationErrors{
				"field1": []string{"error1"},
				"field2": []string{"error2"},
			},
		},
		{
			name: "odd number of arguments",
			args: []string{"field1", "error1", "field2"},
			expected: validator.ValidationErrors{
				"field1": []string{"error1"},
			},
		},
		{
			name:     "empty strings",
			args:     []string{"", ""},
			expected: validator.ValidationErrors{"": []string{""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.NewValidationError(tt.args...)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewValidationError() = %v, want %v", got, tt.expected)
			}
		})
	}
}
