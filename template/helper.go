package template

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// DefaultFuncMap returns a set of default template functions.
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// Date/time functions
		"formatDate": func(t time.Time, layout string) string {
			return t.Format(layout)
		},
		"now": func() time.Time {
			return time.Now()
		},
		"currentYear": func() int {
			return time.Now().Year()
		},

		// String functions
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"title":     strings.Title,
		"trim":      strings.TrimSpace,
		"replace":   strings.Replace,
		"split":     strings.Split,
		"join":      strings.Join,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// URL functions
		"urlEncode": func(s string) string {
			return template.URLQueryEscaper(s)
		},

		// HTML functions
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"safeCSS": func(s string) template.CSS {
			return template.CSS(s)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},

		// Conditional functions
		"eq": func(a, b any) bool {
			return a == b
		},
		"neq": func(a, b any) bool {
			return a != b
		},
		"lt": func(a, b any) bool {
			return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
		},
		"lte": func(a, b any) bool {
			return fmt.Sprintf("%v", a) <= fmt.Sprintf("%v", b)
		},
		"gt": func(a, b any) bool {
			return fmt.Sprintf("%v", a) > fmt.Sprintf("%v", b)
		},
		"gte": func(a, b any) bool {
			return fmt.Sprintf("%v", a) >= fmt.Sprintf("%v", b)
		},

		// Math functions
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
		},
	}
}

// WithDefaultFuncs adds the default function map to the engine.
func WithDefaultFuncs() Option {
	return func(e *Engine) error {
		for name, fn := range DefaultFuncMap() {
			e.funcMap[name] = fn
		}
		return nil
	}
}
