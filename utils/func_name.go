package utils

import (
	"reflect"
	"runtime"
	"strings"
)

// QualifiedFuncName returns a function's fully qualified name in format [package].[func name].
// It omits the function signature and returns only the package and function name.
//
// Example:
//
//	func example() {}
//	QualifiedFuncName(example) // returns "package_name.example"
//
// If the value is not a function or is nil, an empty string is returned.
func QualifiedFuncName(v any) string {
	if v == nil {
		return ""
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Func {
		return ""
	}

	fullName := runtime.FuncForPC(val.Pointer()).Name()
	// Strip the package path to match the expected output
	if lastSlash := strings.LastIndex(fullName, "/"); lastSlash != -1 {
		fullName = fullName[lastSlash+1:]
	}
	return fullName
}

// Deprecated: FullyQualifiedFuncName is deprecated, use QualifiedFuncName instead.
func FullyQualifiedFuncName(v any) string {
	return QualifiedFuncName(v)
}
