package i18n

import (
	"context"
	"net/http"
)

// Translator is an interface that defines methods for language translation.
type translator interface {
	// Lang determines the language based on the provided header.
	// It accepts a header string and optional default locale values.
	// If no valid language can be determined from the header,
	// the first defaultLocale will be used (if provided).
	// Returns the language code as a string.
	Lang(header string, defaultLocale ...string) string
}

// LangExtractor is a function type that extracts language information from an HTTP request.
// It takes an *http.Request as input and returns a string representing the language code.
// This is typically used to determine the user's preferred language for localization.
type langExtractor func(r *http.Request) string

// ContextKeyLocale is the context key for the locale in the request context.
var ContextKeyLocale = struct{ name string }{name: "gokit/i18n:locale"}

// SetLocale sets the locale in the context.
func SetLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, ContextKeyLocale, locale)
}

// GetLocale returns the locale from the context.
// If no locale is set, will return default locale - "en".
func GetLocale(ctx context.Context) string {
	if locale, ok := ctx.Value(ContextKeyLocale).(string); ok {
		return locale
	}
	return "en"
}

// Middleware returns an HTTP middleware that determines the client's preferred language
// and stores it in the request context.
//
// Parameters:
//   - t: A translator instance that converts Accept-Language headers to language codes
//   - extr: An optional langExtractor function that can extract language from the request
//
// The middleware first attempts to determine the language using the provided extractor function.
// If no extractor is provided or it returns an empty string, the middleware falls back to
// using the Accept-Language header with the translator.
//
// The determined language is stored in the request context using SetLocale and
// can be retrieved later with GetLocale.
func Middleware(t translator, extr langExtractor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var lang string

			if extr != nil {
				lang = extr(r)
			}

			if lang == "" {
				acceptLanguage := r.Header.Get("Accept-Language")
				lang = t.Lang(acceptLanguage)
			}

			next.ServeHTTP(w, r.WithContext(SetLocale(r.Context(), lang)))
		})
	}
}
