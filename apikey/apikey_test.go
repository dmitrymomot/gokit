package apikey_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/apikey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandom(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		apiKey, err := apikey.GenerateRandom()
		require.NoError(t, err)
		
		// Verify key properties
		assert.Len(t, apiKey, apikey.APIKeyLength*2) // hex encoded, so length is doubled
		assert.Regexp(t, "^[0-9a-f]{64}$", apiKey)   // hex encoded string pattern
		
		// Verify uniqueness of generated keys
		anotherKey, err := apikey.GenerateRandom()
		require.NoError(t, err)
		assert.NotEqual(t, apiKey, anotherKey, "Random keys should be unique")
		
		// Verify it's valid hex
		_, err = hex.DecodeString(apiKey)
		assert.NoError(t, err, "Generated key should be valid hex")
	})
}

func TestGenerateTimeOrdered(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		key, err := apikey.GenerateTimeOrdered()
		require.NoError(t, err)
		assert.NotEmpty(t, key)
		
		// Ensure we get a different key on the second call
		time.Sleep(1 * time.Millisecond) // Ensure clock advances
		key2, err := apikey.GenerateTimeOrdered()
		require.NoError(t, err)
		assert.NotEqual(t, key, key2, "Time-ordered keys should be different due to timestamp")
	})
}

func TestHashKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "successful hashing",
			apiKey:    "testkey123",
			secretKey: "secretkey456",
			wantErr:   false,
		},
		{
			name:      "empty api key",
			apiKey:    "",
			secretKey: "secretkey456",
			wantErr:   true,
		},
		{
			name:      "empty secret key",
			apiKey:    "testkey123",
			secretKey: "",
			wantErr:   true,
		},
		{
			name:      "both empty",
			apiKey:    "",
			secretKey: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := apikey.HashKey(tt.apiKey, tt.secretKey)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, apikey.ErrEmptyInput, err)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, got)
			assert.Len(t, got, 64) // SHA256 hash is 32 bytes, hex encoded to 64 chars

			// Verify deterministic output
			got2, err := apikey.HashKey(tt.apiKey, tt.secretKey)
			assert.NoError(t, err)
			assert.Equal(t, got, got2)
			
			// Verify it's valid hex 
			_, err = hex.DecodeString(got)
			assert.NoError(t, err, "Hash should be valid hex")
		})
	}
}

func TestSecureCompare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{
			name: "equal strings",
			a:    "test123",
			b:    "test123",
			want: true,
		},
		{
			name: "different strings",
			a:    "test123",
			b:    "test456",
			want: false,
		},
		{
			name: "different lengths",
			a:    "test123",
			b:    "test1234",
			want: false,
		},
		{
			name: "empty strings",
			a:    "",
			b:    "",
			want: false,
		},
		{
			name: "only first empty",
			a:    "",
			b:    "test123",
			want: false,
		},
		{
			name: "only second empty",
			a:    "test123",
			b:    "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apikey.SecureCompare(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateAndHash(t *testing.T) {
	// Test the combination of Generate and Hash
	apiKey, err := apikey.GenerateTimeOrdered()
	assert.NoError(t, err)

	secretKey := "mysecretkey123"
	hash1, err := apikey.HashKey(apiKey, secretKey)
	assert.NoError(t, err)

	// Verify the hash is consistent
	hash2, err := apikey.HashKey(apiKey, secretKey)
	assert.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Verify different API keys produce different hashes
	differentApiKey, err := apikey.GenerateRandom()
	assert.NoError(t, err)
	differentHash, err := apikey.HashKey(differentApiKey, secretKey)
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, differentHash)
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		hash      string
		secretKey string
		want      bool
	}{
		{
			name:      "valid verification",
			apiKey:    "testkey123",
			secretKey: "secretkey456",
			hash:      "", // Will be filled in setup
			want:      true,
		},
		{
			name:      "invalid api key",
			apiKey:    "wrongkey",
			secretKey: "secretkey456",
			hash:      "", // Will be filled in setup
			want:      false,
		},
		{
			name:      "empty api key",
			apiKey:    "",
			secretKey: "secretkey456",
			hash:      "somehash",
			want:      false,
		},
		{
			name:      "empty hash",
			apiKey:    "testkey123",
			secretKey: "secretkey456",
			hash:      "",
			want:      false,
		},
		{
			name:      "empty secret key",
			apiKey:    "testkey123",
			secretKey: "",
			hash:      "somehash",
			want:      false,
		},
		{
			name:      "all empty",
			apiKey:    "",
			secretKey: "",
			hash:      "",
			want:      false,
		},
		{
			name:      "invalid hash format",
			apiKey:    "testkey123",
			secretKey: "secretkey456",
			hash:      "not-hex-string",
			want:      false,
		},
		{
			name:      "wrong secret key",
			apiKey:    "testkey123",
			secretKey: "wrongsecret",
			hash:      "", // Will be filled in setup
			want:      false,
		},
	}

	// Pre-compute a valid hash for test cases that need it
	validHash, err := apikey.HashKey("testkey123", "secretkey456")
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For test cases that need the valid hash
			if tt.hash == "" && tt.name != "empty hash" {
				tt.hash = validHash
			}

			got := apikey.ValidateKey(tt.apiKey, tt.hash, tt.secretKey)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVerifyWithGeneratedKey(t *testing.T) {
	// Generate a new API key
	apiKey, err := apikey.GenerateRandom()
	assert.NoError(t, err)

	secretKey := "mysecretkey123"

	// Create hash
	hash, err := apikey.HashKey(apiKey, secretKey)
	assert.NoError(t, err)

	// Test cases
	tests := []struct {
		name      string
		apiKey    string
		hash      string
		secretKey string
		want      bool
	}{
		{
			name:      "valid key and hash",
			apiKey:    apiKey,
			hash:      hash,
			secretKey: secretKey,
			want:      true,
		},
		{
			name:      "modified api key",
			apiKey:    apiKey + "modified",
			hash:      hash,
			secretKey: secretKey,
			want:      false,
		},
		{
			name:      "modified secret key",
			apiKey:    apiKey,
			hash:      hash,
			secretKey: secretKey + "modified",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apikey.ValidateKey(tt.apiKey, tt.hash, tt.secretKey)
			assert.Equal(t, tt.want, got)
		})
	}
}
