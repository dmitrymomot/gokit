package privacy_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailMasker_Mask(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "john.doe@example.com")
		require.NoError(t, err)
		assert.Equal(t, "j******@example.com", result)

		// Test with a different email structure
		result, err = masker.Mask(ctx, "short@test.io")
		require.NoError(t, err)
		assert.Equal(t, "s****@test.io", result)
	})

	t.Run("CustomConfiguration", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker(
			privacy.WithEmailReplacement('#'),
			privacy.WithVisibleLocalChars(2),
			privacy.WithVisibleDomainChars(1),
			privacy.WithShowDomainExt(true),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "john.doe@example.com")
		require.NoError(t, err)
		assert.Equal(t, "jo######@e######.com", result)
	})

	t.Run("HideDomainExtension", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker(
			privacy.WithShowDomainExt(false),
		)
		require.NoError(t, err)
		require.NotNil(t, masker)

		result, err := masker.Mask(ctx, "john.doe@example.com")
		require.NoError(t, err)
		assert.Equal(t, "j******@**********", result)
	})

	t.Run("InvalidEmailFormat", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Test with invalid email that has no domain part
		_, err = masker.Mask(ctx, "john.doe")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidData)

		// Test with invalid email that has empty parts
		_, err = masker.Mask(ctx, "@example.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidData)

		// Test with invalid email that has empty domain
		_, err = masker.Mask(ctx, "john.doe@")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidData)
	})

	t.Run("CanMask", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Valid emails should return true
		assert.True(t, masker.CanMask("john.doe@example.com"))
		assert.True(t, masker.CanMask("short@test.io"))

		// Invalid emails or non-string types should return false
		assert.False(t, masker.CanMask("john.doe"))
		assert.False(t, masker.CanMask("@example.com"))
		assert.False(t, masker.CanMask("john.doe@"))
		assert.False(t, masker.CanMask(123))
		assert.False(t, masker.CanMask([]string{"not", "an", "email"}))
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		_, err = masker.Mask(ctx, 123)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrUnsupportedType)
	})

	t.Run("CanceledContext", func(t *testing.T) {
		masker, err := privacy.NewEmailMasker()
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Create canceled context
		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err = masker.Mask(canceledCtx, "john.doe@example.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		// Test negative visible character count
		_, err := privacy.NewEmailMasker(
			privacy.WithVisibleLocalChars(-1),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)

		_, err = privacy.NewEmailMasker(
			privacy.WithVisibleDomainChars(-1),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMask)
	})
}
