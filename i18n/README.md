# i18n Package

A simple, powerful internationalization solution for Go applications.

## Installation

```bash
go get github.com/dmitrymomot/gokit/i18n
```

## Overview

The `i18n` package provides internationalization capabilities for Go applications with support for multiple languages, dynamic language switching, and HTTP middleware integration. The package focuses on simplicity, performance, and robust error handling. It is thread-safe and designed for concurrent use in production environments.

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
- Thread-safe implementation for concurrent usage

## Usage

### Basic Translations

```go
import (
	"context"
	"fmt"
	"github.com/dmitrymomot/gokit/i18n"
)

// Initialize translator with a filesystem adapter
adapter, err := i18n.NewFileSystemAdapter("./translations")
if err != nil {
	// Handle adapter creation error
	panic(fmt.Sprintf("Failed to create adapter: %v", err))
}

// Initialize translator
translator, err := i18n.NewTranslator(context.Background(), adapter, 
	i18n.WithDefaultLanguage("en"),
	i18n.WithFallbackToKey(true),
)
if err != nil {
	// Handle initialization error
	panic(fmt.Sprintf("Failed to initialize translator: %v", err))
}

// Get translation in default language
greeting, err := translator.T("en", "greeting")
if err != nil {
	// Handle translation error
	fmt.Println("Error:", err)
} else {
	fmt.Println(greeting) 
	// Output: "Hello, world!"
}

// Get translation in specific language
frGreeting, err := translator.T("fr", "greeting")
if err != nil {
	// Handle translation error
	fmt.Println("Error:", err)
} else {
	fmt.Println(frGreeting) 
	// Output: "Bonjour, le monde!"
}
```

### Variable Substitution

```go
// Translation with variables
welcome, err := translator.T("en", "welcome", "name", "John")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(welcome) 
	// Output: "Welcome to our application, John!"
}

// With specific language and variables
frWelcome, err := translator.T("fr", "welcome", "name", "John")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(frWelcome) 
	// Output: "Bienvenue dans notre application, John!"
}
```

### Pluralization

```go
// Pluralized translation (key has different forms based on count)
items, err := translator.N("en", "items", 1, "count", "1")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(items) 
	// Output: "1 item"
}

multiItems, err := translator.N("en", "items", 5, "count", "5")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(multiItems) 
	// Output: "5 items"
}
```

### Translation with Default Fallback

```go
// If translation is missing, use the provided default value
message, err := translator.Td("en", "admin.welcome", "Welcome, Admin!", "name", "John")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(message) 
	// Output: "Welcome, Admin!" if key doesn't exist, with variables substituted
}
```

### Duration Formatting

```go
// Convert durations to human-readable localized strings
duration, err := translator.Duration("en", 90 * time.Minute)
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(duration) 
	// Output: "1 hour 30 minutes" (depending on translation files)
}

// Different languages format durations differently
frDuration, err := translator.Duration("fr", 2 * 24 * time.Hour)
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(frDuration) 
	// Output: "2 jours" (depending on translation files)
}
```

### Context-Based Translations

```go
// Set language in context
ctx := i18n.SetLocale(context.Background(), "fr")

// Use context-based translation (uses language from context)
message, err := translator.Tc(ctx, "greeting")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(message) 
	// Output: Uses French translation - "Bonjour, le monde!"
}

// Pluralized context-based translation
items, err := translator.Nc(ctx, "items", 3, "count", "3")
if err != nil {
	fmt.Println("Error:", err)
} else {
	fmt.Println(items) 
	// Output: Uses French pluralization rules with count 3
}
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
	greeting, err := translator.T(locale, "greeting")
	if err != nil {
		http.Error(w, "Translation error", http.StatusInternalServerError)
		return
	}
	
	fmt.Fprintf(w, "Greeting: %s\n", greeting)
	// If request has Accept-Language: fr
	// Output: Greeting: Bonjour, le monde!
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
// Now visiting /?lang=fr will use French translations
```

### Client-Side Translations

```go
// Export translations for a language as JSON string
// Useful for sending translations to browser-based applications
jsonData, err := translator.ExportJSON("en")
if err != nil {
	// Handle error
	http.Error(w, "Export error", http.StatusInternalServerError)
	return
}

// Send to client
w.Header().Set("Content-Type", "application/json")
w.Write([]byte(jsonData))
// Output: JSON containing all English translations
```

### Error Handling

```go
import (
	"context"
	"errors"
	"fmt"
	
	"github.com/dmitrymomot/gokit/i18n"
)

// Example 1: Handling unsupported language
translation, err := translator.T("xyz", "greeting")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrLanguageNotSupported):
		fmt.Printf("Language 'xyz' is not supported: %v\n", err)
		// Use default language as fallback
		translation, _ = translator.T("en", "greeting")
	default:
		fmt.Printf("Unexpected error: %v\n", err)
	}
}

// Example 2: Handling missing translation
translation, err = translator.T("en", "nonexistent.key")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrTranslationNotFound):
		fmt.Printf("Translation key 'nonexistent.key' not found: %v\n", err)
		// Use a default message
		translation = "Default message"
	default:
		fmt.Printf("Unexpected error: %v\n", err)
	}
}

// Example 3: Handling invalid translation format
// Imagine a corrupted translation file
translation, err = translator.T("en", "corrupt.key")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrInvalidTranslationFormat):
		fmt.Printf("Invalid translation format: %v\n", err)
		// Report the issue and use a safe fallback
		translation = "Error in translation system"
	default:
		fmt.Printf("Unexpected error: %v\n", err)
	}
}

// Example 4: Handling file system errors
// This might happen if translation files become inaccessible
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrFileSystemError):
		fmt.Printf("File system error accessing translations: %v\n", err)
		// Use in-memory fallback translations for critical messages
		translation = inMemoryFallbackTranslations["greeting"]
	default:
		fmt.Printf("Unexpected error: %v\n", err)
	}
}

// Example 5: Comprehensive error handling with fallbacks
func getTranslation(t *i18n.Translator, lang, key string) string {
	translation, err := t.T(lang, key)
	if err != nil {
		switch {
		case errors.Is(err, i18n.ErrLanguageNotSupported):
			// Try default language
			translation, err = t.T("en", key)
			if err != nil {
				return key // Use key as last resort
			}
			return translation
		case errors.Is(err, i18n.ErrTranslationNotFound):
			return key // Use key as fallback
		case errors.Is(err, i18n.ErrInvalidTranslationFormat):
			fmt.Printf("Invalid translation format for %s.%s\n", lang, key)
			return key
		case errors.Is(err, i18n.ErrFileSystemError):
			fmt.Printf("File system error: %v\n", err)
			return key
		default:
			fmt.Printf("Unknown error: %v\n", err)
			return key
		}
	}
	return translation
}
```

## Best Practices

1. **Organization**:
   - Organize translations by language and category
   - Keep translation keys consistent across languages
   - Use hierarchical keys for related translations (e.g., `user.greeting`, `user.farewell`)

2. **Error Handling**:
   - Always check for errors when initializing
   - Implement appropriate fallbacks for missing translations
   - Enable missing translation logging during development
   - Handle specific error types with appropriate responses

3. **Performance**:
   - Use the context-based methods when possible to avoid redundant language detection
   - Consider caching frequently used translations in memory
   - Limit the use of complex variable substitutions in hot paths
   - Use the appropriate translation method for your needs (T for simple, N for plurals)

4. **Maintenance**:
   - Keep translation files in a clearly defined structure
   - Document the supported languages and translation keys
   - Consider using a translation management system for large projects
   - Export translations to JSON for client-side applications when needed

## API Reference

### Types

```go
type Translator struct {
    // Contains unexported fields
}
```
Main translator implementation.

```go
type Option func(*Translator)
```
Configuration option function type for customizing the translator.

```go
type TranslationAdapter interface {
    GetTranslation(lang, key string) (string, error)
    SupportedLanguages() []string
    // Other methods
}
```
Interface for translation storage adapters.

### Functions

```go
func NewTranslator(ctx context.Context, adapter TranslationAdapter, options ...Option) (*Translator, error)
```
Creates a new translator with the specified adapter and options.

```go
func NewFileSystemAdapter(path string) (TranslationAdapter, error)
```
Creates a new filesystem-based translation adapter.

```go
func Middleware(t translator, extr langExtractor) func(http.Handler) http.Handler
```
Creates HTTP middleware for automatic language detection.

```go
func SetLocale(ctx context.Context, locale string) context.Context
```
Sets the locale in the context.

```go
func GetLocale(ctx context.Context) string
```
Gets the locale from the context.

### Configuration Options

```go
func WithDefaultLanguage(lang string) Option
```
Sets the default language for the translator.

```go
func WithFallbackToKey(fallback bool) Option
```
Configures whether to fall back to the key when translation is missing.

```go
func WithLogger(logger *slog.Logger) Option
```
Sets a custom logger for the translator.

```go
func WithMissingTranslationsLogging(log bool) Option
```
Enables or disables logging of missing translations.

```go
func WithNoLogging() Option
```
Disables all logging for the translator.

### Translation Methods

```go
func (t *Translator) T(lang, key string, args ...string) (string, error)
```
Basic translation method with variable substitution.

```go
func (t *Translator) N(lang, key string, n int, args ...string) (string, error)
```
Pluralized translation based on count.

```go
func (t *Translator) Td(lang, key, defaultValue string, args ...string) (string, error)
```
Translation with a default fallback value.

```go
func (t *Translator) Duration(lang string, d time.Duration) (string, error)
```
Converts duration to a localized string.

```go
func (t *Translator) Tc(ctx context.Context, key string, args ...string) (string, error)
```
Context-based translation using language from context.

```go
func (t *Translator) Nc(ctx context.Context, key string, n int, args ...string) (string, error)
```
Context-based pluralized translation.

```go
func (t *Translator) ExportJSON(lang string) (string, error)
```
Exports all translations for a language as JSON.

```go
func (t *Translator) HasTranslation(lang, key string) bool
```
Checks if a translation exists.

```go
func (t *Translator) SupportedLanguages() []string
```
Returns all supported languages.

```go
func (t *Translator) Lang(header string, defaultLocale ...string) string
```
Parses Accept-Language header to determine language.

### Error Types

```go
var ErrLanguageNotSupported = errors.New("language not supported")
var ErrTranslationNotFound = errors.New("translation not found")
var ErrInvalidTranslationFormat = errors.New("invalid translation format")
var ErrFileSystemError = errors.New("file system error")
```
