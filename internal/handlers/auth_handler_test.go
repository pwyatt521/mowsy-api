package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mowsy-api/internal/services"
	"mowsy-api/internal/testutils"
	"mowsy-api/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthHandler() (*AuthHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Set up test environment
	originalSecret := os.Getenv("JWT_SECRET")
	if originalSecret == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-testing")
	}
	
	handler := NewAuthHandler()
	return handler, r
}

func TestAuthHandler_Register(t *testing.T) {
	handler, r := setupAuthHandler()
	
	// Setup route
	r.POST("/register", handler.Register)
	
	// Override the user service with one that uses test DB
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	handler.userService = services.NewUserServiceWithDB(db)

	t.Run("ValidRegistration", func(t *testing.T) {
		reqBody := services.RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Phone:     "555-123-4567",
		}

		jsonData, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("InvalidRequestBody", func(t *testing.T) {
		invalidJSON := `{"email": "test@example.com", "password": 123}` // invalid password type

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		reqBody := services.RegisterRequest{
			Email: "test@example.com",
			// Missing password and other required fields
		}

		jsonData, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})
}

func TestAuthHandler_Login(t *testing.T) {
	handler, r := setupAuthHandler()
	
	// Setup route
	r.POST("/login", handler.Login)

	t.Run("InvalidRequestBody", func(t *testing.T) {
		invalidJSON := `{"email": 123, "password": "password123"}` // invalid email type

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		reqBody := services.LoginRequest{
			Email: "test@example.com",
			// Missing password
		}

		jsonData, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	handler, r := setupAuthHandler()
	
	// Setup route
	r.POST("/refresh", handler.RefreshToken)
	
	// Override the user service with one that uses test DB
	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)
	handler.userService = services.NewUserServiceWithDB(db)

	t.Run("InvalidRequestBody", func(t *testing.T) {
		invalidJSON := `{"refresh_token": 123}` // invalid token type

		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer([]byte(invalidJSON)))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("MissingRefreshToken", func(t *testing.T) {
		reqBody := map[string]string{} // Empty body

		jsonData, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request body")
	})

	t.Run("ValidRefreshTokenFormat", func(t *testing.T) {
		// Generate a valid refresh token for testing
		userID := uint(123)
		email := "test@example.com"
		refreshToken, err := auth.GenerateRefreshToken(userID, email)
		require.NoError(t, err)

		reqBody := map[string]string{
			"refresh_token": refreshToken,
		}

		jsonData, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// This will fail because user doesn't exist in DB, but it tests the request format
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	handler, r := setupAuthHandler()
	
	// Setup route
	r.POST("/logout", handler.Logout)

	t.Run("LogoutAlwaysSucceeds", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/logout", nil)
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Successfully logged out")
	})
}

// Integration test that demonstrates the full flow with a test database
func TestAuthHandler_IntegrationTest(t *testing.T) {
	// This test would require proper dependency injection to work fully
	t.Skip("Skipping integration test - requires proper DI setup")

	db := testutils.SetupTestDB()
	defer testutils.CleanupTestDB(db)

	// Setup handler with test database
	userService := &services.UserService{} // Would need to inject test DB here
	handler := &AuthHandler{userService: userService}

	r := gin.New()
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	// Test full registration and login flow
	t.Run("FullRegistrationAndLoginFlow", func(t *testing.T) {
		// Register
		regReq := services.RegisterRequest{
			Email:     "integration@example.com",
			Password:  "password123",
			FirstName: "Integration",
			LastName:  "Test",
		}

		jsonData, err := json.Marshal(regReq)
		require.NoError(t, err)

		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var regResponse services.LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &regResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, regResponse.AccessToken)
		assert.NotEmpty(t, regResponse.RefreshToken)

		// Login with same credentials
		loginReq := services.LoginRequest{
			Email:    "integration@example.com",
			Password: "password123",
		}

		jsonData, err = json.Marshal(loginReq)
		require.NoError(t, err)

		req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var loginResponse services.LoginResponse
		err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResponse.AccessToken)
		assert.NotEmpty(t, loginResponse.RefreshToken)
	})
}