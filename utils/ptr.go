package utils

// Ptr is a generic utility function that returns a pointer to the provided value.
// This is useful for creating pointers to values in one-liners, especially when
// working with structs or API requests that require pointers.
func Ptr[T any](v T) *T {
	return &v
}
