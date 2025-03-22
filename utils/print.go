package utils

import (
	"encoding/json"
	"fmt"
)

// FormatJSON returns a pretty-printed JSON string of the given value(s).
// If the value cannot be marshaled to JSON, it returns the value as a string.
// This function is useful for debugging purposes.
//
// Example:
//
//	type User struct {
//	    Name string
//	    Age  int
//	}
//	user := User{Name: "John", Age: 30}
//	FormatJSON(user) // returns formatted JSON representation
func FormatJSON(v ...interface{}) string {
	var result string
	for i, val := range v {
		if i > 0 {
			result += "\n"
		}
		b, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			result += fmt.Sprintf("%+v", val)
		} else {
			result += string(b)
		}
	}
	return result
}

// Deprecated: PrettyPrint is deprecated, use FormatJSON instead.
func PrettyPrint(v ...interface{}) string {
	return FormatJSON(v...)
}
