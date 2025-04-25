package utils_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/utils"
)

func TestQualifiedFuncName(t *testing.T) {
	testFunc := func() {}

	tests := []struct {
		name string
		v    any
		want string
	}{
		{
			name: "BasicTestFunction",
			v:    testFunc,
			want: "utils_test.TestQualifiedFuncName.func1",
		},
		{
			name: "QualifiedFuncNameItself",
			v:    utils.QualifiedFuncName,
			want: "utils.QualifiedFuncName",
		},
		{
			name: "NonFunctionVariable",
			v:    5,
			want: "",
		},
		{
			name: "Nil",
			v:    nil,
			want: "",
		},
	}

	for _, tt := range tests {
		actual := utils.QualifiedFuncName(tt.v)
		if actual != tt.want {
			t.Errorf("%s: got %s, want %s", tt.name, actual, tt.want)
		}
	}
}

// Test to ensure the deprecated function still works correctly
func TestFullyQualifiedFuncName(t *testing.T) {
	result := utils.FullyQualifiedFuncName(utils.QualifiedFuncName)
	expected := "utils.QualifiedFuncName"

	if result != expected {
		t.Errorf("FullyQualifiedFuncName: got %s, want %s", result, expected)
	}
}
