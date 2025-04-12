// Package template provides a flexible templating system that supports layouts and context access.
package template

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Engine represents the template engine.
type Engine struct {
	// Configuration
	dir        string
	defaultExt string
	layout     string

	// Template management
	cache        bool
	templates    map[string]*template.Template
	funcMap      template.FuncMap
	mutex        sync.RWMutex
	cacheManager *CacheManager

	// Filesystem
	fs TemplateFS
}

// Option is a function that configures the Engine.
type Option func(*Engine) error

// New creates a new template engine with the given options.
func New(options ...Option) (*Engine, error) {
	e := &Engine{
		dir:        "views",
		defaultExt: ".html",
		cache:      true,
		templates:  make(map[string]*template.Template),
		funcMap:    make(template.FuncMap),
	}

	// Add default functions
	e.funcMap["yield"] = func() (string, error) {
		return "", fmt.Errorf("yield called outside of layout context")
	}

	e.funcMap["context"] = func(key string) any {
		return nil // Will be replaced during render with actual context access
	}

	e.funcMap["include"] = func(name string, data any) (template.HTML, error) {
		return "", fmt.Errorf("include called outside of render context")
	}

	// Apply all options
	for _, opt := range options {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	// If no filesystem is set, use the default local filesystem
	if e.fs == nil {
		localFS, err := NewLocalFS(e.dir)
		if err != nil {
			return nil, err
		}
		e.fs = localFS
	}

	// Initialize cache manager if caching is enabled and no custom cache manager was set
	if e.cache && e.cacheManager == nil {
		// Default TTL: 1 hour
		e.cacheManager = NewCacheManager(time.Hour)
		// Start cleanup routine to run every 30 minutes
		e.cacheManager.StartCleanupRoutine(30 * time.Minute)
	}

	return e, nil
}

// WithDir sets the directory where templates are stored.
func WithDir(dir string) Option {
	return func(e *Engine) error {
		// Check if directory exists
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidDirectory, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("%w: path is not a directory", ErrInvalidDirectory)
		}

		// Set the directory and create a new local filesystem
		e.dir = dir
		localFS, err := NewLocalFS(dir)
		if err != nil {
			return err
		}
		e.fs = localFS
		return nil
	}
}

// WithExtension sets the template file extension.
func WithExtension(ext string) Option {
	return func(e *Engine) error {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		e.defaultExt = ext
		return nil
	}
}

// WithLayout sets the default layout template.
func WithLayout(layout string) Option {
	return func(e *Engine) error {
		e.layout = layout
		return nil
	}
}

// WithCache enables or disables template caching.
func WithCache(enabled bool) Option {
	return func(e *Engine) error {
		e.cache = enabled
		return nil
	}
}

// WithFuncMap adds custom template functions.
func WithFuncMap(funcMap template.FuncMap) Option {
	return func(e *Engine) error {
		for name, fn := range funcMap {
			e.funcMap[name] = fn
		}
		return nil
	}
}

// AddFunc adds a custom template function.
func (e *Engine) AddFunc(name string, fn any) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.funcMap[name] = fn

	// Clear cache since functions have changed
	if e.cache {
		e.templates = make(map[string]*template.Template)
	}

	return nil
}

// getTemplatePath returns the full path for a template name.
func (e *Engine) getTemplatePath(name string) string {
	// If it already has the extension, don't add it
	if filepath.Ext(name) == "" {
		name = name + e.defaultExt
	}
	return filepath.Join(e.dir, name)
}

// ParseTemplate parses a template with the given name.
func (e *Engine) ParseTemplate(name string) (*template.Template, error) {
	// Use cache manager if available
	if e.cache && e.cacheManager != nil {
		// Try to get from cache manager
		if tmpl := e.cacheManager.Get(name); tmpl != nil {
			return tmpl, nil
		}
	} else if e.cache {
		// Fall back to old caching mechanism if cache manager is not available
		e.mutex.RLock()
		if tmpl, ok := e.templates[name]; ok {
			e.mutex.RUnlock()
			return tmpl, nil
		}
		e.mutex.RUnlock()

		e.mutex.Lock()
		defer e.mutex.Unlock()

		// Check again in case another goroutine loaded it
		if tmpl, ok := e.templates[name]; ok {
			return tmpl, nil
		}
	}

	// Use a mutex lock only when using the old caching mechanism
	if e.cache && e.cacheManager == nil {
		// Lock is already acquired above
	} else {
		e.mutex.Lock()
		defer e.mutex.Unlock()
	}

	// Check if the template exists
	if !e.fs.Exists(name, e.defaultExt) {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}

	// Get the template content from the filesystem
	content, err := e.fs.ParseTemplate(name, e.defaultExt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	// Create a new template
	tmplName := name
	if filepath.Ext(name) == "" {
		tmplName = name + e.defaultExt
	}
	tmpl := template.New(filepath.Base(tmplName)).Funcs(e.funcMap)

	// Parse the template string
	tmpl, err = tmpl.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	// Cache the template
	if e.cache {
		if e.cacheManager != nil {
			// Store in the cache manager
			e.cacheManager.Set(name, tmpl)
		} else {
			// Fall back to old caching mechanism
			e.templates[name] = tmpl
		}
	}

	return tmpl, nil
}

// Render renders a template with the given name and data.
func (e *Engine) Render(w io.Writer, name string, data any, ctx context.Context) error {
	var content string
	var layoutTmpl *template.Template

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

	// Get the content template
	contentTmpl, err := e.ParseTemplate(name)
	if err != nil {
		return err
	}

	// If we have a layout, use it
	if e.layout != "" {
		// Override the yield function to return the rendered content
		funcMap["yield"] = func() (template.HTML, error) {
			return template.HTML(content), nil
		}

		// Parse the layout template
		layoutTmpl, err = e.ParseTemplate(e.layout)
		if err != nil {
			return err
		}

		// Render the content template to a string
		var contentBuffer strings.Builder
		if err := contentTmpl.Funcs(funcMap).Execute(&contentBuffer, data); err != nil {
			return err
		}
		content = contentBuffer.String()

		// Render the layout template
		return layoutTmpl.Funcs(funcMap).Execute(w, data)
	}

	// No layout, just render the content template
	return contentTmpl.Funcs(funcMap).Execute(w, data)
}
