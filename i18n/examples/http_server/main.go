package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dmitrymomot/gokit/i18n"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>i18n Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .lang-button { margin-right: 10px; padding: 5px 10px; }
        .content { margin-top: 20px; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>i18n HTTP Example</h1>
    
    <div>
        <p>Try changing your browser's language preference or click:</p>
        <button class="lang-button" onclick="window.location.href='?lang=en'">English</button>
        <button class="lang-button" onclick="window.location.href='?lang=fr'">French</button>
        <button class="lang-button" onclick="window.location.href='?lang=es'">Spanish (Unsupported)</button>
    </div>
    
    <div class="content">
        <h2>{{.Greeting}}</h2>
        <p>{{.Items}}</p>
        <p>{{.NestedGreeting}}</p>
        <p><strong>Detected Language:</strong> {{.DetectedLang}}</p>
        <p><strong>Accept-Language Header:</strong> {{.AcceptLanguage}}</p>
    </div>
</body>
</html>
`

func main() {
	// Create translations directory if it doesn't exist
	err := os.MkdirAll("translations", 0755)
	if err != nil {
		log.Fatalf("Failed to create translations directory: %v", err)
	}

	// Create a sample translations file
	yamlContent := `
en:
  welcome: "Welcome to i18n package, %{name}!"
  items:
    zero: "No items"
    one: "%{count} item"
    other: "%{count} items"
  nested:
    greeting: "Hello from nested key!"

fr:
  welcome: "Bienvenue au package i18n, %{name}!"
  items:
    zero: "Aucun élément"
    one: "%{count} élément"
    other: "%{count} éléments"
  nested:
    greeting: "Bonjour depuis une clé imbriquée!"
`
	err = os.WriteFile(filepath.Join("translations", "messages.yaml"), []byte(yamlContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write translations file: %v", err)
	}

	// Load translations from a directory
	err = i18n.LoadTranslationsDir("translations")
	if err != nil {
		log.Fatalf("Failed to load translations: %v", err)
	}

	// Parse the HTML template
	tmpl, err := template.New("webpage").Parse(htmlTemplate)
	if err != nil {
		log.Fatal(err)
	}

	// Set up the HTTP handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get language from query parameter or Accept-Language header
		var lang string
		queryLang := r.URL.Query().Get("lang")
		acceptLang := r.Header.Get("Accept-Language")
		
		if queryLang != "" {
			lang = queryLang
		} else {
			lang = i18n.BestLangFromAcceptLanguage(acceptLang, "en")
		}

		// Items count based on query parameter or default to 5
		count := r.URL.Query().Get("count")
		if count == "" {
			count = "5"
		}
		
		// Prepare template data
		data := map[string]string{
			"Greeting":      i18n.T(lang, "welcome", "name", "User"),
			"Items":         i18n.N(lang, "items", 5, "count", count),
			"NestedGreeting": i18n.T(lang, "nested.greeting"),
			"DetectedLang":  lang,
			"AcceptLanguage": acceptLang,
		}

		// Execute template
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	port := ":8080"
	fmt.Printf("Starting server on http://localhost%s\n", port)
	fmt.Println("Press Ctrl+C to stop")
	log.Fatal(http.ListenAndServe(port, nil))
}
