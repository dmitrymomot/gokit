# i18n

A lightweight, thread-safe internationalization and localization package for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/i18n
```

## Features

- Load translations from YAML files or directories
- Support for embedded translations via `http.FileSystem`
- Thread-safe operations for concurrent applications
- Text translation with variable substitution
- Pluralization support for different languages
- Accept-Language header parsing for automatic language selection
- Fallback mechanisms for missing translations
- Custom error types for better error handling
- Validation of translation files
- Configurable logging for diagnostics

## Usage

### Basic Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/i18n"
)

func main() {
	// Load translations from a YAML file
	err := i18n.LoadTranslations("translations.yaml")
	if err != nil {
		log.Fatalf("Failed to load translations: %v", err)
	}

	// Simple translation
	fmt.Println(i18n.T("en", "welcome", "name", "John"))
	// Output: "Welcome, John!"

	// Translation with pluralization
	fmt.Println(i18n.N("en", "items", 1, "count", "1"))
	// Output: "1 item"
	fmt.Println(i18n.N("en", "items", 5, "count", "5"))
	// Output: "5 items"
}
```

### Translation File Format

Translation files use YAML format. The structure is organized by language code at the top level:

```yaml
en:
  welcome: "Welcome, %{name}!"
  items:
    one: "%{count} item"
    other: "%{count} items"
  nested:
    greeting: "Hello from nested key!"

fr:
  welcome: "Bienvenue, %{name}!"
  items:
    one: "%{count} élément"
    other: "%{count} éléments"
  nested:
    greeting: "Bonjour depuis une clé imbriquée!"
```

### Loading Translations

You can load translations in several ways:

```go
// From a single file
i18n.LoadTranslations("translations.yaml")

// From a directory (recursively loads all .yaml and .yml files)
i18n.LoadTranslationsDir("./translations")

// From an embedded file system (e.g., with Go 1.16+ embed)
//go:embed translations
var translationsFS embed.FS
httpFS := http.FS(translationsFS)
i18n.LoadTranslationsFS(httpFS, ".")
```

### Translation Functions

```go
// Simple translation with variable substitution
i18n.T("en", "welcome", "name", "John")

// Nested keys using dot notation
i18n.T("en", "nested.greeting")

// Pluralization based on count
i18n.N("en", "items", 1, "count", "1")  // Uses "one" form
i18n.N("en", "items", 5, "count", "5")  // Uses "other" form

// Language selection from Accept-Language header
acceptLang := "fr-CA,fr;q=0.9,en;q=0.8"
lang := i18n.BestLangFromAcceptLanguage(acceptLang)
i18n.T(lang, "welcome", "name", "User")
```

### Pluralization Rules

The package follows a simple pluralization system:

- `zero`: Used when the count is 0 (optional)
- `one`: Used when the count is 1
- `other`: Used for all other cases

More complex pluralization rules may be added in future versions.

## Thread Safety

All operations in this package are thread-safe and can be used in concurrent applications.

## Error Handling

The package now provides custom error types for better error diagnostics:

```go
// Handling file loading errors
err := i18n.LoadTranslations("nonexistent.yaml")
if err != nil {
    // Error will be of type ErrFileSystemError
    log.Fatalf("Failed to load translations: %v", err)
}

// Invalid YAML format
err = i18n.LoadTranslations("invalid.yaml")
if err != nil {
    // Error will be of type ErrInvalidTranslationFormat
    log.Fatalf("Invalid translation format: %v", err)
}
```

By default, when translations are missing, the package falls back to the provided key:

```go
// If "missing_key" doesn't exist for "en"
result := i18n.T("en", "missing_key")
// result will be "missing_key"
```

You can customize this behavior with the following configuration options:

```go
// Disable fallback to key (will return empty string instead)
i18n.FallbackToKey = false

// Enable logging of missing translations
i18n.LogMissingTranslations = true

// Configure custom logger (uses slog)
i18n.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
```

## Validation

The package now validates translation files during loading:

1. Checks for empty language codes
2. Ensures all language entries are properly formatted
3. Warns about empty translation files

This helps catch configuration errors early and provides better error messages.