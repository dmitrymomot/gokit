# Go Template Engine

A simple, flexible template engine for Go that supports layouts and provides access to request context from templates.

## Features

- Layout support (master templates)
- Access to request context within templates
- Simple and intuitive API
- Built on top of Go's standard html/template package
- Template caching for improved performance

## Installation

```bash
go get -u github.com/yourusername/gokit/template
```

## Quick Start

```go
package main

import (
	"context"
	"net/http"

	"github.com/yourusername/gokit/template"
)

func main() {
	// Initialize the template engine
	engine, err := template.New(
		template.WithDir("views"),
		template.WithLayout("layouts/main"),
		template.WithExtension(".html"),
	)
	if err != nil {
		panic(err)
	}

	// Set up HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Create context with data
		ctx := context.WithValue(r.Context(), "user", "John Doe")

		// Render template with context
		err := engine.Render(w, "home", map[string]any{
			"title": "Home Page",
			"message": "Welcome to our website!",
		}, ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.ListenAndServe(":8080", nil)
}
```

## Template Example

**layouts/main.html**:

```html
<!DOCTYPE html>
<html>
    <head>
        <title>{{ .title }}</title>
    </head>
    <body>
        <header>
            <h1>My Website</h1>
            <p>Welcome, {{ context "user" }}</p>
        </header>
        <main>{{ yield }}</main>
        <footer>© {{ currentYear }} My Company</footer>
    </body>
</html>
```

**views/home.html**:

```html
<div class="content">
    <h2>{{ .title }}</h2>
    <p>{{ .message }}</p>
</div>
```

## API Reference

### Initialization Options

- `WithDir(dir string)` - Set the directory where templates are stored
- `WithLayout(layout string)` - Set the default layout template
- `WithExtension(ext string)` - Set the template file extension
- `WithFuncMap(funcMap template.FuncMap)` - Add custom template functions
- `WithCache(enabled bool)` - Enable or disable template caching

### Methods

- `Render(w io.Writer, name string, data any, ctx context.Context) error` - Render a template
- `ParseTemplate(name string) (*template.Template, error)` - Parse a template
- `AddFunc(name string, fn any) error` - Add a custom template function

## License

MIT
