package i18nx_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/dmitrymomot/gokit/i18nx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTranslator(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {
			"hello": "Hello",
			"welcome": "Welcome, %{name}!",
			"items": map[string]any{
				"zero":  "No items",
				"one":   "%{count} item",
				"other": "%{count} items",
			},
			"nested": map[string]any{
				"greeting": "Nested greeting",
			},
		},
		"fr": {
			"hello": "Bonjour",
			"welcome": "Bienvenue, %{name}!",
			"items": map[string]any{
				"zero":  "Aucun élément",
				"one":   "%{count} élément",
				"other": "%{count} éléments",
			},
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	
	// Assert that no error occurred
	require.NoError(t, err)
	require.NotNil(t, translator)
}

func TestTranslatorSupportedLanguages(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {"hello": "Hello"},
		"fr": {"hello": "Bonjour"},
		"es": {"hello": "Hola"},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Get supported languages
	languages := translator.SupportedLanguages()
	
	// Assert that the supported languages are correct
	assert.Len(t, languages, 3)
	assert.Contains(t, languages, "en")
	assert.Contains(t, languages, "fr")
	assert.Contains(t, languages, "es")
}

func TestTranslatorHasTranslation(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {
			"hello": "Hello",
			"nested": map[string]any{
				"greeting": "Nested greeting",
			},
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Test existing translations
	assert.True(t, translator.HasTranslation("en", "hello"))
	assert.True(t, translator.HasTranslation("en", "nested.greeting"))
	
	// Test non-existing translations
	assert.False(t, translator.HasTranslation("en", "goodbye"))
	assert.False(t, translator.HasTranslation("fr", "hello"))
	assert.False(t, translator.HasTranslation("en", "nested.missing"))
}

func TestTranslatorT(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {
			"hello": "Hello",
			"welcome": "Welcome, %{name}!",
			"nested": map[string]any{
				"greeting": "Hello, %{name}!",
			},
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Test simple translation
	result := translator.T("en", "hello")
	assert.Equal(t, "Hello", result)
	
	// Test translation with parameters
	result = translator.T("en", "welcome", "name", "John")
	assert.Equal(t, "Welcome, John!", result)
	
	// Test nested translation
	result = translator.T("en", "nested.greeting", "name", "Alice")
	assert.Equal(t, "Hello, Alice!", result)
	
	// Test fallback to key
	result = translator.T("en", "missing")
	assert.Equal(t, "missing", result)
	
	// Test non-existing language
	result = translator.T("fr", "hello")
	assert.Equal(t, "hello", result)
}

func TestTranslatorTWithComplexCases(t *testing.T) {
	// Create a translation map with complex structures
	translations := map[string]map[string]any{
		"en": {
			"complex": map[string]any{
				"nested": map[string]any{
					"deep": map[string]any{
						"value": "Deep nested value with %{param}",
					},
				},
			},
			"mixed_types": map[any]any{
				"key": "Value with %{param}",
			},
			"non_string": 123,
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator with missing translations logging
	translator, err := i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithMissingTranslationsLogging(true),
	)
	require.NoError(t, err)
	
	// Test deeply nested translation
	result := translator.T("en", "complex.nested.deep.value", "param", "test")
	assert.Equal(t, "Deep nested value with test", result)
	
	// Test translation with map[any]any
	result = translator.T("en", "mixed_types.key", "param", "mixed")
	assert.Equal(t, "Value with mixed", result)
	
	// Test with non-string value
	result = translator.T("en", "non_string")
	assert.Equal(t, "non_string", result)
	
	// Test with invalid path format
	result = translator.T("en", "complex.invalid.path")
	assert.Equal(t, "complex.invalid.path", result)
	
	// Test with odd number of parameters
	result = translator.T("en", "complex.nested.deep.value", "param")
	assert.Equal(t, "Deep nested value with %{param}", result)
	
	// Create a translator with fallback to key disabled
	translator, err = i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithFallbackToKey(false),
	)
	require.NoError(t, err)
	
	// Test missing translation with fallback disabled
	result = translator.T("en", "missing.key")
	assert.Equal(t, "", result)
}

func TestTranslatorN(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {
			"items": map[string]any{
				"zero":  "No items",
				"one":   "%{count} item",
				"other": "%{count} items",
			},
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Test zero case
	result := translator.N("en", "items", 0, "count", "0")
	assert.Equal(t, "No items", result)
	
	// Test one case
	result = translator.N("en", "items", 1, "count", "1")
	assert.Equal(t, "1 item", result)
	
	// Test other case
	result = translator.N("en", "items", 5, "count", "5")
	assert.Equal(t, "5 items", result)
	
	// Test fallback to other when specific form is missing
	translations = map[string]map[string]any{
		"en": {
			"items": map[string]any{
				"one":   "%{count} item",
				"other": "%{count} items",
			},
		},
	}
	
	adapter = &i18nx.MapAdapter{
		Data: translations,
	}
	
	translator, err = i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Test fallback from zero to other
	result = translator.N("en", "items", 0, "count", "0")
	assert.Equal(t, "0 items", result)
}

func TestTranslatorNWithComplexCases(t *testing.T) {
	// Create a translation map with realistic nested plural structures
	translations := map[string]map[string]any{
		"en": {
			"nested": map[string]any{
				"items": map[string]any{
					"zero":  "No nested items",
					"one":   "%{count} nested item",
					"other": "%{count} nested items",
				},
			},
			"products": map[string]any{
				"inventory": map[string]any{
					"zero":  "Out of stock",
					"one":   "Last item available (%{count} remaining)",
					"other": "%{count} items in stock",
				},
			},
		},
		"fr": {
			"nested": map[string]any{
				"items": map[string]any{
					"zero":  "Aucun élément imbriqué",
					"one":   "%{count} élément imbriqué",
					"other": "%{count} éléments imbriqués",
				},
			},
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(
		context.Background(), 
		adapter,
	)
	require.NoError(t, err)
	
	t.Run("Basic nested plural functionality", func(t *testing.T) {
		// Test zero case
		result := translator.N("en", "nested.items", 0, "count", "0")
		assert.Equal(t, "No nested items", result)
		
		// Test one case
		result = translator.N("en", "nested.items", 1, "count", "1")
		assert.Equal(t, "1 nested item", result)
		
		// Test other case
		result = translator.N("en", "nested.items", 5, "count", "5")
		assert.Equal(t, "5 nested items", result)
	})
	
	t.Run("Deeply nested plural translations", func(t *testing.T) {
		// Test zero case
		result := translator.N("en", "products.inventory", 0, "count", "0")
		assert.Equal(t, "Out of stock", result)
		
		// Test one case
		result = translator.N("en", "products.inventory", 1, "count", "1")
		assert.Equal(t, "Last item available (1 remaining)", result)
		
		// Test other case
		result = translator.N("en", "products.inventory", 10, "count", "10")
		assert.Equal(t, "10 items in stock", result)
	})
	
	t.Run("Different languages support", func(t *testing.T) {
		// Test French translations
		result := translator.N("fr", "nested.items", 0, "count", "0")
		assert.Equal(t, "Aucun élément imbriqué", result)
		
		result = translator.N("fr", "nested.items", 1, "count", "1")
		assert.Equal(t, "1 élément imbriqué", result)
		
		result = translator.N("fr", "nested.items", 5, "count", "5")
		assert.Equal(t, "5 éléments imbriqués", result)
	})
}

func TestTranslatorNErrorHandling(t *testing.T) {
	// Create a translation map with invalid structures to test error handling
	translations := map[string]map[string]any{
		"en": {
			"invalid_plural": map[string]any{
				"zero": 123,       // Invalid: number instead of string
				"one":  true,      // Invalid: boolean instead of string
			},
			"string_plural": "Not a map but a string", // Invalid for pluralization
		},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator with missing translations logging
	translator, err := i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithMissingTranslationsLogging(true),
	)
	require.NoError(t, err)
	
	t.Run("Invalid plural format", func(t *testing.T) {
		// When a plural key contains non-string values
		result := translator.N("en", "invalid_plural", 0, "count", "0")
		// Should fallback to key
		assert.Equal(t, "invalid_plural", result)
	})
	
	t.Run("String used for plural translation", func(t *testing.T) {
		// When a string is used instead of a map for pluralization
		result := translator.N("en", "string_plural", 0, "count", "0")
		// Should return the string as-is
		assert.Equal(t, "Not a map but a string", result)
	})
	
	t.Run("Missing translation key", func(t *testing.T) {
		// When translation key doesn't exist
		result := translator.N("en", "missing.key", 0, "count", "0")
		// Should fallback to key
		assert.Equal(t, "missing.key", result)
	})
	
	t.Run("Missing language", func(t *testing.T) {
		// When language isn't available
		result := translator.N("de", "invalid_plural", 0, "count", "0")
		// Should fallback to key
		assert.Equal(t, "invalid_plural", result)
	})
	
	t.Run("Fallback to key disabled", func(t *testing.T) {
		// Create a translator with fallback to key disabled
		noFallbackTranslator, err := i18nx.NewTranslator(
			context.Background(), 
			adapter,
			i18nx.WithFallbackToKey(false),
		)
		require.NoError(t, err)
		
		// When translation is missing and fallback is disabled
		result := noFallbackTranslator.N("en", "missing.key", 0, "count", "0")
		// Should return empty string
		assert.Equal(t, "", result)
	})
}

func TestTranslatorParameterHandling(t *testing.T) {
	translations := map[string]map[string]any{
		"en": {
			"greeting": map[string]any{
				"zero":  "Hello, %{name}! You have no messages.",
				"one":   "Hello, %{name}! You have %{count} message.",
				"other": "Hello, %{name}! You have %{count} messages.",
			},
		},
	}

	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	t.Run("Multiple parameters", func(t *testing.T) {
		// Test with multiple parameters
		result := translator.N("en", "greeting", 0, "name", "John", "count", "0")
		assert.Equal(t, "Hello, John! You have no messages.", result)
		
		result = translator.N("en", "greeting", 1, "name", "John", "count", "1")
		assert.Equal(t, "Hello, John! You have 1 message.", result)
		
		result = translator.N("en", "greeting", 5, "name", "John", "count", "5")
		assert.Equal(t, "Hello, John! You have 5 messages.", result)
	})
}

func TestTranslatorLang(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {"hello": "Hello"},
		"fr": {"hello": "Bonjour"},
		"fr-CA": {"hello": "Bonjour Canadien"},
		"es": {"hello": "Hola"},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Test empty header returns default language
	lang := translator.Lang("")
	assert.Equal(t, "en", lang)
	
	// Test exact match
	lang = translator.Lang("fr")
	assert.Equal(t, "fr", lang)
	
	// Test with quality values
	lang = translator.Lang("es;q=0.8,fr;q=0.9")
	assert.Equal(t, "fr", lang)
	
	// Test with region
	lang = translator.Lang("fr-CA,fr;q=0.9,en;q=0.8")
	assert.Equal(t, "fr-CA", lang)
	
	// Test with unsupported language
	lang = translator.Lang("de,en;q=0.8")
	assert.Equal(t, "en", lang)
	
	// Test with default locale parameter
	lang = translator.Lang("", "fr")
	assert.Equal(t, "fr", lang)
	
	// Test with unsupported default locale
	lang = translator.Lang("", "de")
	assert.Equal(t, "en", lang)
}

func TestTranslatorLangWithComplexCases(t *testing.T) {
	// Create a translation map with realistic language variants
	translations := map[string]map[string]any{
		"en": {
			"greeting": "Hello, world!",
			"app_name": "My Application",
		},
		"en-US": {
			"greeting": "Hello, America!",
			"app_name": "My Application (US)",
		},
		"en-GB": {
			"greeting": "Hello, Britain!",
			"app_name": "My Application (UK)",
		},
		"fr": {
			"greeting": "Bonjour, monde!",
			"app_name": "Mon Application",
		},
		"fr-CA": {
			"greeting": "Bonjour, Canada!",
			"app_name": "Mon Application (CA)",
		},
		"es": {
			"greeting": "¡Hola, mundo!",
			"app_name": "Mi Aplicación",
		},
		"zh-Hans": {
			"greeting": "你好，世界！",
			"app_name": "我的应用程序",
		},
		"zh-Hant": {
			"greeting": "你好，世界！",
			"app_name": "我的應用程式",
		},
	}

	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	t.Run("Common browser language preferences", func(t *testing.T) {
		// Chrome en-US typical header
		lang := translator.Lang("en-US,en;q=0.9")
		assert.Equal(t, "en-US", lang)
		
		// Firefox en-GB typical header
		lang = translator.Lang("en-GB,en;q=0.7,en-US;q=0.3")
		assert.Equal(t, "en-GB", lang)
		
		// Safari fr-CA typical header
		lang = translator.Lang("fr-CA,fr;q=0.9,en;q=0.8")
		assert.Equal(t, "fr-CA", lang)
	})
	
	t.Run("Regional variations", func(t *testing.T) {
		// Region-specific variation preferred
		lang := translator.Lang("en-US,en;q=0.8")
		assert.Equal(t, "en-US", lang)
		
		// Requested region not available, fallback to base language
		lang = translator.Lang("es-MX,es;q=0.9")
		assert.Equal(t, "es", lang)
		
		// Multiple language variants
		lang = translator.Lang("fr-FR,fr-CA;q=0.9,fr;q=0.8")
		assert.Equal(t, "fr-CA", lang)
	})
	
	t.Run("Script variations", func(t *testing.T) {
		// Simplified Chinese
		lang := translator.Lang("zh-Hans,zh;q=0.9,en;q=0.8")
		assert.Equal(t, "zh-Hans", lang)
		
		// Traditional Chinese
		lang = translator.Lang("zh-Hant,zh;q=0.9,en;q=0.8")
		assert.Equal(t, "zh-Hant", lang)
	})
	
	t.Run("Multi-language preferences", func(t *testing.T) {
		// Multilingual user with several preferences
		lang := translator.Lang("es;q=1.0,fr;q=0.9,en;q=0.8")
		assert.Equal(t, "es", lang)
		
		// No preferred language available, fallback to highest quality available
		lang = translator.Lang("de;q=1.0,it;q=0.9,fr;q=0.8,en;q=0.7")
		assert.Equal(t, "fr", lang)
	})
}

// TestTranslatorLangErrorHandling tests language detection edge cases and error handling
func TestTranslatorLangErrorHandling(t *testing.T) {
	translations := map[string]map[string]any{
		"en": {"greeting": "Hello"},
		"fr": {"greeting": "Bonjour"},
	}

	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	t.Run("Empty or invalid Accept-Language headers", func(t *testing.T) {
		// Empty header
		lang := translator.Lang("")
		assert.Equal(t, "en", lang, "Should default to 'en' when Accept-Language is empty")
		
		// Header with only commas and whitespace
		lang = translator.Lang("  ,  ,  ")
		assert.Equal(t, "en", lang, "Should default to 'en' when no valid languages are specified")
	})
	
	t.Run("Missing quality values", func(t *testing.T) {
		// Missing q-value (should default to q=1.0)
		lang := translator.Lang("fr,en;q=0.8")
		assert.Equal(t, "fr", lang, "Language without q-value should default to q=1.0")
	})
	
	t.Run("No requested languages available", func(t *testing.T) {
		// None of the requested languages are available
		lang := translator.Lang("de,it,es")
		assert.Equal(t, "en", lang, "Should default to 'en' when no requested languages are available")
	})
}

// TestTranslatorWithEmptyLanguageSet tests behavior with an empty language set
func TestTranslatorWithEmptyLanguageSet(t *testing.T) {
	// Create an empty adapter
	emptyAdapter := &i18nx.MapAdapter{
		Data: make(map[string]map[string]any),
	}
	
	// Create a translator with no languages
	translator, err := i18nx.NewTranslator(context.Background(), emptyAdapter)
	require.NoError(t, err)
	
	// Test language detection with empty language set
	lang := translator.Lang("en-US,en;q=0.9")
	assert.Equal(t, "en", lang, "Should default to 'en' even when no languages are available")
	
	// Test with custom default language
	customDefaultTranslator, err := i18nx.NewTranslator(
		context.Background(), 
		emptyAdapter,
		i18nx.WithDefaultLanguage("fr"),
	)
	require.NoError(t, err)
	
	lang = customDefaultTranslator.Lang("en-US,en;q=0.9")
	assert.Equal(t, "fr", lang, "Should use the configured default language")
}

func TestTranslatorWithOptions(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {"hello": "Hello"},
		"fr": {"hello": "Bonjour"},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator with options
	translator, err := i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithDefaultLanguage("fr"),
		i18nx.WithFallbackToKey(false),
	)
	require.NoError(t, err)
	
	// Test default language is set correctly
	lang := translator.Lang("")
	assert.Equal(t, "fr", lang)
	
	// Test fallback to key is disabled
	result := translator.T("en", "missing")
	assert.Equal(t, "", result)
}

func TestTranslatorWithLoggerOptions(t *testing.T) {
	// Create a simple translation map
	translations := map[string]map[string]any{
		"en": {"hello": "Hello"},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a custom logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	// Create a new translator with logger option
	translator, err := i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithLogger(logger),
	)
	require.NoError(t, err)
	
	// Test a translation to ensure the translator works
	result := translator.T("en", "hello")
	assert.Equal(t, "Hello", result)
	
	// Create a new translator with missing translations logging
	translator, err = i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithMissingTranslationsLogging(true),
	)
	require.NoError(t, err)
	
	// Test a missing translation
	result = translator.T("en", "missing")
	assert.Equal(t, "missing", result)
	
	// Create a new translator with no logging
	translator, err = i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithNoLogging(),
	)
	require.NoError(t, err)
	
	// Test a translation to ensure the translator works
	result = translator.T("en", "hello")
	assert.Equal(t, "Hello", result)
	
	// Test with nil logger (should use default)
	translator, err = i18nx.NewTranslator(
		context.Background(), 
		adapter,
		i18nx.WithLogger(nil),
	)
	require.NoError(t, err)
	
	// Test a translation to ensure the translator works
	result = translator.T("en", "hello")
	assert.Equal(t, "Hello", result)
}

func TestTranslatorWithEmptyAdapter(t *testing.T) {
	// Create an empty adapter
	adapter := &i18nx.MapAdapter{
		Data: nil,
	}

	// Create a new translator
	translator, err := i18nx.NewTranslator(context.Background(), adapter)
	require.NoError(t, err)
	
	// Test that supported languages is empty
	languages := translator.SupportedLanguages()
	assert.Empty(t, languages)
	
	// Test that translations are not found
	assert.False(t, translator.HasTranslation("en", "hello"))
	
	// Test fallback to key
	result := translator.T("en", "hello")
	assert.Equal(t, "hello", result)
}

func TestTranslatorWithInvalidTranslations(t *testing.T) {
	// Create a translation map with an empty language code
	translations := map[string]map[string]any{
		"": {"hello": "Hello"},
	}

	// Create a MapAdapter with the translations
	adapter := &i18nx.MapAdapter{
		Data: translations,
	}

	// Create a new translator
	_, err := i18nx.NewTranslator(context.Background(), adapter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty language code found")
	
	// Create a translation map with a nil translations map
	translations = map[string]map[string]any{
		"en": nil,
	}
	
	adapter = &i18nx.MapAdapter{
		Data: translations,
	}
	
	// Create a new translator
	_, err = i18nx.NewTranslator(context.Background(), adapter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil translations map for language: en")
}
