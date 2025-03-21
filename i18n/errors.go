package i18n

import "fmt"

// Error types for the i18n package
var (
	// ErrLanguageNotSupported is returned when a requested language doesn't have translations.
	ErrLanguageNotSupported = func(lang string) error {
		return fmt.Errorf("language not supported: %s", lang)
	}

	// ErrTranslationNotFound is returned when a translation key doesn't exist.
	ErrTranslationNotFound = func(lang, key string) error {
		return fmt.Errorf("translation not found for %s: %s", lang, key)
	}

	// ErrInvalidTranslationFormat is returned when a translation file has an invalid format.
	ErrInvalidTranslationFormat = func(file string, err error) error {
		return fmt.Errorf("invalid translation format in %s: %w", file, err)
	}

	// ErrFileSystemError is returned when there's an error accessing the file system.
	ErrFileSystemError = func(err error) error {
		return fmt.Errorf("file system error: %w", err)
	}
)
