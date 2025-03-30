package privacy_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringMasker_Mask(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultRedactStrategy", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyRedact),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "sensitive data")
		require.NoError(t, err)
		assert.Equal(t, "[REDACTED]", result)
	})

	t.Run("PartialMaskStrategy", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithReplacement('*'),
			privacy.WithVisibleChars(2, 2),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "HelloWorld")
		require.NoError(t, err)
		assert.Equal(t, "He******ld", result)
	})

	t.Run("CustomReplacement", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithReplacement('#'),
			privacy.WithVisibleChars(0, 3),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "HelloWorld")
		require.NoError(t, err)
		assert.Equal(t, "#######rld", result)
	})

	t.Run("MinLengthHandling", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(1, 1),
			privacy.WithMinLength(5),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		// String shorter than min length should be returned unchanged
		result, err := masker.Mask(ctx, "Test")
		require.NoError(t, err)
		assert.Equal(t, "Test", result)

		// String equal or longer than min length should be masked
		result, err = masker.Mask(ctx, "TestMe")
		require.NoError(t, err)
		assert.Equal(t, "T****e", result)
	})

	t.Run("InvalidStrategy", func(t *testing.T) {
		// Create a masker with an invalid strategy (using a string that isn't a valid MaskingStrategy)
		invalidStrategy := privacy.MaskingStrategy("invalid_strategy")
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(invalidStrategy),
		)
		require.NoError(t, err)
		
		// Try to mask using the invalid strategy
		_, err = masker.Mask(ctx, "Hello")
		assert.Error(t, err)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(1, 1),
		)
		require.NoError(t, err)

		_, err = masker.Mask(ctx, 123) // Integer is not supported
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrUnsupportedType)
	})

	t.Run("VisibleCharactersExceedLength", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(3, 3),
		)
		require.NoError(t, err)

		result, err := masker.Mask(ctx, "Hello")
		require.NoError(t, err)
		assert.Equal(t, "Hello", result)
	})

	t.Run("CanceledContext", func(t *testing.T) {
		masker, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(1, 1),
		)
		require.NoError(t, err)

		// Create canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err = masker.Mask(canceledCtx, "Hello")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		// Negative visible characters
		_, err := privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(-1, 1),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)

		// Partial strategy with no visible chars
		_, err = privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(0, 0),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)

		// Negative min length
		_, err = privacy.NewStringMasker(
			privacy.WithStrategy(privacy.StrategyPartialMask),
			privacy.WithVisibleChars(1, 1),
			privacy.WithMinLength(-1),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)
	})
}
