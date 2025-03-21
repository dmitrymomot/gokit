package i18nx

import "context"

// TranslationAdapter interface defines how translations are loaded
type TranslationAdapter interface {
	// Load returns the translations map
	Load(ctx context.Context) (map[string]map[string]any, error)
}

// MapAdapter is a simple adapter that uses an in-memory map as the translation source
type MapAdapter struct {
	Data map[string]map[string]any
}

// Load implements the TranslationAdapter interface
func (a *MapAdapter) Load(_ context.Context) (map[string]map[string]any, error) {
	if a.Data == nil {
		return make(map[string]map[string]any), nil
	}
	return a.Data, nil
}
