package privacy_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCardMasker_Mask(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		masker, err := privacy.NewCardMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Test with different card formats
		result, err := masker.Mask(ctx, "4111 1111 1111 1111")
		require.NoError(t, err)
		assert.Equal(t, "**** **** **** 1111", result)

		result, err = masker.Mask(ctx, "4111111111111111")
		require.NoError(t, err)
		assert.Equal(t, "************1111", result)

		result, err = masker.Mask(ctx, "4111-1111-1111-1111")
		require.NoError(t, err)
		assert.Equal(t, "**** **** **** 1111", result)
	})

	t.Run("CustomReplacement", func(t *testing.T) {
		masker, err := privacy.NewCardMasker(
			privacy.WithCardReplacement('#'),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "4111 1111 1111 1111")
		require.NoError(t, err)
		assert.Equal(t, "#### #### #### 1111", result)
	})

	t.Run("CustomVisibleDigits", func(t *testing.T) {
		masker, err := privacy.NewCardMasker(
			privacy.WithVisibleEndDigits(2),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "4111 1111 1111 1111")
		require.NoError(t, err)
		assert.Equal(t, "**** **** **** **11", result)
	})

	t.Run("NoFormatting", func(t *testing.T) {
		masker, err := privacy.NewCardMasker(
			privacy.WithCardFormatting(false),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "4111 1111 1111 1111")
		require.NoError(t, err)
		assert.Equal(t, "************1111", result)
	})

	t.Run("AmexFormat", func(t *testing.T) {
		masker, err := privacy.NewCardMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "3782 822463 10005")
		require.NoError(t, err)
		assert.Equal(t, "**** ***** *0005", result)
	})

	t.Run("InvalidCardNumber", func(t *testing.T) {
		masker, err := privacy.NewCardMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Card number too short
		_, err = masker.Mask(ctx, "1234")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidData)

		// Card number too long
		_, err = masker.Mask(ctx, "1234567890123456789012345")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidData)
	})

	t.Run("CanMask", func(t *testing.T) {
		masker, err := privacy.NewCardMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Valid card numbers should return true
		assert.True(t, masker.CanMask("4111 1111 1111 1111"))
		assert.True(t, masker.CanMask("4111111111111111"))
		assert.True(t, masker.CanMask("3782 822463 10005"))
		assert.True(t, masker.CanMask("5555-5555-5555-4444"))

		// Invalid card numbers or non-string types should return false
		assert.False(t, masker.CanMask("1234"))
		assert.False(t, masker.CanMask("abcd efgh ijkl mnop"))
		assert.False(t, masker.CanMask(123))
		assert.False(t, masker.CanMask([]string{"not", "a", "card"}))
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		masker, err := privacy.NewCardMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		_, err = masker.Mask(ctx, 4111111111111111)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrUnsupportedType)
	})

	t.Run("CanceledContext", func(t *testing.T) {
		masker, err := privacy.NewCardMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Create canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err = masker.Mask(canceledCtx, "4111 1111 1111 1111")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		// Test negative visible digit count
		_, err := privacy.NewCardMasker(
			privacy.WithVisibleEndDigits(-1),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)

		// Test too many visible digits (security risk)
		_, err = privacy.NewCardMasker(
			privacy.WithVisibleEndDigits(8),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)
	})
}
