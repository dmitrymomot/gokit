package i18n_test

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/dmitrymomot/gokit/i18n"
)

// resetTranslations loads an empty translations file to clear any existing translations.
func resetTranslations(t *testing.T) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "i18n_reset_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.WriteString("{}"); err != nil {
		_ = tmpFile.Close()
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal("error closing temp file in resetTranslations:", err)
	}
	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})
	if err := i18n.LoadTranslations(tmpFile.Name()); err != nil {
		t.Fatal(err)
	}
}

// createTempYAML creates a temporary YAML file with the given content.
func createTempYAML(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "i18n_test_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		t.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatal("error closing temp file in createTempYAML:", err)
	}
	filename := tmpFile.Name()
	t.Cleanup(func() {
		os.Remove(filename)
	})
	return filename
}

// TestEmptyTranslations verifies that an empty translation file results in fallback to keys.
func TestEmptyTranslations(t *testing.T) {
	resetTranslations(t)
	emptyContent := "{}"
	filename := createTempYAML(t, emptyContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}
	result := i18n.T("en", "anykey", "unused", "arg")
	if result != "anykey" {
		t.Errorf("Expected fallback to key 'anykey', got %q", result)
	}
}

// TestT verifies that T returns the proper localized string with named substitution.
func TestT(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  welcome: "Hi, %{name}"
fr:
  welcome: "Salut, %{name}"
`
	filename := createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	result := i18n.T("en", "welcome", "name", "John")
	if result != "Hi, John" {
		t.Errorf("Expected 'Hi, John', got %q", result)
	}

	result = i18n.T("fr", "welcome", "name", "Jean")
	if result != "Salut, Jean" {
		t.Errorf("Expected 'Salut, Jean', got %q", result)
	}

	// Unsupported language: fallback to key.
	result = i18n.T("es", "welcome", "name", "Juan")
	if result != "welcome" {
		t.Errorf("Expected fallback to key 'welcome' for unsupported language, got %q", result)
	}

	// Missing key.
	result = i18n.T("en", "nonexistent", "name", "Test")
	if result != "nonexistent" {
		t.Errorf("Expected 'nonexistent', got %q", result)
	}
}

// TestN verifies that pluralization works.
func TestN(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  datetime:
    days:
      one: "%{count} day"
      other: "%{count} days"
    hours:
      one: "%{count} hour"
      other: "%{count} hours"
    minutes:
      zero: "less than a minute"
      one: "%{count} minute"
      other: "%{count} minutes"
  greeting:
    one: "You have %{count} message"
    other: "You have %{count} messages"
`
	filename := createTempYAML(t, yamlContent)
	err := i18n.LoadTranslations(filename)
	if err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	result := i18n.N("en", "greeting", 5, "count", "5")
	if result != "You have 5 messages" {
		t.Errorf("Expected 'You have 5 messages', got %q", result)
	}

	result = i18n.N("en", "greeting", 1, "count", "1")
	if result != "You have 1 message" {
		t.Errorf("Expected 'You have 1 message', got %q", result)
	}

	result = i18n.N("en", "datetime.days", 5, "count", "5")
	if result != "5 days" {
		t.Errorf("Expected '5 days', got %q", result)
	}

	result = i18n.N("en", "datetime.days", 1, "count", "1")
	if result != "1 day" {
		t.Errorf("Expected '1 day', got %q", result)
	}
}

// TestNamedParameters verifies that named parameters work in pluralized strings.
func TestNamedParameters(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  greeting:
    one: "You have %{count} message"
    other: "You have %{count} messages"
  datetime:
    days:
      one: "%{count} day"
      other: "%{count} days"
    hours:
      one: "%{count} hour"
      other: "%{count} hours"
    minutes:
      zero: "less than a minute"
      one: "%{count} minute"
      other: "%{count} minutes"
`
	filename := createTempYAML(t, yamlContent)
	err := i18n.LoadTranslations(filename)
	if err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	result := i18n.N("en", "datetime.days", 5, "count", "5")
	if result != "5 days" {
		t.Errorf("Expected '5 days', got %q", result)
	}

	result = i18n.N("en", "datetime.hours", 1, "count", "1")
	if result != "1 hour" {
		t.Errorf("Expected '1 hour', got %q", result)
	}

	result = i18n.N("en", "datetime.minutes", 0, "count", "0")
	if result != "less than a minute" {
		t.Errorf("Expected 'less than a minute', got %q", result)
	}
}

// TestLoadTranslations verifies that translations are loaded correctly from a single file.
func TestLoadTranslations(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  hello: "Hello, %{name}!"
fr:
  hello: "Bonjour, %{name}!"
`
	filename := createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	result := i18n.T("en", "hello", "name", "World")
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", result)
	}

	result = i18n.T("fr", "hello", "name", "Monde")
	if result != "Bonjour, Monde!" {
		t.Errorf("Expected 'Bonjour, Monde!', got %q", result)
	}
}

// TestLoadTranslationsDirectoryWithSubdirs verifies that translations are recursively loaded from directories.
func TestLoadTranslationsDirectoryWithSubdirs(t *testing.T) {
	resetTranslations(t)
	dir, err := os.MkdirTemp("", "i18n_test_dir")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	file1 := filepath.Join(dir, "file1.yaml")
	content1 := `
en:
  greeting: "Hello, %{name}!"
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}

	subDir := filepath.Join(dir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	file2 := filepath.Join(subDir, "file2.yaml")
	content2 := `
fr:
  greeting: "Bonjour, %{name}!"
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}

	if err := i18n.LoadTranslationsDir(dir); err != nil {
		t.Fatalf("LoadTranslationsDir failed: %v", err)
	}

	result := i18n.T("en", "greeting", "name", "World")
	if result != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got %q", result)
	}

	result = i18n.T("fr", "greeting", "name", "Monde")
	if result != "Bonjour, Monde!" {
		t.Errorf("Expected 'Bonjour, Monde!', got %q", result)
	}
}

// TestLoadTranslationsFS verifies that translations are loaded from an http.FileSystem.
func TestLoadTranslationsFS(t *testing.T) {
	resetTranslations(t)
	fsMap := fstest.MapFS{
		"file1.yaml": &fstest.MapFile{
			Data: []byte(`
en:
  farewell: "Goodbye, %{name}!"
`),
		},
		"sub/file2.yaml": &fstest.MapFile{
			Data: []byte(`
fr:
  farewell: "Au revoir, %{name}!"
`),
		},
	}

	httpFS := http.FS(fsMap)
	if err := i18n.LoadTranslationsFS(httpFS, "."); err != nil {
		t.Fatalf("LoadTranslationsFS failed: %v", err)
	}

	result := i18n.T("en", "farewell", "name", "World")
	if result != "Goodbye, World!" {
		t.Errorf("Expected 'Goodbye, World!', got %q", result)
	}

	result = i18n.T("fr", "farewell", "name", "Monde")
	if result != "Au revoir, Monde!" {
		t.Errorf("Expected 'Au revoir, Monde!', got %q", result)
	}
}

// TestBestLangFromAcceptLanguage verifies that the best matching language is selected from an Accept-Language header.
func TestBestLangFromAcceptLanguage(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  dummy: "dummy"
fr:
  dummy: "dummy"
`
	filename := createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	header := "de, en-US;q=0.8, fr;q=0.9"
	best := i18n.BestLangFromAcceptLanguage(header, "fr")
	if best != "fr" {
		t.Errorf("Expected best lang 'fr', got %q", best)
	}

	resetTranslations(t)
	yamlContent = `
en:
  dummy: "dummy"
`
	filename = createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}
	header = "de, en-US;q=0.8, fr;q=0.9"
	best = i18n.BestLangFromAcceptLanguage(header, "en")
	if best != "en" {
		t.Errorf("Expected best lang 'en', got %q", best)
	}
}

// TestSprintfWithMultipleArgs verifies that standard formatting is ignored if no placeholder exists.
func TestSprintfWithMultipleArgs(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  combo: "Hello, %{name}! You have %{count} new messages."
`
	filename := createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	result := i18n.T("en", "combo", "count", "5", "name", "Alice")
	expected := "Hello, Alice! You have 5 new messages."
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestParseQValueErrors ensures that malformed q values in Accept-Language headers default to q=1.0.
func TestParseQValueErrors(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  dummy: "dummy"
fr:
  dummy: "dummy"
`
	filename := createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}

	header := "en;q=abc, fr;q=0.8"
	best := i18n.BestLangFromAcceptLanguage(header, "en")
	if best != "en" {
		t.Errorf("Expected best lang 'en' due to default q, got %q", best)
	}
}

// TestFallbackWhenNoSupportedLanguages verifies that an empty best language is returned when no translations are loaded.
func TestFallbackWhenNoSupportedLanguages(t *testing.T) {
	resetTranslations(t)
	header := "en, fr;q=0.9"
	best := i18n.BestLangFromAcceptLanguage(header)
	if best != "en" {
		t.Errorf("Expected en as best lang when no translations loaded, got %q", best)
	}
}

// TestLoadTranslationsFSRecursion verifies that nested directories within an FS are processed.
func TestLoadTranslationsFSRecursion(t *testing.T) {
	resetTranslations(t)
	fsMap := fstest.MapFS{
		"nested/level/file.yaml": &fstest.MapFile{
			Data: []byte(`
en:
  nested: "Nested %{value}"
`),
		},
	}
	httpFS := http.FS(fsMap)
	if err := i18n.LoadTranslationsFS(httpFS, "nested"); err != nil {
		t.Fatalf("LoadTranslationsFS (recursive) failed: %v", err)
	}

	result := i18n.T("en", "nested", "value", "test")
	if result != "Nested test" {
		t.Errorf("Expected 'Nested test', got %q", result)
	}
}

// TestLoadTranslationFileFromFS verifies that a single file (non-directory) is correctly loaded from an FS.
func TestLoadTranslationFileFromFS(t *testing.T) {
	resetTranslations(t)
	fsMap := fstest.MapFS{
		"only.yaml": &fstest.MapFile{
			Data: []byte(`
en:
  single: "Single %{value}"
`),
		},
	}
	httpFS := http.FS(fsMap)
	if err := i18n.LoadTranslationsFS(httpFS, "only.yaml"); err != nil {
		t.Fatalf("LoadTranslationsFS (file) failed: %v", err)
	}

	result := i18n.T("en", "single", "value", "value")
	if result != "Single value" {
		t.Errorf("Expected 'Single value', got %q", result)
	}
}

// BenchmarkLoadTranslationsDir benchmarks the performance of loading translations from a directory with many files.
func BenchmarkLoadTranslationsDir(b *testing.B) {
	dir, err := os.MkdirTemp("", "i18n_benchmark")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	numFiles := 100
	for i := 0; i < numFiles; i++ {
		filePath := filepath.Join(dir, "file"+strconv.Itoa(i)+".yaml")
		content := "en:\n  key" + strconv.Itoa(i) + ": \"Value " + strconv.Itoa(i) + "\"\n"
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := i18n.LoadTranslationsDir(dir); err != nil {
			b.Fatal(err)
		}
	}
}

// TestValidateTranslations verifies that validation catches invalid translation structures.
func TestValidateTranslations(t *testing.T) {
	// Original logger and settings
	origLogger := i18n.Logger
	origLogMissing := i18n.LogMissingTranslations
	origFallbackToKey := i18n.FallbackToKey
	
	// Restore original settings after test
	defer func() {
		i18n.Logger = origLogger
		i18n.LogMissingTranslations = origLogMissing
		i18n.FallbackToKey = origFallbackToKey
	}()
	
	// Configure package for testing
	i18n.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid translations",
			content: `en: {greeting: "Hello"}`,
			wantErr: false,
		},
		{
			name:    "empty translations",
			content: `{}`,
			wantErr: false, // This produces a warning but not an error
		},
		{
			name:    "empty language code",
			content: `"": {greeting: "Hello"}`,
			wantErr: true,
		},
		{
			name:    "nil entries",
			content: `en: null`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTranslations(t)
			filename := createTempYAML(t, tt.content)
			err := i18n.LoadTranslations(filename)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTranslations() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestErrorTypes verifies that the package returns the expected error types.
func TestErrorTypes(t *testing.T) {
	t.Run("file system error", func(t *testing.T) {
		err := i18n.LoadTranslations("nonexistent_file.yaml")
		if err == nil {
			t.Fatal("Expected error for nonexistent file, got nil")
		}
		if !strings.Contains(err.Error(), "file system error") {
			t.Errorf("Expected file system error, got: %v", err)
		}
	})
	
	t.Run("invalid format error", func(t *testing.T) {
		content := `invalid: yaml: [`
		filename := createTempYAML(t, content)
		err := i18n.LoadTranslations(filename)
		if err == nil {
			t.Fatal("Expected error for invalid YAML, got nil")
		}
		if !strings.Contains(err.Error(), "invalid translation format") {
			t.Errorf("Expected invalid format error, got: %v", err)
		}
	})
}

// TestFallbackBehavior verifies the configurable fallback behavior.
func TestFallbackBehavior(t *testing.T) {
	resetTranslations(t)
	yamlContent := `
en:
  welcome: "Welcome"
`
	filename := createTempYAML(t, yamlContent)
	if err := i18n.LoadTranslations(filename); err != nil {
		t.Fatalf("LoadTranslations failed: %v", err)
	}
	
	// Original settings
	origFallbackToKey := i18n.FallbackToKey
	origLogMissing := i18n.LogMissingTranslations
	
	// Restore settings after test
	defer func() {
		i18n.FallbackToKey = origFallbackToKey
		i18n.LogMissingTranslations = origLogMissing
	}()
	
	// Test with fallback enabled (default)
	i18n.FallbackToKey = true
	result := i18n.T("en", "missing_key")
	if result != "missing_key" {
		t.Errorf("Expected fallback to key with FallbackToKey=true, got %q", result)
	}
	
	// Test with fallback disabled
	i18n.FallbackToKey = false
	result = i18n.T("en", "missing_key")
	if result != "" {
		t.Errorf("Expected empty string with FallbackToKey=false, got %q", result)
	}
	
	// Test with unsupported language
	result = i18n.T("fr", "welcome")
	if i18n.FallbackToKey && result != "welcome" {
		t.Errorf("Expected fallback to key for unsupported language, got %q", result)
	} else if !i18n.FallbackToKey && result != "" {
		t.Errorf("Expected empty string for unsupported language with FallbackToKey=false, got %q", result)
	}
}
