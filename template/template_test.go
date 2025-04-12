package template_test

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Using local import
	tmpl "github.com/dmitrymomot/gokit/template"
)

func TestTemplateBasicRendering(t *testing.T) {
	// Initialize the template engine
	engine, err := tmpl.New(
		tmpl.WithDir("testdata"),
		tmpl.WithExtension(".html"),
		tmpl.WithDefaultFuncs(),
	)
	require.NoError(t, err)

	// Render a template without layout
	var buffer bytes.Buffer
	err = engine.Render(&buffer, "content", map[string]any{
		"title":   "Test Page",
		"message": "Hello, World!",
	}, nil)
	require.NoError(t, err)

	// Verify the output
	result := buffer.String()
	assert.Contains(t, result, "<h2>Test Page</h2>")
	assert.Contains(t, result, "<p>Hello, World!</p>")
}

func TestTemplateWithLayout(t *testing.T) {
	// Initialize the template engine with layout
	engine, err := tmpl.New(
		tmpl.WithDir("testdata"),
		tmpl.WithExtension(".html"),
		tmpl.WithLayout("layout"),
		tmpl.WithDefaultFuncs(),
	)
	require.NoError(t, err)

	// Create a context with a user
	ctx := context.WithValue(context.Background(), "user", "John Doe")

	// Render a template with layout
	var buffer bytes.Buffer
	err = engine.Render(&buffer, "content", map[string]any{
		"title":   "Test Page",
		"message": "Hello, World!",
	}, ctx)
	require.NoError(t, err)

	// Verify the output
	result := buffer.String()
	assert.Contains(t, result, "<title>Test Page</title>")
	assert.Contains(t, result, "<h2>Test Page</h2>")
	assert.Contains(t, result, "<p>Hello, World!</p>")
	assert.Contains(t, result, "Welcome, John Doe")
	assert.Contains(t, result, "&copy; "+time.Now().Format("2006"))
}

func TestTemplateWithPartials(t *testing.T) {
	// Initialize the template engine
	engine, err := tmpl.New(
		tmpl.WithDir("testdata"),
		tmpl.WithExtension(".html"),
		tmpl.WithDefaultFuncs(),
	)
	require.NoError(t, err)

	// Render a template that includes a partial
	var buffer bytes.Buffer
	err = engine.Render(&buffer, "content", map[string]any{
		"title":   "Test Page",
		"message": "Hello, World!",
	}, nil)
	require.NoError(t, err)

	// Verify the output
	result := buffer.String()
	assert.Contains(t, result, "<div class=\"partial\">")
	assert.Contains(t, result, "<p>This is a partial template</p>")
	assert.Contains(t, result, "<p>Data from parent: Hello, World!</p>")
}

func TestTemplateWithCustomFunctions(t *testing.T) {
	// Initialize the template engine with custom functions
	engine, err := tmpl.New(
		tmpl.WithDir("testdata"),
		tmpl.WithExtension(".html"),
		tmpl.WithFuncMap(template.FuncMap{
			"customFunc": func() string {
				return "Custom function result"
			},
		}),
	)
	require.NoError(t, err)

	// Create a custom template
	customTemplate := `<div>{{ customFunc }}</div>`
	os.WriteFile("testdata/custom.html", []byte(customTemplate), 0644)
	defer os.Remove("testdata/custom.html")

	// Render the custom template
	var buffer bytes.Buffer
	err = engine.Render(&buffer, "custom", nil, nil)
	require.NoError(t, err)

	// Verify the output
	result := buffer.String()
	assert.Contains(t, result, "<div>Custom function result</div>")
}

func TestTemplateFileSystem(t *testing.T) {
	// Initialize with an embedded filesystem (mock it for testing)
	fs := &mockFS{content: "<p>{{ .message }}</p>"}
	engine, err := tmpl.New(
		tmpl.WithFS(fs),
		tmpl.WithExtension(".html"),
	)
	require.NoError(t, err)

	// Render a template using the mock filesystem
	var buffer bytes.Buffer
	err = engine.Render(&buffer, "test", map[string]any{
		"message": "Hello from FS",
	}, nil)
	require.NoError(t, err)

	// Verify the output
	result := buffer.String()
	assert.Equal(t, "<p>Hello from FS</p>", result)
}

// mockFS is a mock implementation of TemplateFS for testing
type mockFS struct {
	content string
}

func (m *mockFS) ParseTemplate(name, extension string) (string, error) {
	return m.content, nil
}

func (m *mockFS) Exists(name, extension string) bool {
	return true
}
