package privacy_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhoneMasker_Mask(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Test with different phone number formats
		result, err := masker.Mask(ctx, "555-123-4567")
		require.NoError(t, err)
		assert.Equal(t, "***-***-4567", result)

		result, err = masker.Mask(ctx, "(555) 123-4567")
		require.NoError(t, err)
		assert.Equal(t, "(***) ***-4567", result)

		result, err = masker.Mask(ctx, "5551234567")
		require.NoError(t, err)
		assert.Equal(t, "******4567", result)
	})

	t.Run("CustomReplacement", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker(
			privacy.WithPhoneReplacement('#'),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "555-123-4567")
		require.NoError(t, err)
		assert.Equal(t, "###-###-4567", result)
	})

	t.Run("CustomVisibleDigits", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker(
			privacy.WithVisiblePhoneDigits(2),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "555-123-4567")
		require.NoError(t, err)
		assert.Equal(t, "***-***-**67", result)
	})

	t.Run("NoPreserveFormat", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker(
			privacy.WithPreserveFormat(false),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "555-123-4567")
		require.NoError(t, err)
		assert.Equal(t, "******4567", result)
	})

	t.Run("InternationalNumber", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "+1 555-123-4567")
		require.NoError(t, err)
		assert.Equal(t, "+1 ***-***-4567", result)
	})

	t.Run("NoPreserveCountryCode", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker(
			privacy.WithPreserveCountryCode(false),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "+1 555-123-4567")
		require.NoError(t, err)
		assert.Equal(t, "** ***-***-4567", result)
	})

	t.Run("CanMask", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Valid phone numbers should return true
		assert.True(t, masker.CanMask("555-123-4567"))
		assert.True(t, masker.CanMask("(555) 123-4567"))
		assert.True(t, masker.CanMask("5551234567"))
		assert.True(t, masker.CanMask("+1 555-123-4567"))

		// Invalid phone numbers or non-string types should return false
		assert.False(t, masker.CanMask("123"))  // Too short
		assert.False(t, masker.CanMask("abcdefghijklm"))  // Not a phone number
		assert.False(t, masker.CanMask(123))
		assert.False(t, masker.CanMask([]string{"not", "a", "phone"}))
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		_, err = masker.Mask(ctx, 5551234567)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrUnsupportedType)
	})

	t.Run("CanceledContext", func(t *testing.T) {
		masker, err := privacy.NewPhoneMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Create canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err = masker.Mask(canceledCtx, "555-123-4567")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		// Test negative visible digit count
		_, err := privacy.NewPhoneMasker(
			privacy.WithVisiblePhoneDigits(-1),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)
	})
}
