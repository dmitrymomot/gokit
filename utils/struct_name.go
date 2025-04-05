package utils

import (
	"fmt"
	"strings"
)

// QualifiedStructName returns the fully qualified struct name in format [package].[type name].
// It ignores whether the value is a pointer or not.
//
// Example:
//
//	type Example struct{}
//	QualifiedStructName(Example{}) // returns "package_name.Example"
//	QualifiedStructName(&Example{}) // returns "package_name.Example"
func QualifiedStructName(v any) string {
	s := fmt.Sprintf("%T", v)
	s = strings.TrimLeft(s, "*")

	return s
}

// StructName returns only the struct name without the package prefix.
// It ignores whether the value is a pointer or not.
//
// Example:
//
//	type Example struct{}
//	StructName(Example{}) // returns "Example"
func StructName(v any) string {
	segments := strings.Split(fmt.Sprintf("%T", v), ".")

	return segments[len(segments)-1]
}

// NamedEntity represents a type that can return its own name.
type NamedEntity interface {
	Name() string
}

// GetNameFromStruct returns the name from a struct that implements the NamedEntity interface,
// or falls back to the provided function if the struct doesn't implement the interface.
//
// Example:
//
//	type Person struct{}
//	func (p Person) Name() string { return "Person" }
//
//	// Using with a struct that implements NamedEntity
//	GetNameFromStruct(Person{}, StructName) // returns "Person" from the Name() method
//
//	// Using with a struct that doesn't implement NamedEntity
//	GetNameFromStruct(42, StructName) // returns result from fallback function
func GetNameFromStruct(v any, fallback func(v any) string) string {
	if v, ok := v.(NamedEntity); ok {
		return v.Name()
	}

	return fallback(v)
}

// Deprecated: FullyQualifiedStructName is deprecated, use QualifiedStructName instead.
func FullyQualifiedStructName(v any) string {
	return QualifiedStructName(v)
}
