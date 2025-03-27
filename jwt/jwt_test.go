package jwt_test

import (
	"context"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Custom claims type for testing
type TestClaims struct {
	jwt.StandardClaims
	Name  string `json:"name,omitempty"`
	Admin bool   `json:"admin,omitempty"`
}

func TestNew(t *testing.T) {
	t.Run("with valid signing key", func(t *testing.T) {
		service, err := jwt.New([]byte("secret"))
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("with empty signing key", func(t *testing.T) {
		service, err := jwt.New([]byte{})
		require.Error(t, err)
		require.Equal(t, jwt.ErrMissingSigningKey, err)
		require.Nil(t, service)
	})
}

func TestNewFromString(t *testing.T) {
	t.Run("with valid signing key", func(t *testing.T) {
		service, err := jwt.NewFromString("secret")
		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("with empty signing key", func(t *testing.T) {
		service, err := jwt.NewFromString("")
		require.Error(t, err)
		require.Equal(t, jwt.ErrMissingSigningKey, err)
		require.Nil(t, service)
	})
}

func TestGenerate(t *testing.T) {
	service, err := jwt.New([]byte("secret"))
	require.NoError(t, err)
	require.NotNil(t, service)

	ctx := context.Background()

	t.Run("with standard claims", func(t *testing.T) {
		claims := jwt.StandardClaims{
			Subject:   "user123",
			Issuer:    "gokit",
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		token, err := service.Generate(ctx, claims)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Token should have 3 parts separated by dots
		parts := len(token)
		assert.True(t, parts > 0)
	})

	t.Run("with custom claims", func(t *testing.T) {
		claims := TestClaims{
			StandardClaims: jwt.StandardClaims{
				Subject:   "user123",
				Issuer:    "gokit",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
			Name:  "John Doe",
			Admin: true,
		}

		token, err := service.Generate(ctx, claims)
		require.NoError(t, err)
		require.NotEmpty(t, token)
	})

	t.Run("with nil claims", func(t *testing.T) {
		token, err := service.Generate(ctx, nil)
		require.Error(t, err)
		require.Equal(t, jwt.ErrMissingClaims, err)
		require.Empty(t, token)
	})
}

func TestParse(t *testing.T) {
	service, err := jwt.New([]byte("secret"))
	require.NoError(t, err)
	require.NotNil(t, service)

	ctx := context.Background()

	t.Run("with standard claims", func(t *testing.T) {
		// Generate a token
		originalClaims := jwt.StandardClaims{
			Subject:   "user123",
			Issuer:    "gokit",
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		token, err := service.Generate(ctx, originalClaims)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Parse the token
		var parsedClaims jwt.StandardClaims
		err = service.Parse(ctx, token, &parsedClaims)
		require.NoError(t, err)

		// Verify the claims
		assert.Equal(t, originalClaims.Subject, parsedClaims.Subject)
		assert.Equal(t, originalClaims.Issuer, parsedClaims.Issuer)
		assert.Equal(t, originalClaims.ExpiresAt, parsedClaims.ExpiresAt)
	})

	t.Run("with custom claims", func(t *testing.T) {
		// Generate a token
		originalClaims := TestClaims{
			StandardClaims: jwt.StandardClaims{
				Subject:   "user123",
				Issuer:    "gokit",
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
			},
			Name:  "John Doe",
			Admin: true,
		}

		token, err := service.Generate(ctx, originalClaims)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Parse the token
		var parsedClaims TestClaims
		err = service.Parse(ctx, token, &parsedClaims)
		require.NoError(t, err)

		// Verify the claims
		assert.Equal(t, originalClaims.Subject, parsedClaims.Subject)
		assert.Equal(t, originalClaims.Issuer, parsedClaims.Issuer)
		assert.Equal(t, originalClaims.ExpiresAt, parsedClaims.ExpiresAt)
		assert.Equal(t, originalClaims.Name, parsedClaims.Name)
		assert.Equal(t, originalClaims.Admin, parsedClaims.Admin)
	})

	t.Run("with invalid token format", func(t *testing.T) {
		var claims jwt.StandardClaims
		err := service.Parse(ctx, "invalid-token", &claims)
		require.Error(t, err)
		require.Equal(t, jwt.ErrInvalidToken, err)
	})

	t.Run("with invalid signature", func(t *testing.T) {
		// Generate a token
		originalClaims := jwt.StandardClaims{
			Subject:   "user123",
			Issuer:    "gokit",
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		}

		token, err := service.Generate(ctx, originalClaims)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Tamper with the signature by changing the last character
		parts := token[:len(token)-1] + "X"

		// Parse the token
		var parsedClaims jwt.StandardClaims
		err = service.Parse(ctx, parts, &parsedClaims)
		require.Error(t, err)
	})

	t.Run("with expired token", func(t *testing.T) {
		// Generate a token that's already expired
		expiredClaims := jwt.StandardClaims{
			Subject:   "user123",
			Issuer:    "gokit",
			ExpiresAt: time.Now().Add(-time.Hour).Unix(), // Expired
		}

		token, err := service.Generate(ctx, expiredClaims)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Parse the token
		var parsedClaims jwt.StandardClaims
		err = service.Parse(ctx, token, &parsedClaims)
		require.Error(t, err)
		require.Equal(t, jwt.ErrExpiredToken, err)
	})

	t.Run("with future token", func(t *testing.T) {
		// Generate a token that's not valid yet
		futureClaims := jwt.StandardClaims{
			Subject:    "user123",
			Issuer:     "gokit",
			ExpiresAt:  time.Now().Add(time.Hour).Unix(),
			NotBefore: time.Now().Add(time.Hour).Unix(), // Not valid yet
		}

		token, err := service.Generate(ctx, futureClaims)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Parse the token
		var parsedClaims jwt.StandardClaims
		err = service.Parse(ctx, token, &parsedClaims)
		require.Error(t, err)
		require.Equal(t, jwt.ErrInvalidToken, err)
	})
}

func TestSigningKeyDifference(t *testing.T) {
	// Create two services with different keys
	service1, err := jwt.New([]byte("secret1"))
	require.NoError(t, err)

	service2, err := jwt.New([]byte("secret2"))
	require.NoError(t, err)

	ctx := context.Background()

	// Generate a token with service1
	claims := jwt.StandardClaims{
		Subject:   "user123",
		Issuer:    "gokit",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	token, err := service1.Generate(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Try to parse the token with service2 (should fail)
	var parsedClaims jwt.StandardClaims
	err = service2.Parse(ctx, token, &parsedClaims)
	require.Error(t, err)
	require.Equal(t, jwt.ErrInvalidSignature, err)
}
