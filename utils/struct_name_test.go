package utils_test

import (
	"testing"

	"github.com/dmitrymomot/gokit/utils"
)

type Person struct {
	Name string
	Age  int
}

func TestQualifiedStructName(t *testing.T) {
	// Test non-pointer struct type
	p := Person{}
	expected1 := "utils_test.Person"
	actual1 := utils.QualifiedStructName(p)
	if actual1 != expected1 {
		t.Errorf("Expected %q but got %q", expected1, actual1)
	}

	// Test pointer struct type
	ptr := &Person{}
	expected2 := "utils_test.Person"
	actual2 := utils.QualifiedStructName(ptr)
	if actual2 != expected2 {
		t.Errorf("Expected %q but got %q", expected2, actual2)
	}
}

// Test deprecated function for backward compatibility
func TestFullyQualifiedStructName(t *testing.T) {
	p := Person{}
	expected := "utils_test.Person"
	actual := utils.FullyQualifiedStructName(p)
	if actual != expected {
		t.Errorf("Expected %q but got %q", expected, actual)
	}
}

func TestStructName(t *testing.T) {
	// Test non-pointer struct type
	p := Person{}
	expected1 := "Person"
	actual1 := utils.StructName(p)
	if actual1 != expected1 {
		t.Errorf("Expected %q but got %q", expected1, actual1)
	}

	// Test pointer struct type
	ptr := &Person{}
	expected2 := "Person"
	actual2 := utils.StructName(ptr)
	if actual2 != expected2 {
		t.Errorf("Expected %q but got %q", expected2, actual2)
	}
}

type NamedPerson struct {
	FirstName string
	Age       int
}

func (p NamedPerson) Name() string {
	return "NamedPerson"
}

func TestGetNameFromStruct(t *testing.T) {
	// Define a fallback function
	fallback := func(v any) string {
		return "fallback"
	}

	// Test with struct implementing NamedEntity interface
	p := NamedPerson{}
	expected1 := "NamedPerson"
	actual1 := utils.GetNameFromStruct(p, fallback)
	if actual1 != expected1 {
		t.Errorf("Expected %q but got %q", expected1, actual1)
	}

	// Test with pointer to struct implementing NamedEntity interface
	ptr := &NamedPerson{}
	expected2 := "NamedPerson"
	actual2 := utils.GetNameFromStruct(ptr, fallback)
	if actual2 != expected2 {
		t.Errorf("Expected %q but got %q", expected2, actual2)
	}

	// Test with type not implementing NamedEntity interface
	i := 42
	expected3 := "fallback"
	actual3 := utils.GetNameFromStruct(i, fallback)
	if actual3 != expected3 {
		t.Errorf("Expected %q but got %q", expected3, actual3)
	}
}

// Test deprecated function for backward compatibility
func TestNamedStruct(t *testing.T) {
	// Define a fallback function
	fallback := func(v any) string {
		return "fallback"
	}

	// Test with struct implementing NamedEntity interface
	p := NamedPerson{}
	expected1 := "NamedPerson"
	actual1 := utils.GetNameFromStruct(p, fallback)
	if actual1 != expected1 {
		t.Errorf("Expected %q but got %q", expected1, actual1)
	}

	// Test with type not implementing NamedEntity interface
	i := 42
	expected2 := "fallback"
	actual2 := utils.GetNameFromStruct(i, fallback)
	if actual2 != expected2 {
		t.Errorf("Expected %q but got %q", expected2, actual2)
	}
}
