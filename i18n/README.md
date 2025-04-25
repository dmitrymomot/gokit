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
- Pluralization support with count-based templates
- Duration formatting in localized strings
- Context-based translation methods
- Comprehensive error handling with specific error types
- Accept-Language header parsing
- JSON export for client-side translations

## Usage

### Basic Translations

```go
import (
	"context"
	"fmt"
	"github.com/dmitrymomot/gokit/i18n"
)

// Initialize translator
translator, err := i18n.NewTranslator(context.Background(), adapter, 
	i18n.WithDefaultLanguage("en"),
	i18n.WithFallbackToKey(true),
)
if err != nil {
	// Handle initialization error
}

// Get translation in default language
greeting, err := translator.T("en", "greeting")
if err != nil {
	// Handle translation error
}
fmt.Println(greeting) // "Hello, world!"

// Get translation in specific language
frGreeting := translator.T("fr", "greeting")
fmt.Println(frGreeting) // "Bonjour, le monde!"
```

### Variable Substitution

```go
// Translation with variables
welcome := translator.T("en", "welcome", "name", "John")
fmt.Println(welcome) // "Welcome to our application, John!"

// With specific language and variables
frWelcome := translator.T("fr", "welcome", "name", "John")
fmt.Println(frWelcome) // "Bienvenue dans notre application, John!"
```

### Pluralization

```go
// Pluralized translation (key has different forms based on count)
items := translator.N("en", "items", 1, "count", "1")
fmt.Println(items) // "1 item"

items = translator.N("en", "items", 5, "count", "5")
fmt.Println(items) // "5 items"
```

### Translation with Default Fallback

```go
// If translation is missing, use the provided default value
message := translator.Td("en", "admin.welcome", "Welcome, Admin!", "name", "John")
fmt.Println(message) // "Welcome, Admin!" if key doesn't exist, with variables substituted
```

### Duration Formatting

```go
// Convert durations to human-readable localized strings
duration := translator.Duration("en", 90 * time.Minute)
fmt.Println(duration) // "1 hour 30 minutes" (depending on translation files)

// Different languages format durations differently
frDuration := translator.Duration("fr", 2 * 24 * time.Hour)
fmt.Println(frDuration) // "2 jours" (depending on translation files)
```

### Context-Based Translations

```go
// Set language in context
ctx = i18n.SetLocale(context.Background(), "fr")

// Use context-based translation (uses language from context)
message := translator.Tc(ctx, "greeting")
fmt.Println(message) // Uses French translation

// Pluralized context-based translation
items := translator.Nc(ctx, "items", 3, "count", "3")
fmt.Println(items) // Uses French pluralization rules with count 3
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
    "items": {
        "one": "{{count}} item",
        "other": "{{count}} items"
    },
    "datetime": {
        "hours": {
            "one": "{{count}} hour",
            "other": "{{count}} hours"
        }
    }
}
```

### HTTP Integration

```go
import (
	"net/http"
	"github.com/dmitrymomot/gokit/i18n"
)

// Initialize translator with adapter
translator, _ := i18n.NewTranslator(context.Background(), adapter)

// Create a handler that uses translations
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Get the locale from the request context
	locale := i18n.GetLocale(r.Context())
	
	// Use the detected language for translations
	greeting := translator.T(locale, "greeting")
	
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

### Client-Side Translations

```go
// Export translations for a language as JSON string
// Useful for sending translations to browser-based applications
jsonData, err := translator.ExportJSON("en")
if err != nil {
	// Handle error
}

// Send to client
w.Header().Set("Content-Type", "application/json")
w.Write([]byte(jsonData))
```

## API Reference

### Core Types

```go
// Main translator implementation
func NewTranslator(ctx context.Context, adapter TranslationAdapter, options ...Option) (*Translator, error)

// Configuration options
type Option func(*Translator)
```

### Translation Methods

```go
// Basic translation methods
func (t *Translator) T(lang, key string, args ...string) string
func (t *Translator) N(lang, key string, n int, args ...string) string
func (t *Translator) Td(lang, key, defaultValue string, args ...string) string
func (t *Translator) Duration(lang string, d time.Duration) string

// Context-based translation methods
func (t *Translator) Tc(ctx context.Context, key string, args ...string) string
func (t *Translator) Nc(ctx context.Context, key string, n int, args ...string) string

// Export and utility methods
func (t *Translator) ExportJSON(lang string) (string, error)
func (t *Translator) HasTranslation(lang, key string) bool
func (t *Translator) SupportedLanguages() []string
func (t *Translator) Lang(header string, defaultLocale ...string) string
```

### Middleware and Context Utilities

```go
// HTTP middleware
func Middleware(t translator, extr langExtractor) func(http.Handler) http.Handler

// Context utilities
func SetLocale(ctx context.Context, locale string) context.Context
func GetLocale(ctx context.Context) string
```

### Configuration Options

```go
// Translator configuration
func WithDefaultLanguage(lang string) Option
func WithFallbackToKey(fallback bool) Option
func WithLogger(logger *slog.Logger) Option
func WithMissingTranslationsLogging(log bool) Option
func WithNoLogging() Option
```

### Error Types

```go
// Error types
var ErrLanguageNotSupported
var ErrTranslationNotFound
var ErrInvalidTranslationFormat
var ErrFileSystemError
```

## Error Handling

```go
// Handle specific error types
translation, err := translator.TranslateWithLang("greeting", "de")
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

## Best Practices

1. **Organization**:
   - Organize translations by language and category
   - Keep translation keys consistent across languages
   - Use hierarchical keys for related translations

2. **Error Handling**:
   - Always check for errors when initializing
   - Use appropriate fallbacks for missing translations
   - Consider enabling missing translation logging during development

3. **Performance**:
   - Use the context-based methods when possible to avoid redundant language detection
   - Consider caching frequently used translations
   - Limit the use of complex variable substitutions in hot paths

4. **Maintenance**:
   - Keep translation files in a clearly defined structure
   - Document the supported languages and translation keys
   - Consider using a translation management system for large projects
   - Export translations to JSON for client-side applications when needed
