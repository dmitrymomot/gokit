package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/dmitrymomot/gokit/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatJSON(t *testing.T) {
	t.Run("format simple object", func(t *testing.T) {
		type testObject struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		obj := testObject{Name: "John", Age: 30}

		formatted := utils.FormatJSON(obj)

		expected, err := json.MarshalIndent(obj, "", "  ")
		require.NoError(t, err)
		assert.Equal(t, string(expected), formatted)
	})

	t.Run("format multiple objects", func(t *testing.T) {
		obj1 := map[string]string{"key": "value"}
		obj2 := []int{1, 2, 3}

		formatted := utils.FormatJSON(obj1, obj2)

		expected1, err := json.MarshalIndent(obj1, "", "  ")
		require.NoError(t, err)
		expected2, err := json.MarshalIndent(obj2, "", "  ")
		require.NoError(t, err)
		expected := string(expected1) + "\n" + string(expected2)
		assert.Equal(t, expected, formatted)
	})

	t.Run("handle non-marshalable object", func(t *testing.T) {
		ch := make(chan int) // channels can't be marshaled to JSON

		formatted := utils.FormatJSON(ch)

		// Instead of checking for an exact string, just verify it's not empty
		// and doesn't cause a panic
		assert.NotEmpty(t, formatted)
		// The fmt.Sprintf formatting will represent the channel somehow
		// but the exact representation may vary by Go version
	})
}

func TestDeprecatedPrettyPrint(t *testing.T) {
	t.Run("calls FormatJSON", func(t *testing.T) {
		obj := map[string]string{"key": "value"}

		prettyPrinted := utils.PrettyPrint(obj)
		formatted := utils.FormatJSON(obj)

		assert.Equal(t, formatted, prettyPrinted)
	})
}
