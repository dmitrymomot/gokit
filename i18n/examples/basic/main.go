package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dmitrymomot/gokit/i18n"
)

func main() {
	// Create a translations directory if it doesn't exist
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

	// Basic translation example
	fmt.Println("English:")
	fmt.Println(i18n.T("en", "welcome", "name", "John"))
	fmt.Println(i18n.N("en", "items", 0, "count", "0"))
	fmt.Println(i18n.N("en", "items", 1, "count", "1"))
	fmt.Println(i18n.N("en", "items", 5, "count", "5"))
	fmt.Println(i18n.T("en", "nested.greeting"))

	fmt.Println("\nFrench:")
	fmt.Println(i18n.T("fr", "welcome", "name", "Marie"))
	fmt.Println(i18n.N("fr", "items", 0, "count", "0"))
	fmt.Println(i18n.N("fr", "items", 1, "count", "1"))
	fmt.Println(i18n.N("fr", "items", 5, "count", "5"))
	fmt.Println(i18n.T("fr", "nested.greeting"))

	// Fallback example
	fmt.Println("\nFallback example:")
	fmt.Println(i18n.T("es", "welcome", "name", "Pedro")) // Unsupported language falls back to key
	fmt.Println(i18n.T("en", "missing.key"))              // Missing key falls back to key
}
