# i18n

The `i18n` package provides a simple, powerful internationalization (i18n) solution for Go applications. It supports multiple languages, custom translation files, and extensive error handling to make your applications globally accessible.

## Features

- 🌎 Support for multiple languages
- 🔄 Dynamic language switching
- 📁 Custom translation file formats
- 🚨 Comprehensive error handling with custom error types
- ✅ Validation for translation files
- 📊 Configurable logging for diagnostics
- 🔍 Support for common browser Accept-Language header patterns
- 🧪 Well-tested with realistic usage scenarios

## Installation

```bash
go get -u github.com/dmitrymomot/gokit/i18n
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/dmitrymomot/gokit/i18n"
)

func main() {
	// Initialize translator with English as default language
	translator, err := i18n.New(i18n.Config{
		DefaultLanguage: "en",
		SupportedLanguages: []string{"en", "fr", "es"},
		TranslationsPath: "./translations", // Path to translation files
	})
	if err != nil {
		log.Fatalf("Failed to initialize translator: %v", err)
	}

	// Get translation in default language
	greeting, err := translator.Translate("greeting")
	if err != nil {
		log.Printf("Translation error: %v", err)
	}

	fmt.Println(greeting) // Output: Hello, world!

	// Get translation in a specific language
	frenchGreeting, err := translator.TranslateWithLang("greeting", "fr")
	if err != nil {
		log.Printf("Translation error: %v", err)
	}

	fmt.Println(frenchGreeting) // Output: Bonjour, le monde!
}
```

### Translation Files

Translation files should be organized by language code in the translations directory. For example:

```
/translations
  /en
    common.json
    errors.json
  /fr
    common.json
    errors.json
  /es
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

Example `fr/common.json`:

```json
{
    "greeting": "Bonjour, le monde!",
    "welcome": "Bienvenue dans notre application, {{name}}!",
    "farewell": "Au revoir, à bientôt!"
}
```

### Working with Variables

```go
// Translate with variables
welcomeMsg, err := translator.TranslateWithVars("welcome", map[string]interface{}{
	"name": "John",
})
fmt.Println(welcomeMsg) // Output: Welcome to our application, John!

// Translate with variables in a specific language
frenchWelcome, err := translator.TranslateWithLangAndVars("welcome", "fr", map[string]interface{}{
	"name": "John",
})
fmt.Println(frenchWelcome) // Output: Bienvenue dans notre application, John!
```

### Error Handling

The package provides several custom error types for better error handling:

```go
// Handle specific error types
greeting, err := translator.TranslateWithLang("greeting", "de")
if err != nil {
	switch {
	case errors.Is(err, i18n.ErrLanguageNotSupported):
		log.Printf("Language is not supported: %v", err)
		// Fall back to default language
		greeting, _ = translator.Translate("greeting")
	case errors.Is(err, i18n.ErrTranslationNotFound):
		log.Printf("Translation not found: %v", err)
		greeting = "Hello!" // Fallback text
	case errors.Is(err, i18n.ErrInvalidTranslationFormat):
		log.Printf("Invalid translation format: %v", err)
	case errors.Is(err, i18n.ErrFileSystemError):
		log.Printf("File system error: %v", err)
	default:
		log.Printf("Unknown error: %v", err)
	}
}
```

### Working with HTTP Requests

The package supports parsing Accept-Language headers from HTTP requests:

```go
func handler(w http.ResponseWriter, r *http.Request) {
	// Parse Accept-Language header from request
	lang := i18n.ParseAcceptLanguage(r.Header.Get("Accept-Language"),
		[]string{"en", "fr", "es"}, "en")

	// Use the detected language for translations
	greeting, err := translator.TranslateWithLang("greeting", lang)
	if err != nil {
		// Handle error
		greeting, _ = translator.Translate("greeting") // Fallback to default
	}

	fmt.Fprintf(w, greeting)
}
```

### Advanced Configuration

```go
translator, err := i18n.New(i18n.Config{
	DefaultLanguage: "en",
	SupportedLanguages: []string{"en", "fr", "es", "de", "ja"},
	TranslationsPath: "./translations",
	FileFormat: "json",          // Default is "json", can be customized
	EnableLogging: true,         // Enable logging for debugging
	LogLevel: i18n.LogLevelInfo, // Set log level
	FallbackToDefault: true,     // Fall back to default language if translation not found
	CustomLoader: myCustomLoader, // Optional custom translation loader
})
```

### HTTP Middleware

The package includes middleware support for integrating with HTTP servers to automatically detect and set the user's preferred language based on request headers.

#### Standard Go HTTP Middleware

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dmitrymomot/gokit/i18n"
)

func main() {
	// Initialize translator
	translator, err := i18n.New(i18n.Config{
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en", "fr", "es"},
		TranslationsPath:   "./translations",
	})
	if err != nil {
		log.Fatalf("Failed to initialize translator: %v", err)
	}

	// Create a custom language extractor (optional)
	// This example extracts language from a URL query parameter
	customExtractor := func(r *http.Request) string {
		return r.URL.Query().Get("lang")
	}

	// Create a handler that uses translations
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the locale from the request context
		locale := i18n.GetLocale(r.Context())
		
		// Use the detected language for translations
		greeting, err := translator.TranslateWithLang("greeting", locale)
		if err != nil {
			greeting, _ = translator.Translate("greeting") // Fallback to default
		}
		
		fmt.Fprintf(w, "Detected language: %s\n", locale)
		fmt.Fprintf(w, "Greeting: %s\n", greeting)
	})

	// Apply the i18n middleware to automatically detect language
	// Pass nil instead of customExtractor if you want to rely only on Accept-Language header
	http.Handle("/", i18n.Middleware(translator, customExtractor)(handler))
	
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

#### Context Utilities

The package provides utility functions for working with locale in request contexts:

```go
// Set locale in context
ctx = i18n.SetLocale(context.Background(), "fr")

// Get locale from context (returns "en" as default if not set)
locale := i18n.GetLocale(ctx)
```

#### Usage with Third-Party Frameworks

While the package provides middleware for standard Go HTTP servers, you can easily adapt it for use with third-party web frameworks by:

1. Creating a custom middleware for your framework
2. Using the `SetLocale` and `GetLocale` functions to store and retrieve the locale
3. Using the translator's `Lang` method to parse Accept-Language headers

Example with a custom framework middleware:

```go
// Pseudocode for framework middleware
func I18nMiddleware(translator *i18n.Translator) FrameworkMiddleware {
	return func(c *Context) error {
		// Extract language from request headers
		acceptLanguage := c.Request.Header.Get("Accept-Language")
		lang := translator.Lang(acceptLanguage)
		
		// Store in framework context
		c.Set("locale", lang)
		
		return c.Next()
	}
}

// In your handler
func Handler(c *Context) error {
	locale := c.Get("locale").(string)
	greeting, _ := translator.TranslateWithLang("greeting", locale)
	return c.String(200, greeting)
}
```

## Error Types

The package defines the following error types:

- `ErrLanguageNotSupported`: Requested language is not in the supported languages list
- `ErrTranslationNotFound`: Translation key doesn't exist in the specified language
- `ErrInvalidTranslationFormat`: Translation file has an invalid format
- `ErrFileSystemError`: Error accessing or reading translation files

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## See Also

This package is part of the [gokit](https://github.com/dmitrymomot/gokit) library, which provides various utility packages for Go applications.
