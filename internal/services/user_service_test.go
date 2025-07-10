package services

import (
	"testing"

	"mowsy-api/internal/models"
	"mowsy-api/internal/testutils"
	"mowsy-api/pkg/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupUserService() (*UserService, *gorm.DB) {
	db := testutils.SetupTestDB()
	
	// Create a user service with the test database
	service := NewUserServiceWithDB(db)
	
	return service, db
}

func TestUserService_Register(t *testing.T) {
	service, db := setupUserService()
	defer testutils.CleanupTestDB(db)

	t.Run("ValidRegistration", func(t *testing.T) {
		req := RegisterRequest{
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Phone:     "555-123-4567",
			Address:   "123 Test St",
			City:      "Test City",
			State:     "TS",
			ZipCode:   "12345",
		}

		response, err := service.Register(req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, req.Email, response.User.Email)
		assert.Equal(t, req.FirstName, response.User.FirstName)
		assert.Equal(t, req.LastName, response.User.LastName)
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		req := RegisterRequest{
			Email:     "invalid-email",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		response, err := service.Register(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("WeakPassword", func(t *testing.T) {
		req := RegisterRequest{
			Email:     "test2@example.com",
			Password:  "weak",
			FirstName: "Test",
			LastName:  "User",
		}

		response, err := service.Register(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "password must be at least 8 characters")
	})

	t.Run("InvalidPhone", func(t *testing.T) {
		req := RegisterRequest{
			Email:     "test3@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			Phone:     "invalid-phone",
		}

		response, err := service.Register(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid phone number format")
	})

	t.Run("InvalidZipCode", func(t *testing.T) {
		req := RegisterRequest{
			Email:     "test4@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			ZipCode:   "invalid",
		}

		response, err := service.Register(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid zip code format")
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		// First registration
		req1 := RegisterRequest{
			Email:     "duplicate@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		_, err := service.Register(req1)
		require.NoError(t, err)

		// Second registration with same email
		req2 := RegisterRequest{
			Email:     "duplicate@example.com",
			Password:  "password456",
			FirstName: "Another",
			LastName:  "User",
		}

		response, err := service.Register(req2)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user with this email already exists")
	})
}

func TestUserService_Login(t *testing.T) {
	service, db := setupUserService()
	defer testutils.CleanupTestDB(db)

	// Create a test user
	password := "password123"
	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err)

	user := &models.User{
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		FirstName:    "Test",
		LastName:     "User",
		IsActive:     true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	t.Run("ValidLogin", func(t *testing.T) {
		req := LoginRequest{
			Email:    "test@example.com",
			Password: password,
		}

		response, err := service.Login(req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, user.Email, response.User.Email)
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		req := LoginRequest{
			Email:    "invalid-email",
			Password: password,
		}

		response, err := service.Login(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("WrongPassword", func(t *testing.T) {
		req := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		response, err := service.Login(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		req := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: password,
		}

		response, err := service.Login(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("InactiveUser", func(t *testing.T) {
		// Create inactive user
		inactivePassword := "password123"
		inactiveHashedPassword, err := auth.HashPassword(inactivePassword)
		require.NoError(t, err)

		inactiveUser := &models.User{
			Email:        "inactive@example.com",
			PasswordHash: inactiveHashedPassword,
			FirstName:    "Inactive",
			LastName:     "User",
			IsActive:     false,
		}
		err = db.Create(inactiveUser).Error
		require.NoError(t, err)

		// Verify the user was created with IsActive = false
		var createdUser models.User
		err = db.Where("email = ?", "inactive@example.com").First(&createdUser).Error
		require.NoError(t, err)
		assert.False(t, createdUser.IsActive, "User should be inactive")

		req := LoginRequest{
			Email:    "inactive@example.com",
			Password: inactivePassword,
		}

		response, err := service.Login(req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	service, db := setupUserService()
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := testutils.CreateTestUser(db)

	t.Run("ValidUpdate", func(t *testing.T) {
		req := UpdateUserRequest{
			FirstName: "Updated",
			LastName:  "Name",
			Phone:     "555-987-6543",
		}

		response, err := service.UpdateUser(user.ID, req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Updated", response.FirstName)
		assert.Equal(t, "Name", response.LastName)
		assert.Equal(t, "555-987-6543", response.Phone)
	})

	t.Run("InvalidPhone", func(t *testing.T) {
		req := UpdateUserRequest{
			Phone: "invalid-phone",
		}

		response, err := service.UpdateUser(user.ID, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid phone number format")
	})

	t.Run("InvalidZipCode", func(t *testing.T) {
		req := UpdateUserRequest{
			ZipCode: "invalid",
		}

		response, err := service.UpdateUser(user.ID, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid zip code format")
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		req := UpdateUserRequest{
			FirstName: "Updated",
		}

		response, err := service.UpdateUser(99999, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	service, db := setupUserService()
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := testutils.CreateTestUser(db)

	t.Run("ValidUserID", func(t *testing.T) {
		foundUser, err := service.GetUserByID(user.ID)

		require.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.FirstName, foundUser.FirstName)
	})

	t.Run("NonexistentUserID", func(t *testing.T) {
		foundUser, err := service.GetUserByID(99999)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("InactiveUser", func(t *testing.T) {
		// Create inactive user
		inactiveUser := &models.User{
			Email:     "inactive@example.com",
			FirstName: "Inactive",
			LastName:  "User",
			IsActive:  false,
		}
		err := db.Create(inactiveUser).Error
		require.NoError(t, err)

		foundUser, err := service.GetUserByID(inactiveUser.ID)

		assert.Error(t, err)
		assert.Nil(t, foundUser)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_GetUserPublicProfile(t *testing.T) {
	service, db := setupUserService()
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := testutils.CreateTestUser(db)

	t.Run("ValidUserProfile", func(t *testing.T) {
		profile, err := service.GetUserPublicProfile(user.ID)

		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, user.ID, profile.ID)
		assert.Equal(t, user.FirstName, profile.FirstName)
		assert.Equal(t, user.LastName, profile.LastName)
		assert.Equal(t, user.ElementarySchoolDistrictName, profile.ElementarySchoolDistrictName)
		assert.Equal(t, user.InsuranceVerified, profile.InsuranceVerified)
		// Should not contain sensitive information like email, phone, etc.
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		profile, err := service.GetUserPublicProfile(99999)

		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_UploadInsuranceDocument(t *testing.T) {
	service, db := setupUserService()
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := testutils.CreateTestUser(db)

	t.Run("ValidInsuranceUpload", func(t *testing.T) {
		documentURL := "https://s3.amazonaws.com/bucket/insurance-doc.pdf"

		err := service.UploadInsuranceDocument(user.ID, documentURL)

		require.NoError(t, err)

		// Verify the document was saved and insurance is not yet verified
		var updatedUser models.User
		err = db.First(&updatedUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, documentURL, updatedUser.InsuranceDocumentURL)
		assert.False(t, updatedUser.InsuranceVerified)
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		documentURL := "https://s3.amazonaws.com/bucket/insurance-doc.pdf"

		err := service.UploadInsuranceDocument(99999, documentURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}