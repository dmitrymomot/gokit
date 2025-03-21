package i18nx

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Translator is the core type of the i18nx package that provides internationalization
// capabilities with an object-oriented design.
type Translator struct {
	translations   map[string]map[string]any
	defaultLang    string
	fallbackToKey  bool
	missingLogMode bool
	logger         *slog.Logger
	mu             sync.RWMutex
}

// NewTranslator creates a new Translator instance with the given adapter and options.
func NewTranslator(ctx context.Context, adapter TranslationAdapter, options ...Option) (*Translator, error) {
	t := &Translator{
		translations:   make(map[string]map[string]any),
		defaultLang:    "en",
		fallbackToKey:  true,
		missingLogMode: false,
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	// Apply options
	for _, option := range options {
		option(t)
	}

	// Load translations from adapter
	translations, err := adapter.Load(ctx)
	if err != nil {
		return nil, err
	}

	// Validate translations
	if err := t.validateTranslations(translations); err != nil {
		return nil, err
	}

	t.translations = translations
	t.logger.Info("Translations loaded", "languages", t.mapKeys(translations))
	return t, nil
}

// validateTranslations checks if the translations map has a valid structure.
// It ensures that language codes are valid and that translations are properly formatted.
func (t *Translator) validateTranslations(trans map[string]map[string]any) error {
	if len(trans) == 0 {
		t.logger.Warn("No translations provided")
		return nil
	}

	for lang, translations := range trans {
		if lang == "" {
			return fmt.Errorf("empty language code found")
		}
		if translations == nil {
			return fmt.Errorf("nil translations map for language: %s", lang)
		}
	}
	return nil
}

// mapKeys returns a slice containing all keys of the provided map.
func (t *Translator) mapKeys(m map[string]map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// getTranslation traverses a nested map using dot-separated keys.
// For example, key "datetime.days.other" will traverse m["datetime"] then ["days"] then ["other"].
func (t *Translator) getTranslation(m map[string]any, key string) (any, bool) {
	parts := strings.Split(key, ".")
	current := m

	for i, part := range parts {
		if i == len(parts)-1 {
			val, ok := current[part]
			return val, ok
		}

		next, ok := current[part]
		if !ok {
			return nil, false
		}

		currentMap, ok := next.(map[string]any)
		if !ok {
			// Try to convert from map[any]any to map[string]any
			anyMap, isAnyMap := next.(map[any]any)
			if !isAnyMap {
				return nil, false
			}

			currentMap = make(map[string]any, len(anyMap))
			for k, v := range anyMap {
				if ks, ok := k.(string); ok {
					currentMap[ks] = v
				}
			}
		}

		current = currentMap
	}

	return nil, false
}

// SupportedLanguages returns a list of language codes that have translations available.
func (t *Translator) SupportedLanguages() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mapKeys(t.translations)
}

// HasTranslation checks if a translation exists for the given language and key.
func (t *Translator) HasTranslation(lang, key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	langMap, ok := t.translations[lang]
	if !ok {
		return false
	}

	_, ok = t.getTranslation(langMap, key)
	return ok
}

// Lang selects the best matching language from the Accept-Language header.
// It takes an Accept-Language string and optional default languages.
// Returns the most suitable language based on supported translations.
//
// Example:
//
//	acceptLang := "fr-CA,fr;q=0.9,en;q=0.8"
//	lang := translator.Lang(acceptLang) // Returns the best matching language code
func (t *Translator) Lang(header string, defaultLocale ...string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if header == "" {
		defLang := t.defaultLang
		if len(defaultLocale) > 0 && defaultLocale[0] != "" {
			defLang = defaultLocale[0]
		}

		if slices.Contains(t.supportedLanguages(), defLang) {
			return defLang
		}
		return t.defaultLang
	}

	supported := t.supportedLanguages()
	if len(supported) == 0 {
		return t.defaultLang
	}

	langQ := make(map[string]float64)
	parts := strings.SplitN(header, ",", -1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split language and quality
		langAndQ := strings.Split(part, ";")
		lang := langAndQ[0]
		q := 1.0 // Default quality is 1.0

		// Parse quality if provided
		if len(langAndQ) > 1 {
			qPart := strings.TrimSpace(langAndQ[1])
			if strings.HasPrefix(qPart, "q=") {
				qVal, err := strconv.ParseFloat(qPart[2:], 64)
				if err == nil {
					q = qVal
				}
			}
		}

		// Extract language without region
		if idx := strings.Index(lang, "-"); idx != -1 {
			primaryLang := lang[:idx]
			// Add both language-region and just language
			langQ[lang] = q
			// Only add primary language if not already present or with higher q
			if existingQ, exists := langQ[primaryLang]; !exists || q > existingQ {
				langQ[primaryLang] = q
			}
		} else {
			langQ[lang] = q
		}
	}

	// Sort languages by quality
	type langWithQ struct {
		lang string
		q    float64
	}
	languages := make([]langWithQ, 0, len(langQ))
	for lang, q := range langQ {
		languages = append(languages, langWithQ{lang, q})
	}
	sort.Slice(languages, func(i, j int) bool {
		return languages[i].q > languages[j].q
	})

	// Find the first matching language
	for _, lq := range languages {
		if slices.Contains(supported, lq.lang) {
			return lq.lang
		}
	}

	// Return default if specified and supported
	if len(defaultLocale) > 0 && defaultLocale[0] != "" && slices.Contains(supported, defaultLocale[0]) {
		return defaultLocale[0]
	}

	// Last resort: return the default language
	return t.defaultLang
}

// supportedLanguages returns a list of language codes that have translations available.
func (t *Translator) supportedLanguages() []string {
	langs := make([]string, 0, len(t.translations))
	for lang := range t.translations {
		langs = append(langs, lang)
	}
	sort.Strings(langs)
	return langs
}

// buildParams converts a slice of strings (expected as key, value, key, value, …)
// into a map. If the number of arguments is odd, the last one is ignored.
func (t *Translator) buildParams(args []string) map[string]string {
	params := make(map[string]string)
	for i := 0; i < len(args)-1; i += 2 {
		params[args[i]] = args[i+1]
	}
	return params
}

// sprintf always uses named substitution. It builds a parameter map from the key-value pairs.
func (t *Translator) sprintf(tmpl string, args []string) string {
	params := t.buildParams(args)
	return t.namedSprintf(tmpl, params)
}

// Regex to find named parameters in the form %{name}
var paramRegex = regexp.MustCompile(`%\{([^}]+)\}`)

// namedSprintf performs substitution of named placeholders in the form "%{key}"
// using the provided map.
func (t *Translator) namedSprintf(tmpl string, params map[string]string) string {
	result := paramRegex.ReplaceAllStringFunc(tmpl, func(match string) string {
		// Extract parameter name
		name := match[2 : len(match)-1]
		// Replace with parameter value if exists
		if val, ok := params[name]; ok {
			return val
		}
		// Keep original placeholder if parameter not found
		return match
	})
	return result
}

// T translates a key for the given language.
// It supports formatting with additional arguments provided as key-value pairs.
// For example: translator.T("en", "welcome", "name", "John") will substitute "%{name}" in the template.
//
// If the requested translation is not found and FallbackToKey is true, the function returns
// the key as a fallback. Otherwise, it returns an empty string and logs the error if
// missingLogMode is enabled.
//
// Example:
//
//	// With translation "welcome": "Hello, %{name}!"
//	msg := translator.T("en", "welcome", "name", "John")
//	// Returns: "Hello, John!"
//
//	// With nested translation using dot notation
//	msg := translator.T("en", "messages.greeting", "name", "Alice")
//	// Returns corresponding nested translation with "Alice" substituted
func (t *Translator) T(lang, key string, args ...string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check if the language is supported
	langMap, ok := t.translations[lang]
	if !ok {
		if t.missingLogMode {
			t.logger.Warn("Language not supported", "lang", lang, "key", key)
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
		return ""
	}

	// Get the translation
	val, ok := t.getTranslation(langMap, key)
	if !ok {
		if t.missingLogMode {
			t.logger.Warn("Translation not found", "lang", lang, "key", key)
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
		return ""
	}

	// Handle different types of translation values
	switch v := val.(type) {
	case string:
		return t.sprintf(v, args)
	case map[string]any, map[any]any:
		if t.missingLogMode {
			t.logger.Warn("Translation is not a string", "lang", lang, "key", key, "type", fmt.Sprintf("%T", v))
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
	default:
		// Try to convert to string
		if s, ok := val.(fmt.Stringer); ok {
			return t.sprintf(s.String(), args)
		}

		if t.missingLogMode {
			t.logger.Warn("Translation is not a string", "lang", lang, "key", key, "type", fmt.Sprintf("%T", v))
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
	}

	return ""
}

// N translates a key with pluralization for the given language.
// The parameter n is used to select the plural form. It supports formatting with additional
// arguments provided as key-value pairs.
//
// The function first tries the exact key with the appropriate plural suffix:
// - For n=0, it tries key+".zero" first, falling back to key+".other"
// - For n=1, it tries key+".one"
// - For all other values, it uses key+".other"
//
// If no translation is found and fallbackToKey is true, it falls back to the key itself.
// Otherwise, it returns an empty string and logs the error if missingLogMode is enabled.
//
// Example:
//
//	// With translations:
//	// "items.zero": "No items"
//	// "items.one": "%{count} item"
//	// "items.other": "%{count} items"
//
//	msg := translator.N("en", "items", 0, "count", "0")
//	// Returns: "No items"
//
//	msg := translator.N("en", "items", 1, "count", "1")
//	// Returns: "1 item"
//
//	msg := translator.N("en", "items", 5, "count", "5")
//	// Returns: "5 items"
func (t *Translator) N(lang, key string, n int, args ...string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Check if the language is supported
	langMap, ok := t.translations[lang]
	if !ok {
		if t.missingLogMode {
			t.logger.Warn("Language not supported", "lang", lang, "key", key, "n", n)
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
		return ""
	}

	// Try to get the translation with appropriate plural form
	var val any
	var found bool

	// For n=0, try "zero" form first
	if n == 0 {
		val, found = t.getTranslation(langMap, key+".zero")
		if found {
			goto translate
		}
		// Fall back to "other" form for n=0
		val, found = t.getTranslation(langMap, key+".other")
		if found {
			goto translate
		}
	}

	// For n=1, try "one" form
	if n == 1 {
		val, found = t.getTranslation(langMap, key+".one")
		if found {
			goto translate
		}
	}

	// For n>1, use "other" form
	if n != 0 && n != 1 {
		val, found = t.getTranslation(langMap, key+".other")
		if found {
			goto translate
		}
	}

	// Try the key itself (might be a string with embedded pluralization logic)
	val, found = t.getTranslation(langMap, key)
	if !found {
		if t.missingLogMode {
			t.logger.Warn("Pluralization not found", "lang", lang, "key", key, "n", n)
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
		return ""
	}

translate:
	switch v := val.(type) {
	case string:
		// Always include the count in args if not already present
		hasCount := false
		for i := 0; i < len(args)-1; i += 2 {
			if args[i] == "count" {
				hasCount = true
				break
			}
		}
		if !hasCount {
			newArgs := make([]string, len(args)+2)
			copy(newArgs, args)
			newArgs[len(args)] = "count"
			newArgs[len(args)+1] = strconv.Itoa(n)
			args = newArgs
		}
		return t.sprintf(v, args)
	default:
		if t.missingLogMode {
			t.logger.Warn("Pluralization translation is not a string", "lang", lang, "key", key, "n", n, "type", fmt.Sprintf("%T", v))
		}
		if t.fallbackToKey {
			return t.sprintf(key, args)
		}
		return ""
	}
}
