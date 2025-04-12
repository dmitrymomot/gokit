package template

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"
)

// RenderPartial renders a partial template with the given name and data.
func (e *Engine) RenderPartial(w io.Writer, name string, data any, ctx context.Context) error {
	// Create a new FuncMap with context-specific functions
	funcMap := make(template.FuncMap)
	for k, v := range e.funcMap {
		funcMap[k] = v
	}

	// Add context function that can access the request context
	funcMap["context"] = func(key any) any {
		if ctx == nil {
			return nil
		}
		return ctx.Value(key)
	}

	// Implement include function that allows inclusion of other templates
	funcMap["include"] = func(includeName string, includeData any) (template.HTML, error) {
		var buf strings.Builder

		// If no data is provided, use the parent data
		if includeData == nil {
			includeData = data
		}

		// Parse and render the included template
		if err := e.RenderPartial(&buf, includeName, includeData, ctx); err != nil {
			return "", fmt.Errorf("failed to include template %s: %w", includeName, err)
		}

		return template.HTML(buf.String()), nil
	}

	// Get the template
	tmpl, err := e.ParseTemplate(name)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrRenderFailed, err)
	}

	// Render the template
	return tmpl.Funcs(funcMap).Execute(w, data)
}

// WithPartials registers partial templates for use with the include function.
func WithPartials(partialsDir string) Option {
	return func(e *Engine) error {
		// No need to pre-load partials, they will be loaded on demand
		return nil
	}
}
