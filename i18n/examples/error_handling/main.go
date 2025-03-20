package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dmitrymomot/gokit/i18n"
)

func main() {
	// Set up a custom logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	i18n.Logger = logger
	
	// Enable logging for missing translations
	i18n.LogMissingTranslations = true

	fmt.Println("=== Error Handling and Validation Examples ===")

	// Create a temporary directory for our examples
	tempDir, err := os.MkdirTemp("", "i18n_examples_*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Example 1: Valid translation file
	fmt.Println("\n--- Example 1: Loading a valid translation file ---")
	validYAML := `
en:
  welcome: "Welcome, %{name}!"
  items:
    zero: "No items"
    one: "%{count} item"
    other: "%{count} items"
`
	validFile := filepath.Join(tempDir, "valid.yaml")
	if err := os.WriteFile(validFile, []byte(validYAML), 0644); err != nil {
		log.Fatalf("Failed to write valid file: %v", err)
	}

	err = i18n.LoadTranslations(validFile)
	if err != nil {
		fmt.Printf("Error loading valid translations: %v\n", err)
	} else {
		fmt.Println("Valid translations loaded successfully")
		fmt.Printf("Translation result: %s\n", i18n.T("en", "welcome", "name", "User"))
	}

	// Example 2: Invalid YAML format
	fmt.Println("\n--- Example 2: Handling invalid YAML format ---")
	invalidYAML := `
en:
  welcome: "Unclosed quote
  items:
    - broken list format
`
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidFile, []byte(invalidYAML), 0644); err != nil {
		log.Fatalf("Failed to write invalid file: %v", err)
	}

	err = i18n.LoadTranslations(invalidFile)
	if err != nil {
		fmt.Printf("Error detected (expected): %v\n", err)
	} else {
		fmt.Println("Invalid file loaded without errors (unexpected)")
	}

	// Example 3: Non-existent file
	fmt.Println("\n--- Example 3: Handling non-existent file ---")
	err = i18n.LoadTranslations(filepath.Join(tempDir, "nonexistent.yaml"))
	if err != nil {
		fmt.Printf("Error detected (expected): %v\n", err)
	} else {
		fmt.Println("Non-existent file loaded without errors (unexpected)")
	}

	// Example 4: Missing translation and language fallbacks
	fmt.Println("\n--- Example 4: Missing translations and language fallbacks ---")
	minimumYAML := `
en:
  hello: "Hello!"
`
	minFile := filepath.Join(tempDir, "minimal.yaml")
	if err := os.WriteFile(minFile, []byte(minimumYAML), 0644); err != nil {
		log.Fatalf("Failed to write minimal file: %v", err)
	}

	if err := i18n.LoadTranslations(minFile); err != nil {
		fmt.Printf("Error loading minimal translations: %v\n", err)
	} else {
		// Test with existing translation
		result := i18n.T("en", "hello")
		fmt.Printf("Existing translation: %s\n", result)

		// Test with missing translation (logs warning, returns key)
		result = i18n.T("en", "missing_key") 
		fmt.Printf("Missing translation: %s\n", result)

		// Test with unsupported language (logs warning, returns key)
		result = i18n.T("fr", "hello")
		fmt.Printf("Unsupported language: %s\n", result)
	}

	// Example 5: Disabling fallback to key
	fmt.Println("\n--- Example 5: Disabling fallback to key ---")
	// Save current setting
	oldFallbackSetting := i18n.FallbackToKey
	
	// Disable fallback
	i18n.FallbackToKey = false
	result := i18n.T("en", "another_missing_key")
	fmt.Printf("Missing translation with fallback disabled: %q\n", result)
	
	// Restore setting
	i18n.FallbackToKey = oldFallbackSetting

	// Example 6: Validation of empty translations
	fmt.Println("\n--- Example 6: Validation of empty translations ---")
	emptyYAML := `{}`
	emptyFile := filepath.Join(tempDir, "empty.yaml")
	if err := os.WriteFile(emptyFile, []byte(emptyYAML), 0644); err != nil {
		log.Fatalf("Failed to write empty file: %v", err)
	}

	err = i18n.LoadTranslations(emptyFile)
	if err != nil {
		fmt.Printf("Error loading empty translations: %v\n", err)
	} else {
		fmt.Println("Empty translations loaded with warning")
	}

	// Example 7: Invalid language code
	fmt.Println("\n--- Example 7: Invalid language code ---")
	invalidLangYAML := `
"": 
  test: "This has an empty language code"
`
	invalidLangFile := filepath.Join(tempDir, "invalid_lang.yaml")
	if err := os.WriteFile(invalidLangFile, []byte(invalidLangYAML), 0644); err != nil {
		log.Fatalf("Failed to write invalid lang file: %v", err)
	}

	err = i18n.LoadTranslations(invalidLangFile)
	if err != nil {
		fmt.Printf("Error detected (expected): %v\n", err)
	} else {
		fmt.Println("Invalid language code loaded without errors (unexpected)")
	}

	fmt.Println("\nAll examples completed.")
}
