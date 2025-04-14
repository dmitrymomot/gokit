package i18n_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser(t *testing.T) {
	parser := i18n.NewJSONParser()

	t.Run("Parse valid JSON", func(t *testing.T) {
		content := `{
			"en": {
				"greeting": "Hello",
				"farewell": "Goodbye",
				"nested": {
					"key": "Nested value"
				}
			},
			"fr": {
				"greeting": "Bonjour",
				"farewell": "Au revoir"
			}
		}`

		result, err := parser.Parse(context.Background(), content)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check English translations
		assert.Contains(t, result, "en")
		assert.Equal(t, "Hello", result["en"]["greeting"])
		assert.Equal(t, "Goodbye", result["en"]["farewell"])

		// Check nested values
		nestedMap, ok := result["en"]["nested"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Nested value", nestedMap["key"])

		// Check French translations
		assert.Contains(t, result, "fr")
		assert.Equal(t, "Bonjour", result["fr"]["greeting"])
		assert.Equal(t, "Au revoir", result["fr"]["farewell"])
	})

	t.Run("Parse invalid JSON", func(t *testing.T) {
		content := `{
			"en": {
				"greeting": "Hello",
				"farewell": "Goodbye",
			}
		}`

		result, err := parser.Parse(context.Background(), content)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to parse JSON content")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		content := `{
			"en": {
				"greeting": "Hello"
			}
		}`

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		result, err := parser.Parse(ctx, content)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "json parsing cancelled")
	})

	t.Run("SupportsFileExtension", func(t *testing.T) {
		assert.True(t, parser.SupportsFileExtension("json"))
		assert.True(t, parser.SupportsFileExtension(".json"))
		assert.True(t, parser.SupportsFileExtension("JSON"))
		assert.False(t, parser.SupportsFileExtension("yaml"))
		assert.False(t, parser.SupportsFileExtension("yml"))
	})
}

func TestYAMLParser(t *testing.T) {
	parser := i18n.NewYAMLParser()

	t.Run("Parse valid YAML", func(t *testing.T) {
		content := `
en:
  greeting: Hello
  farewell: Goodbye
  nested:
    key: Nested value
fr:
  greeting: Bonjour
  farewell: Au revoir
`

		result, err := parser.Parse(context.Background(), content)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Check English translations
		assert.Contains(t, result, "en")
		assert.Equal(t, "Hello", result["en"]["greeting"])
		assert.Equal(t, "Goodbye", result["en"]["farewell"])

		// Check nested values
		nestedMap, ok := result["en"]["nested"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Nested value", nestedMap["key"])

		// Check French translations
		assert.Contains(t, result, "fr")
		assert.Equal(t, "Bonjour", result["fr"]["greeting"])
		assert.Equal(t, "Au revoir", result["fr"]["farewell"])
	})

	t.Run("Parse invalid YAML", func(t *testing.T) {
		content := `
en:
  - greeting: Hello  # Invalid structure (array instead of map)
  - farewell: Goodbye
fr:
  greeting: Bonjour
`

		result, err := parser.Parse(context.Background(), content)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid YAML structure for language")
	})

	t.Run("Empty YAML", func(t *testing.T) {
		content := ``

		result, err := parser.Parse(context.Background(), content)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no valid translations found")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		content := `
en:
  greeting: Hello
`

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		result, err := parser.Parse(ctx, content)
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "yaml parsing cancelled")
	})

	t.Run("SupportsFileExtension", func(t *testing.T) {
		assert.True(t, parser.SupportsFileExtension("yaml"))
		assert.True(t, parser.SupportsFileExtension(".yaml"))
		assert.True(t, parser.SupportsFileExtension("YAML"))
		assert.True(t, parser.SupportsFileExtension("yml"))
		assert.True(t, parser.SupportsFileExtension(".yml"))
		assert.False(t, parser.SupportsFileExtension("json"))
	})
}

func TestParserFactory(t *testing.T) {
	t.Run("JSON file extension", func(t *testing.T) {
		parser := i18n.NewParserForFile("translations.json")
		require.NotNil(t, parser)
		_, ok := parser.(*i18n.JSONParser)
		assert.True(t, ok, "Should return a JSONParser for .json files")
	})

	t.Run("YAML file extensions", func(t *testing.T) {
		// Test .yaml extension
		parser := i18n.NewParserForFile("translations.yaml")
		require.NotNil(t, parser)
		_, ok := parser.(*i18n.YAMLParser)
		assert.True(t, ok, "Should return a YAMLParser for .yaml files")

		// Test .yml extension
		parser = i18n.NewParserForFile("translations.yml")
		require.NotNil(t, parser)
		_, ok = parser.(*i18n.YAMLParser)
		assert.True(t, ok, "Should return a YAMLParser for .yml files")
	})

	t.Run("Uppercase extensions", func(t *testing.T) {
		// Test uppercase JSON
		parser := i18n.NewParserForFile("translations.JSON")
		require.NotNil(t, parser)
		_, ok := parser.(*i18n.JSONParser)
		assert.True(t, ok, "Should handle uppercase extensions")

		// Test uppercase YAML
		parser = i18n.NewParserForFile("translations.YAML")
		require.NotNil(t, parser)
		_, ok = parser.(*i18n.YAMLParser)
		assert.True(t, ok, "Should handle uppercase extensions")
	})

	t.Run("Unsupported extension", func(t *testing.T) {
		parser := i18n.NewParserForFile("translations.txt")
		assert.Nil(t, parser, "Should return nil for unsupported extensions")
	})

	t.Run("No extension", func(t *testing.T) {
		parser := i18n.NewParserForFile("translations")
		assert.Nil(t, parser, "Should return nil for files without extensions")
	})
}