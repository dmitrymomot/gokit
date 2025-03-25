package middlewares

// contextKey is a private type for context keys to avoid collisions.
type contextKey struct{ name string }

// String returns the name of the context key.
func (c contextKey) String() string { return c.name }

// newContextKey creates a new context key with the given name.
func newContextKey(name string) *contextKey {
	return &contextKey{name: name}
}
