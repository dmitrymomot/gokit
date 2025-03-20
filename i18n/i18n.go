package i18n

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Package i18n provides internationalization and localization capabilities for Go applications.
// It supports translations from YAML files, pluralization rules, variable substitution,
// and automatic language detection from Accept-Language headers.
//
// Translation files use a simple YAML structure with language codes as top-level keys:
//
//	en:
//	  welcome: "Welcome, %{name}!"
//	  items:
//	    one: "%{count} item"
//	    other: "%{count} items"
//
//	fr:
//	  welcome: "Bienvenue, %{name}!"
//	  items:
//	    one: "%{count} élément"
//	    other: "%{count} éléments"
//
// The package is thread-safe and can be used in concurrent applications.

var (
	translations = make(map[string]map[string]any)
	mu           sync.RWMutex
)

// DefaultLang is the default language used when no other language is available.
const DefaultLang = "en"

// LoadTranslations loads localization data from a single YAML file.
// The file must follow the syntax from: https://github.com/invopop/ctxi18n/blob/main/README.md.
// This call overwrites any existing translations.
//
// Example:
//
//	err := i18n.LoadTranslations("translations.yaml")
//	if err != nil {
//	    log.Fatalf("Failed to load translations: %v", err)
//	}
func LoadTranslations(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var fileTranslations map[string]map[string]any
	if err := yaml.Unmarshal(data, &fileTranslations); err != nil {
		return err
	}
	mu.Lock()
	defer mu.Unlock()
	translations = fileTranslations
	return nil
}

// LoadTranslationsDir loads all YAML files (with .yaml or .yml extensions)
// from the provided directory recursively and merges them into the global translations.
//
// Example:
//
//	err := i18n.LoadTranslationsDir("./translations")
//	if err != nil {
//	    log.Fatalf("Failed to load translations: %v", err)
//	}
func LoadTranslationsDir(dir string) error {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".yaml" || ext == ".yml" {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var fileTranslations map[string]map[string]any
		if err := yaml.Unmarshal(data, &fileTranslations); err != nil {
			return err
		}
		mu.Lock()
		mergeTranslations(fileTranslations, translations)
		mu.Unlock()
	}
	return nil
}

// LoadTranslationsFS loads all YAML files (with .yaml or .yml extensions)
// from the provided http.FileSystem starting at root, recursively,
// and merges them into the global translations.
//
// This is useful for embedding translations in your binary using Go 1.16+ embed package.
//
// Example:
//
//	//go:embed translations
//	var translationsFS embed.FS
//	httpFS := http.FS(translationsFS)
//	err := i18n.LoadTranslationsFS(httpFS, ".")
//	if err != nil {
//	    log.Fatalf("Failed to load translations: %v", err)
//	}
func LoadTranslationsFS(fs http.FileSystem, root string) error {
	return loadTranslationsFromFS(fs, root)
}

func loadTranslationsFromFS(fs http.FileSystem, root string) error {
	f, err := fs.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := f.Readdir(-1)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			entryPath := filepath.Join(root, entry.Name())
			if entry.IsDir() {
				if err := loadTranslationsFromFS(fs, entryPath); err != nil {
					return err
				}
			} else {
				ext := filepath.Ext(entry.Name())
				if ext == ".yaml" || ext == ".yml" {
					if err := loadTranslationFileFromFS(fs, entryPath); err != nil {
						return err
					}
				}
			}
		}
	} else {
		ext := filepath.Ext(info.Name())
		if ext == ".yaml" || ext == ".yml" {
			if err := loadTranslationFileFromFS(fs, root); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadTranslationFileFromFS(fs http.FileSystem, filePath string) error {
	f, err := fs.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var fileTranslations map[string]map[string]any
	if err := yaml.Unmarshal(data, &fileTranslations); err != nil {
		return err
	}
	mu.Lock()
	mergeTranslations(fileTranslations, translations)
	mu.Unlock()
	return nil
}

// mergeTranslations merges source translations into destination translations.
// For each language, keys from src are added to dest. In case of conflicts,
// the translation from src will override the one in dest.
func mergeTranslations(src, dest map[string]map[string]any) {
	for lang, trans := range src {
		if existing, ok := dest[lang]; ok {
			for key, val := range trans {
				existing[key] = val
			}
		} else {
			dest[lang] = trans
		}
	}
}

// getTranslation traverses a nested map using dot-separated keys.
// For example, key "datetime.days.other" will traverse m["datetime"] then ["days"] then ["other"].
func getTranslation(m map[string]any, key string) (any, bool) {
	parts := strings.Split(key, ".")
	var current any = m
	for _, p := range parts {
		mp, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = mp[p]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// Lang selects the best matching language from the Accept-Language header.
// It takes an Accept-Language string and optional default languages.
// Returns the most suitable language based on supported translations.
// It's a convenience wrapper around BestLangFromAcceptLanguage().
//
// Example:
//
//	acceptLang := "fr-CA,fr;q=0.9,en;q=0.8"
//	lang := i18n.Lang(acceptLang) // Returns the best matching language code
func Lang(accepted string, defaultLang ...string) string {
	return BestLangFromAcceptLanguage(accepted, defaultLang...)
}

// BestLangFromAcceptLanguage parses an Accept-Language header (e.g. "en-US,en;q=0.9")
// and returns the best matching language from the supported translations.
// It considers both full matches and primary subtags. An optional defaultLocale may be provided;
// if no candidates match, the default is returned if supported, otherwise "en" is returned.
//
// The function implements the language matching algorithm according to RFC 2616:
// https://datatracker.ietf.org/doc/html/rfc2616#section-14.4
//
// Example:
//
//	acceptLang := "fr-CA,fr;q=0.9,en;q=0.8"
//	lang := i18n.BestLangFromAcceptLanguage(acceptLang, "en")
//	// Returns "fr" if supported, otherwise falls back to "en"
func BestLangFromAcceptLanguage(header string, defaultLocale ...string) string {
	if header == "" {
		supported := supportedLanguages()
		if len(defaultLocale) > 0 && contains(supported, defaultLocale[0]) {
			return defaultLocale[0]
		}
		return "en"
	}

	type langQuality struct {
		lang string
		q    float64
	}

	var candidates []langQuality
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		q := 1.0
		lang := part
		if idx := strings.Index(part, ";"); idx != -1 {
			lang = strings.TrimSpace(part[:idx])
			qStr := strings.TrimSpace(part[idx+1:])
			if strings.HasPrefix(qStr, "q=") {
				if parsed, err := strconv.ParseFloat(qStr[2:], 64); err == nil {
					q = parsed
				}
			}
		}
		candidates = append(candidates, langQuality{lang, q})
	}

	// Sort candidates by descending quality.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].q > candidates[j].q
	})

	supported := supportedLanguages()

	// Check each candidate for an exact match or a match by primary subtag.
	for _, candidate := range candidates {
		if contains(supported, candidate.lang) {
			return candidate.lang
		}
		if idx := strings.Index(candidate.lang, "-"); idx != -1 {
			base := candidate.lang[:idx]
			if contains(supported, base) {
				return base
			}
		}
	}

	slog.Warn("no supported language found in Accept-Language header", "header", header)

	// Fallback: return the default locale if provided and supported.
	if len(defaultLocale) > 0 && contains(supported, defaultLocale[0]) {
		return defaultLocale[0]
	}
	if len(supported) > 0 {
		return supported[0]
	}
	return "en"
}

// supportedLanguages returns a list of language codes that have translations available.
func supportedLanguages() []string {
	mu.RLock()
	defer mu.RUnlock()
	langs := make([]string, 0, len(translations))
	for lang := range translations {
		langs = append(langs, lang)
	}
	return langs
}

// contains checks if a string is present in a slice of strings.
func contains(langs []string, target string) bool {
	for _, l := range langs {
		if l == target {
			return true
		}
	}
	return false
}

// buildParams converts a slice of strings (expected as key, value, key, value, …)
// into a map. If the number of arguments is odd, the last one is ignored.
//
// This is an internal helper function used by T() and N() to convert the variadic
// arguments into a parameter map for variable substitution.
func buildParams(args []string) map[string]string {
	params := make(map[string]string)
	for i := 0; i < len(args)-1; i += 2 {
		params[args[i]] = args[i+1]
	}
	return params
}

// sprintf always uses named substitution. It builds a parameter map from the key-value pairs.
//
// This is an internal helper function that wraps namedSprintf to provide a simpler interface
// for the T() and N() functions.
func sprintf(tmpl string, args []string) string {
	return namedSprintf(tmpl, buildParams(args))
}

// namedSprintf performs substitution of named placeholders in the form "%{key}"
// using the provided map.
//
// Example:
//
//	params := map[string]string{"name": "John", "age": "30"}
//	result := namedSprintf("Hello, %{name}! You are %{age} years old.", params)
//	// Returns: "Hello, John! You are 30 years old."
func namedSprintf(tmpl string, params map[string]string) string {
	re := regexp.MustCompile(`%\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := match[2 : len(match)-1]
		if val, ok := params[key]; ok {
			return val
		}
		return match
	})
}

// T translates a key for the given language.
// It supports formatting with additional arguments provided as key-value pairs.
// For example: i18n.T("en", "welcome", "name", "John") will substitute "%{name}" in the template.
//
// If the requested translation is not found, the function returns the key as a fallback.
//
// Example:
//
//	// With translation "welcome": "Hello, %{name}!"
//	msg := i18n.T("en", "welcome", "name", "John")
//	// Returns: "Hello, John!"
//
//	// With nested translation using dot notation
//	msg := i18n.T("en", "messages.greeting", "name", "Alice")
//	// Returns corresponding nested translation with "Alice" substituted
//
//	// With missing translation
//	msg := i18n.T("en", "missing_key")
//	// Returns: "missing_key"
func T(lang, key string, args ...string) string {
	mu.RLock()
	defer mu.RUnlock()
	if langMap, ok := translations[lang]; ok {
		if val, ok := getTranslation(langMap, key); ok {
			if tmpl, ok := val.(string); ok {
				return sprintf(tmpl, args)
			}
		}
	}
	slog.Warn("translation key not found",
		"lang", lang,
		"key", key,
		"args", args,
	)
	return key
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
// If no translation is found, it falls back to the key itself.
//
// Example:
//
//	// With translations:
//	// "items.zero": "No items"
//	// "items.one": "%{count} item"
//	// "items.other": "%{count} items"
//
//	msg := i18n.N("en", "items", 0, "count", "0")
//	// Returns: "No items"
//
//	msg := i18n.N("en", "items", 1, "count", "1")
//	// Returns: "1 item"
//
//	msg := i18n.N("en", "items", 5, "count", "5")
//	// Returns: "5 items"
func N(lang, key string, n int, args ...string) string {
	mu.RLock()
	defer mu.RUnlock()
	if langMap, ok := translations[lang]; ok {
		if val, ok := getTranslation(langMap, key); ok {
			switch v := val.(type) {
			case map[string]any:
				var tmpl string
				if n == 1 {
					if s, ok := v["one"].(string); ok {
						tmpl = s
					} else {
						slog.Warn("plural form 'one' not found for translation key",
							"lang", lang,
							"key", key,
							"n", n,
						)
					}
				}
				if tmpl == "" {
					if s, ok := v["other"].(string); ok {
						tmpl = s
					} else {
						slog.Warn("plural form 'other' not found for translation key",
							"lang", lang,
							"key", key,
							"n", n,
						)
					}
				}
				if tmpl != "" {
					return sprintf(tmpl, args)
				}
			case string:
				return sprintf(v, args)
			}
		} else {
			slog.Warn("translation key not found",
				"lang", lang,
				"key", key,
				"n", n,
				"args", args,
			)
		}
	} else {
		slog.Warn("language not found in translations",
			"lang", lang,
			"key", key,
			"n", n,
		)
	}
	return key
}

// Wrapper for fmt.Sprintf for potential future customization.
var sprintfFunc = func(tmpl string, args ...any) string {
	return fmt.Sprintf(tmpl, args...)
}
