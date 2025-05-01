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
	return fmt.Errorf("failed to create adapter: %w", err)
}

// Initialize translator
translator, err := i18n.NewTranslator(context.Background(), adapter, 
	i18n.WithDefaultLanguage("en"),
	i18n.WithFallbackToKey(true),
)
if err != nil {
	// Handle initialization error
	return fmt.Errorf("failed to initialize translator: %w", err)
}

// Get translation in default language
greeting := translator.T("en", "greeting")
// greeting = "Hello, world!"

// Get translation in specific language
frGreeting := translator.T("fr", "greeting")
// frGreeting = "Bonjour, le monde!"

// If you need to check if a language is supported first
if !slices.Contains(translator.SupportedLanguages(), "xyz") {
	// Handle unsupported language
	fmt.Println("Language 'xyz' is not supported")
}
```

### Variable Substitution

```go
// Translation with variables
welcome := translator.T("en", "welcome", "name", "John")
// welcome = "Welcome to our application, John!"
```

### Pluralization

```go
// Pluralized translation (key has different forms based on count)
items := translator.N("en", "items", 1, "count", "1")
// items = "1 item"

multiItems := translator.N("en", "items", 5, "count", "5")
// multiItems = "5 items"
```

### HTTP Middleware

```go
import (
	"net/http"
	"github.com/dmitrymomot/gokit/i18n"
)

// Create a handler that uses translations
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Get the translator from the request context
	greeting := translator.Tc(r.Context(), "greeting")
	
	// Response will be in the language determined from the request
	fmt.Fprintf(w, "Greeting: %s\n", greeting)
})

// Apply the i18n middleware to automatically detect language
http.Handle("/", i18n.Middleware(translator, nil)(handler))
```

### Custom Language Detection

```go
// Create a custom language extractor (e.g., from URL query parameter)
extractor := func(r *http.Request) string {
	return r.URL.Query().Get("lang")
}

// Use the custom extractor with the middleware
http.Handle("/custom", i18n.Middleware(translator, extractor)(handler))
```

### Error Handling for Initialization and Export

```go
// Error handling for initialization
adapter, err := i18n.NewFileSystemAdapter("./translations")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrFileSystemError):
		// Handle file system error
		fmt.Printf("Error accessing translation files: %v\n", err)
	default:
		// Handle other unexpected errors
		fmt.Printf("Unexpected error: %v\n", err)
	}
}

// Error handling for JSON export
jsonData, err := translator.ExportJSON("xyz")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrLanguageNotSupported):
		// Handle unsupported language
		fmt.Printf("Language 'xyz' is not supported: %v\n", err)
	default:
		// Handle other unexpected errors
		fmt.Printf("Unexpected error: %v\n", err)
	}
}
```

## Best Practices

1. **Translation Management**:
   - Organize translations in a logical directory structure
   - Use JSON or YAML for translation files
   - Use dot notation for organizing nested translations
   - Keep translation keys consistent across languages

2. **Error Handling**:
   - Always check for errors when initializing the translator
   - Implement appropriate fallbacks when translations are missing
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

### Translation Methods

```go
func (t *Translator) T(lang, key string, args ...string) string
```
Basic translation method with variable substitution.

```go
func (t *Translator) N(lang, key string, n int, args ...string) string
```
Pluralized translation based on count.

```go
func (t *Translator) Td(lang, key, defaultValue string, args ...string) string
```
Translation with a default fallback value.

```go
func (t *Translator) Duration(lang string, d time.Duration) string
```
Converts duration to a localized string.

```go
func (t *Translator) Tc(ctx context.Context, key string, args ...string) string
```
Context-based translation using language from context.

```go
func (t *Translator) Nc(ctx context.Context, key string, n int, args ...string) string
```
Context-based pluralized translation.

```go
func (t *Translator) ExportJSON(lang string) (string, error)
```
Exports all translations for a language as JSON.

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
