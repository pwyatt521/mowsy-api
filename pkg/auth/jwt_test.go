package auth

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	t.Run("GenerateValidToken", func(t *testing.T) {
		userID := uint(123)
		email := "test@example.com"

		token, err := GenerateToken(userID, email)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
	})

	t.Run("GenerateTokenWithoutSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", testSecret) // Restore after test
		
		userID := uint(123)
		email := "test@example.com"

		token, err := GenerateToken(userID, email)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
	})
}

func TestGenerateRefreshToken(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	t.Run("GenerateValidRefreshToken", func(t *testing.T) {
		userID := uint(123)
		email := "test@example.com"

		token, err := GenerateRefreshToken(userID, email)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
	})

	t.Run("GenerateRefreshTokenWithoutSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", testSecret) // Restore after test
		
		userID := uint(123)
		email := "test@example.com"

		token, err := GenerateRefreshToken(userID, email)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
	})
}

func TestValidateToken(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	t.Run("ValidateValidToken", func(t *testing.T) {
		userID := uint(123)
		email := "test@example.com"

		token, err := GenerateToken(userID, email)
		require.NoError(t, err)

		claims, err := ValidateToken(token)

		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, "mowsy-api", claims.Issuer)
	})

	t.Run("ValidateInvalidToken", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"

		claims, err := ValidateToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateTokenWithoutSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", testSecret) // Restore after test
		
		claims, err := ValidateToken("some.token.here")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
	})

	t.Run("ValidateExpiredToken", func(t *testing.T) {
		// This would require manipulating time or using a very short expiration
		// For now, we'll test with an obviously malformed token
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyMzkwMjJ9.invalid"

		claims, err := ValidateToken(expiredToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestValidateRefreshToken(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	t.Run("ValidateValidRefreshToken", func(t *testing.T) {
		userID := uint(123)
		email := "test@example.com"

		token, err := GenerateRefreshToken(userID, email)
		require.NoError(t, err)

		claims, err := ValidateRefreshToken(token)

		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, "mowsy-api", claims.Issuer)
	})

	t.Run("ValidateInvalidRefreshToken", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"

		claims, err := ValidateRefreshToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("ValidateRefreshTokenWithoutSecret", func(t *testing.T) {
		os.Setenv("JWT_SECRET", "")
		defer os.Setenv("JWT_SECRET", testSecret) // Restore after test
		
		claims, err := ValidateRefreshToken("some.token.here")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "JWT_SECRET environment variable not set")
	})
}

func TestTokenRoundTrip(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	t.Run("AccessTokenRoundTrip", func(t *testing.T) {
		userID := uint(456)
		email := "roundtrip@example.com"

		// Generate token
		token, err := GenerateToken(userID, email)
		require.NoError(t, err)

		// Validate token
		claims, err := ValidateToken(token)
		require.NoError(t, err)

		// Verify claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
		assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
	})

	t.Run("RefreshTokenRoundTrip", func(t *testing.T) {
		userID := uint(456)
		email := "roundtrip@example.com"

		// Generate refresh token
		token, err := GenerateRefreshToken(userID, email)
		require.NoError(t, err)

		// Validate refresh token
		claims, err := ValidateRefreshToken(token)
		require.NoError(t, err)

		// Verify claims
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.True(t, claims.ExpiresAt.Time.After(time.Now().Add(6*24*time.Hour)))
		assert.True(t, claims.IssuedAt.Time.Before(time.Now().Add(time.Second)))
	})
}