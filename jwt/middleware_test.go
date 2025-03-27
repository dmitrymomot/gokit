package jwt_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dmitrymomot/gokit/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestContextKey jwt.ContextKey = "test_jwt_claims"
)

func TestMiddleware(t *testing.T) {
	// Create a JWT service for testing
	service, err := jwt.New([]byte("test-secret"))
	require.NoError(t, err)
	require.NotNil(t, service)

	// Create test claims
	testClaims := jwt.StandardClaims{
		Subject:   "test-user",
		Issuer:    "test-issuer",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}

	// Generate a test token
	token, err := service.Generate(testClaims)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Create a test handler that checks for claims in the context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := jwt.GetClaims(r.Context(), TestContextKey)
		if !ok {
			http.Error(w, "Claims not found in context", http.StatusInternalServerError)
			return
		}

		// Check if the claims contain expected values
		if claims["sub"] != testClaims.Subject {
			http.Error(w, "Subject mismatch", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("DefaultTokenExtractor", func(t *testing.T) {
		// Create middleware with default extractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in the Authorization header
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("MissingToken", func(t *testing.T) {
		// Create middleware with default extractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request without a token
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response (should be unauthorized)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create middleware with default extractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with an invalid token
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response (should be unauthorized)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("CustomExtractor", func(t *testing.T) {
		// Create a custom extractor that gets the token from a custom header
		customExtractor := func(r *http.Request) (string, error) {
			return r.Header.Get("X-Auth-Token"), nil
		}

		// Create middleware with the custom extractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
			Extractor:  customExtractor,
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in the custom header
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("X-Auth-Token", token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SkipMiddleware", func(t *testing.T) {
		// Create a skip function that skips requests to a specific path
		skipFunc := func(r *http.Request) bool {
			return r.URL.Path == "/skip"
		}

		// Create middleware with the skip function
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
			Skip:       skipFunc,
		})

		// Create a test handler that always succeeds
		skipHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("skipped"))
		})

		// Create a test server with the middleware
		handler := middleware(skipHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request to the skip path without a token
		req, err := http.NewRequest("GET", server.URL+"/skip", nil)
		require.NoError(t, err)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response (should be OK even without a token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Create a request to a different path without a token
		req, err = http.NewRequest("GET", server.URL+"/other", nil)
		require.NoError(t, err)

		// Send the request
		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response (should be unauthorized)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("GetClaimsAs", func(t *testing.T) {
		// Create a handler that uses GetClaimsAs to parse claims into a struct
		typedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var claims jwt.StandardClaims
			err := jwt.GetClaimsAs(r.Context(), TestContextKey, &claims)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Check if the claims match the original claims
			if claims.Subject != testClaims.Subject {
				http.Error(w, "Subject mismatch", http.StatusInternalServerError)
				return
			}

			// Return the claims as JSON
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(claims)
		})

		// Create middleware with default extractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
		})

		// Create a test server with the middleware
		handler := middleware(typedHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in the Authorization header
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Parse the response body
		var claims jwt.StandardClaims
		err = json.NewDecoder(resp.Body).Decode(&claims)
		require.NoError(t, err)

		// Check if the claims match the original claims
		assert.Equal(t, testClaims.Subject, claims.Subject)
		assert.Equal(t, testClaims.Issuer, claims.Issuer)
	})

	t.Run("WithClaims helper", func(t *testing.T) {
		// Create middleware using the WithClaims helper
		middleware := jwt.WithClaims[jwt.StandardClaims](service, TestContextKey)

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in the Authorization header
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("CookieTokenExtractor", func(t *testing.T) {
		// Create middleware using the CookieTokenExtractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
			Extractor:  jwt.CookieTokenExtractor("jwt"),
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in a cookie
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{
			Name:  "jwt",
			Value: token,
		})

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("QueryTokenExtractor", func(t *testing.T) {
		// Create middleware using the QueryTokenExtractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
			Extractor:  jwt.QueryTokenExtractor("token"),
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in a query parameter
		req, err := http.NewRequest("GET", server.URL+"?token="+token, nil)
		require.NoError(t, err)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("HeaderTokenExtractor", func(t *testing.T) {
		// Create middleware using the HeaderTokenExtractor
		middleware := jwt.Middleware(jwt.MiddlewareConfig{
			Service:    service,
			ContextKey: TestContextKey,
			Extractor:  jwt.HeaderTokenExtractor("X-API-Token"),
		})

		// Create a test server with the middleware
		handler := middleware(testHandler)
		server := httptest.NewServer(handler)
		defer server.Close()

		// Create a request with the token in a custom header
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Set("X-API-Token", token)

		// Send the request
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the response
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
