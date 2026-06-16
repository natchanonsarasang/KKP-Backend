package middlewares

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetJWtHeaderHandler_Configurations(t *testing.T) {
	// Backup env variables
	origJwk := os.Getenv("JWK_SET_URL")
	origSupa := os.Getenv("SUPABASE_URL")
	origSecret := os.Getenv("JWT_SECRET_KEY")

	defer func() {
		os.Setenv("JWK_SET_URL", origJwk)
		os.Setenv("SUPABASE_URL", origSupa)
		os.Setenv("JWT_SECRET_KEY", origSecret)
	}()

	// Start local mock JWKS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"keys": [
				{
					"kty": "RSA",
					"use": "sig",
					"kid": "test-key-id",
					"alg": "RS256",
					"n": "u1W_a359g1G3u2r4",
					"e": "AQAB"
				}
			]
		}`))
	}))
	defer server.Close()

	// 1. Test fallback to JWT_SECRET_KEY
	t.Run("Fallback to JWT_SECRET_KEY", func(t *testing.T) {
		os.Unsetenv("JWK_SET_URL")
		os.Unsetenv("SUPABASE_URL")
		os.Setenv("JWT_SECRET_KEY", "test-secret")

		handler := SetJWtHeaderHandler()
		assert.NotNil(t, handler)
	})

	// 2. Test configuration with JWK_SET_URL
	t.Run("Using JWK_SET_URL", func(t *testing.T) {
		os.Setenv("JWK_SET_URL", server.URL)
		os.Unsetenv("SUPABASE_URL")

		handler := SetJWtHeaderHandler()
		assert.NotNil(t, handler)
	})

	// 3. Test configuration with SUPABASE_URL
	t.Run("Using SUPABASE_URL", func(t *testing.T) {
		os.Unsetenv("JWK_SET_URL")
		// The middleware will append "/auth/v1/.well-known/jwks.json" to the URL,
		// so we mock a supabaseURL pointing to our test server URL.
		os.Setenv("SUPABASE_URL", server.URL)

		// Note: The mock server responds to any path with the JWKS, so the sub-path check works fine.
		handler := SetJWtHeaderHandler()
		assert.NotNil(t, handler)
	})
}
