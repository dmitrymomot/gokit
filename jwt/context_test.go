package jwt_test

import (
	"context"
	"testing"

	"github.com/dmitrymomot/gokit/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	token := "test.jwt.token"

	// Act
	newCtx := jwt.SetToken(ctx, token)

	// Assert
	require.NotNil(t, newCtx, "Context should not be nil")
	assert.NotEqual(t, ctx, newCtx, "New context should be different from original")

	// Verify token can be retrieved
	retrievedToken, ok := jwt.GetToken(newCtx)
	assert.True(t, ok, "Should be able to retrieve token")
	assert.Equal(t, token, retrievedToken, "Retrieved token should match original")
}

func TestGetToken(t *testing.T) {
	// Test successful retrieval
	t.Run("TokenExists", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		token := "test.jwt.token"
		ctx = jwt.SetToken(ctx, token)

		// Act
		retrievedToken, ok := jwt.GetToken(ctx)

		// Assert
		assert.True(t, ok, "Should return true when token exists")
		assert.Equal(t, token, retrievedToken, "Retrieved token should match original")
	})

	// Test token not found
	t.Run("TokenNotFound", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		retrievedToken, ok := jwt.GetToken(ctx)

		// Assert
		assert.False(t, ok, "Should return false when token doesn't exist")
		assert.Empty(t, retrievedToken, "Retrieved token should be empty")
	})
}

func TestSetClaims(t *testing.T) {
	// Arrange
	ctx := context.Background()
	claims := map[string]any{
		"sub":   "1234567890",
		"name":  "John Doe",
		"admin": true,
	}

	// Act
	newCtx := jwt.SetClaims(ctx, claims)

	// Assert
	require.NotNil(t, newCtx, "Context should not be nil")
	assert.NotEqual(t, ctx, newCtx, "New context should be different from original")

	// Verify claims can be retrieved
	retrievedClaims, ok := jwt.GetClaims(newCtx)
	assert.True(t, ok, "Should be able to retrieve claims")
	assert.Equal(t, claims, retrievedClaims, "Retrieved claims should match original")
}

func TestGetClaims(t *testing.T) {
	// Test successful retrieval
	t.Run("ClaimsExist", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		claims := map[string]any{
			"sub":   "1234567890",
			"name":  "John Doe",
			"admin": true,
		}
		ctx = jwt.SetClaims(ctx, claims)

		// Act
		retrievedClaims, ok := jwt.GetClaims(ctx)

		// Assert
		assert.True(t, ok, "Should return true when claims exist")
		assert.Equal(t, claims, retrievedClaims, "Retrieved claims should match original")
	})

	// Test claims not found
	t.Run("ClaimsNotFound", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		retrievedClaims, ok := jwt.GetClaims(ctx)

		// Assert
		assert.False(t, ok, "Should return false when claims don't exist")
		assert.Nil(t, retrievedClaims, "Retrieved claims should be nil")
	})
}

// Define a test struct for claims
type CtxTestClaims struct {
	Sub   string `json:"sub"`
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
}

func TestGetClaimsAs(t *testing.T) {
	// Test successful parsing
	t.Run("SuccessfulParsing", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		claims := map[string]any{
			"sub":   "1234567890",
			"name":  "John Doe",
			"admin": true,
		}
		ctx = jwt.SetClaims(ctx, claims)
		var testClaims CtxTestClaims

		// Act
		err := jwt.GetClaimsAs(ctx, &testClaims)

		// Assert
		require.NoError(t, err, "Should parse claims without error")
		assert.Equal(t, "1234567890", testClaims.Sub, "Sub claim should match")
		assert.Equal(t, "John Doe", testClaims.Name, "Name claim should match")
		assert.True(t, testClaims.Admin, "Admin claim should match")
	})

	// Test claims not found
	t.Run("ClaimsNotFound", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		var testClaims CtxTestClaims

		// Act
		err := jwt.GetClaimsAs(ctx, &testClaims)

		// Assert
		require.Error(t, err, "Should return error when claims don't exist")
		assert.ErrorIs(t, err, jwt.ErrInvalidClaims, "Error should be ErrInvalidClaims")
	})

	// Test invalid claims format
	t.Run("InvalidClaimsFormat", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		// Create claims with incompatible types
		claims := map[string]any{
			"sub":   123,   // Should be string but is int
			"name":  true,  // Should be string but is bool
			"admin": "yes", // Should be bool but is string
		}
		ctx = jwt.SetClaims(ctx, claims)
		var testClaims CtxTestClaims

		// Act
		err := jwt.GetClaimsAs(ctx, &testClaims)

		// Assert
		require.Error(t, err, "Should return error when claims format is invalid")
		assert.Contains(t, err.Error(), "failed to unmarshal claims", "Error should mention unmarshal failure")
	})

	// Test nil claims pointer
	t.Run("NilClaimsPointer", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		claims := map[string]any{
			"sub":   "1234567890",
			"name":  "John Doe",
			"admin": true,
		}
		ctx = jwt.SetClaims(ctx, claims)

		// Act
		err := jwt.GetClaimsAs[CtxTestClaims](ctx, nil)

		// Assert
		require.Error(t, err, "Should return error when claims pointer is nil")
		assert.Contains(t, err.Error(), "failed to unmarshal claims", "Error should mention unmarshal failure")
	})
}

// Test both token and claims in the same context
func TestTokenAndClaimsTogether(t *testing.T) {
	// Arrange
	ctx := context.Background()
	token := "test.jwt.token"
	claims := map[string]any{
		"sub":   "1234567890",
		"name":  "John Doe",
		"admin": true,
	}

	// Act
	ctx = jwt.SetToken(ctx, token)
	ctx = jwt.SetClaims(ctx, claims)

	// Assert
	retrievedToken, tokenOk := jwt.GetToken(ctx)
	retrievedClaims, claimsOk := jwt.GetClaims(ctx)

	assert.True(t, tokenOk, "Should be able to retrieve token")
	assert.Equal(t, token, retrievedToken, "Retrieved token should match original")

	assert.True(t, claimsOk, "Should be able to retrieve claims")
	assert.Equal(t, claims, retrievedClaims, "Retrieved claims should match original")
}
