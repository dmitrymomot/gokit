# i18n Package

A simple, powerful internationalization solution for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/i18n
```

## Overview

The `i18n` package provides internationalization capabilities for Go applications with support for multiple languages, dynamic language switching, and HTTP middleware integration. The package focuses on simplicity, performance, and robust error handling.

## Features

- Support for multiple languages and language detection
- Dynamic language switching based on user preference
- Translation file support (JSON, YAML)
- HTTP middleware for automatic language detection
- Variable substitution in translations
- Comprehensive error handling with specific error types
- Accept-Language header parsing

## Usage

### Basic Translations

```go
import (
	"fmt"
	"github.com/dmitrymomot/gokit/i18n"
)

// Initialize translator
translator, err := i18n.New(i18n.Config{
	DefaultLanguage:    "en",
	SupportedLanguages: []string{"en", "fr", "es"},
	TranslationsPath:   "./translations",
})
if err != nil {
	// Handle initialization error
}

// Get translation in default language
greeting, err := translator.Translate("greeting")
if err != nil {
	// Handle translation error
}
fmt.Println(greeting) // "Hello, world!"

// Get translation in specific language
frGreeting, err := translator.TranslateWithLang("greeting", "fr")
if err != nil {
	// Handle translation error
}
fmt.Println(frGreeting) // "Bonjour, le monde!"
```

### Variable Substitution

```go
// Translation with variables
welcome, err := translator.TranslateWithVars("welcome", map[string]any{
	"name": "John",
})
fmt.Println(welcome) // "Welcome to our application, John!"

// With specific language and variables
frWelcome, err := translator.TranslateWithLangAndVars("welcome", "fr", map[string]any{
	"name": "John",
})
fmt.Println(frWelcome) // "Bienvenue dans notre application, John!"
```

### Translation Files

Translation files should be organized by language code:

```
/translations
  /en
    common.json
    errors.json
  /fr
    common.json
    errors.json
```

Example `en/common.json`:

```json
{
    "greeting": "Hello, world!",
    "welcome": "Welcome to our application, {{name}}!",
    "farewell": "Goodbye, see you soon!"
}
```

### HTTP Integration

```go
import (
	"net/http"
	"github.com/dmitrymomot/gokit/i18n"
)

// Initialize translator
translator, err := i18n.New(i18n.Config{
	DefaultLanguage:    "en",
	SupportedLanguages: []string{"en", "fr", "es"},
	TranslationsPath:   "./translations",
})

// Create a handler that uses translations
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Get the locale from the request context
	locale := i18n.GetLocale(r.Context())
	
	// Use the detected language for translations
	greeting, _ := translator.TranslateWithLang("greeting", locale)
	
	fmt.Fprintf(w, "Greeting: %s\n", greeting)
})

// Apply the i18n middleware to automatically detect language
http.Handle("/", i18n.Middleware(translator, nil)(handler))
```

### Custom Language Detection

```go
// Create a custom language extractor (e.g., from URL query parameter)
customExtractor := func(r *http.Request) string {
	return r.URL.Query().Get("lang")
}

// Apply middleware with custom extractor
http.Handle("/", i18n.Middleware(translator, customExtractor)(handler))
```

### Context Utilities

```go
// Set locale in context
ctx = i18n.SetLocale(context.Background(), "fr")

// Get locale from context (returns default if not set)
locale := i18n.GetLocale(ctx)
```

### Error Handling

```go
greeting, err := translator.TranslateWithLang("greeting", "de")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrLanguageNotSupported):
		// Language not in supported list
	case errors.Is(err, i18n.ErrTranslationNotFound):
		// Translation key doesn't exist
	case errors.Is(err, i18n.ErrInvalidTranslationFormat):
		// Translation file has invalid format
	case errors.Is(err, i18n.ErrFileSystemError):
		// Error accessing or reading files
	}
}
```

## Advanced Configuration

```go
translator, err := i18n.New(i18n.Config{
	DefaultLanguage:    "en",
	SupportedLanguages: []string{"en", "fr", "es"},
	TranslationsPath:   "./translations",
	FileFormat:         "json",          // or "yaml"
	EnableLogging:      true,            // for debugging
	LogLevel:           i18n.LogLevelInfo,
	FallbackToDefault:  true,            // use default if translation missing
})
