package i18n

import "context"

type Parser interface {
	Parse(ctx context.Context, content string) (map[string]map[string]any, error)
	// SupportsFileExtension checks if the parser supports a given file extension
	// The extension may or may not include a leading dot (e.g. both "json" and ".json" are valid)
	SupportsFileExtension(ext string) bool
}
