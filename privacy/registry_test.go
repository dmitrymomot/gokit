package privacy_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/privacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockMasker is a simple masker for testing
type MockMasker struct {
	mockMask    func(ctx context.Context, data any) (any, error)
	mockCanMask func(data any) bool
}

func (m *MockMasker) Mask(ctx context.Context, data any) (any, error) {
	return m.mockMask(ctx, data)
}

func (m *MockMasker) CanMask(data any) bool {
	return m.mockCanMask(data)
}

func NewMockMasker(mockMask func(ctx context.Context, data any) (any, error), mockCanMask func(data any) bool) *MockMasker {
	return &MockMasker{
		mockMask:    mockMask,
		mockCanMask: mockCanMask,
	}
}

func TestMaskingRegistry(t *testing.T) {
	ctx := context.Background()

	t.Run("RegisterAndGetMasker", func(t *testing.T) {
		// Create a registry
		registry := privacy.NewMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Create a mock masker
		mockMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "MASKED", nil },
			func(data any) bool { return true },
		)

		// Register the masker
		err := registry.RegisterMasker(privacy.CategoryCreditCard, mockMasker)
		require.NoError(t, err)

		// Get the masker
		masker, err := registry.GetMasker(privacy.CategoryCreditCard)
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Use the masker
		result, err := masker.Mask(ctx, "4111 1111 1111 1111")
		require.NoError(t, err)
		assert.Equal(t, "MASKED", result)
	})

	t.Run("MaskByCategory", func(t *testing.T) {
		// Create a registry
		registry := privacy.NewMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Create a mock masker
		mockMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "MASKED", nil },
			func(data any) bool { return true },
		)

		// Register the masker
		err := registry.RegisterMasker(privacy.CategoryEmail, mockMasker)
		require.NoError(t, err)

		// Mask data by category
		result, err := registry.MaskByCategory(ctx, privacy.CategoryEmail, "john.doe@example.com")
		require.NoError(t, err)
		assert.Equal(t, "MASKED", result)
	})

	t.Run("UnregisterMasker", func(t *testing.T) {
		// Create a registry
		registry := privacy.NewMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Create a mock masker
		mockMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "MASKED", nil },
			func(data any) bool { return true },
		)

		// Register the masker
		err := registry.RegisterMasker(privacy.CategoryPhone, mockMasker)
		require.NoError(t, err)

		// Unregister the masker
		err = registry.UnregisterMasker(privacy.CategoryPhone)
		require.NoError(t, err)

		// Try to get the masker
		_, err = registry.GetMasker(privacy.CategoryPhone)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrMaskerNotFound)
	})

	t.Run("DefaultMasker", func(t *testing.T) {
		// Create a default masker
		defaultMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "DEFAULT_MASKED", nil },
			func(data any) bool { return true },
		)

		// Create a registry with default masker
		registry := privacy.NewMaskingRegistry(defaultMasker)
		require.NotNil(t, registry)

		// Get a masker for a category that doesn't exist
		masker, err := registry.GetMasker(privacy.CategorySSN)
		require.NoError(t, err)
		require.NotNil(t, masker)

		// Use the default masker
		result, err := masker.Mask(ctx, "123-45-6789")
		require.NoError(t, err)
		assert.Equal(t, "DEFAULT_MASKED", result)
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Create a registry
		registry := privacy.NewMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Try to register a nil masker
		err := registry.RegisterMasker(privacy.CategoryName, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidMasker)

		// Try to unregister a non-existent masker
		err = registry.UnregisterMasker(privacy.CategoryAddress)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrMaskerNotFound)

		// Try to get a non-existent masker
		_, err = registry.GetMasker(privacy.CategoryPassport)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrMaskerNotFound)

		// Try to mask without a masker
		_, err = registry.MaskByCategory(ctx, privacy.CategoryDriverLicense, "DL12345678")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrMaskerNotFound)
	})

	t.Run("MaskerTypeCheck", func(t *testing.T) {
		// Create a registry
		registry := privacy.NewMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Create a masker that can't handle integers
		mockMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "MASKED", nil },
			func(data any) bool { _, ok := data.(string); return ok },
		)

		// Register the masker
		err := registry.RegisterMasker(privacy.CategoryName, mockMasker)
		require.NoError(t, err)

		// Try to mask with an incompatible type
		_, err = registry.MaskByCategory(ctx, privacy.CategoryName, 12345)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrUnsupportedType)
	})
}

func TestAutoMaskingRegistry(t *testing.T) {
	ctx := context.Background()

	t.Run("DetectAndMask", func(t *testing.T) {
		// Create an auto-masking registry
		registry := privacy.NewAutoMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Create maskers for different categories
		emailMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "MASKED_EMAIL", nil },
			func(data any) bool { return true },
		)

		cardMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "MASKED_CARD", nil },
			func(data any) bool { return true },
		)

		// Register maskers
		err := registry.RegisterMasker(privacy.CategoryEmail, emailMasker)
		require.NoError(t, err)
		err = registry.RegisterMasker(privacy.CategoryCreditCard, cardMasker)
		require.NoError(t, err)

		// Register detection rules
		err = registry.RegisterDetectionRule(privacy.CategoryEmail, func(data any) bool {
			str, ok := data.(string)
			return ok && (str == "john.doe@example.com" || str == "test@example.com")
		})
		require.NoError(t, err)

		err = registry.RegisterDetectionRule(privacy.CategoryCreditCard, func(data any) bool {
			str, ok := data.(string)
			return ok && (str == "4111 1111 1111 1111" || str == "5555555555554444")
		})
		require.NoError(t, err)

		// Auto-detect and mask email
		result, err := registry.AutoMask(ctx, "john.doe@example.com")
		require.NoError(t, err)
		assert.Equal(t, "MASKED_EMAIL", result)

		// Auto-detect and mask card
		result, err = registry.AutoMask(ctx, "4111 1111 1111 1111")
		require.NoError(t, err)
		assert.Equal(t, "MASKED_CARD", result)
	})

	t.Run("UnregisterDetectionRule", func(t *testing.T) {
		// Create an auto-masking registry
		registry := privacy.NewAutoMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Register a detection rule
		err := registry.RegisterDetectionRule(privacy.CategoryEmail, func(data any) bool {
			str, ok := data.(string)
			return ok && str == "john.doe@example.com"
		})
		require.NoError(t, err)

		// Unregister the rule
		err = registry.UnregisterDetectionRule(privacy.CategoryEmail)
		require.NoError(t, err)

		// Try to detect the category
		_, detected := registry.DetectCategory("john.doe@example.com")
		assert.False(t, detected)
	})

	t.Run("DefaultMaskerForAutoMask", func(t *testing.T) {
		// Create a default masker
		defaultMasker := NewMockMasker(
			func(ctx context.Context, data any) (any, error) { return "DEFAULT_MASKED", nil },
			func(data any) bool { return true },
		)

		// Create an auto-masking registry with default masker
		registry := privacy.NewAutoMaskingRegistry(defaultMasker)
		require.NotNil(t, registry)

		// Auto-mask data without a detection rule
		result, err := registry.AutoMask(ctx, "some data without a detection rule")
		require.NoError(t, err)
		assert.Equal(t, "DEFAULT_MASKED", result)
	})

	t.Run("ErrorCases", func(t *testing.T) {
		// Create an auto-masking registry
		registry := privacy.NewAutoMaskingRegistry(nil)
		require.NotNil(t, registry)

		// Try to register a nil rule
		err := registry.RegisterDetectionRule(privacy.CategoryEmail, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrInvalidRule)

		// Try to unregister a non-existent rule
		err = registry.UnregisterDetectionRule(privacy.CategoryCreditCard)
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrRuleNotFound)

		// Try to auto-mask without detection rules and default masker
		_, err = registry.AutoMask(ctx, "some data")
		assert.Error(t, err)
		assert.ErrorIs(t, err, privacy.ErrCategoryNotDetected)
	})
}
