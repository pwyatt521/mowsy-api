package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mowsy-api/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	r := setupGin()
	
	// Protected route
	r.GET("/protected", AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(500, gin.H{"error": "user_id not set"})
			return
		}
		c.JSON(200, gin.H{"user_id": userID})
	})

	t.Run("ValidToken", func(t *testing.T) {
		// Generate a valid token
		userID := uint(123)
		email := "test@example.com"
		token, err := auth.GenerateToken(userID, email)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "123")
	})

	t.Run("MissingAuthHeader", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header required")
	})

	t.Run("InvalidAuthHeaderFormat", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("EmptyBearerToken", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})
}

func TestOptionalAuthMiddleware(t *testing.T) {
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)
	
	testSecret := "test-secret-key-for-jwt-testing"
	os.Setenv("JWT_SECRET", testSecret)

	r := setupGin()
	
	// Route with optional auth
	r.GET("/optional", OptionalAuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if exists {
			c.JSON(200, gin.H{"authenticated": true, "user_id": userID})
		} else {
			c.JSON(200, gin.H{"authenticated": false})
		}
	})

	t.Run("ValidTokenOptional", func(t *testing.T) {
		// Generate a valid token
		userID := uint(123)
		email := "test@example.com"
		token, err := auth.GenerateToken(userID, email)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "\"authenticated\":true")
		assert.Contains(t, w.Body.String(), "123")
	})

	t.Run("NoTokenOptional", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/optional", nil)
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "\"authenticated\":false")
	})

	t.Run("InvalidTokenOptional", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer invalid.token")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "\"authenticated\":false")
	})

	t.Run("InvalidHeaderFormatOptional", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "\"authenticated\":false")
	})
}