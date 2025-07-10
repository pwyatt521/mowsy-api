package services

import (
	"errors"
	"fmt"

	"mowsy-api/internal/models"
	"mowsy-api/pkg/auth"
	"mowsy-api/pkg/database"
	"mowsy-api/internal/utils"

	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

func NewUserServiceWithDB(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	User         models.UserResponse  `json:"user"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	City      string `json:"city"`
	State     string `json:"state"`
	ZipCode   string `json:"zip_code"`
}

func (s *UserService) Register(req RegisterRequest) (*LoginResponse, error) {
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	if !utils.IsValidPassword(req.Password) {
		return nil, errors.New("password must be at least 8 characters long")
	}

	if req.Phone != "" && !utils.IsValidPhone(req.Phone) {
		return nil, errors.New("invalid phone number format")
	}

	if req.ZipCode != "" && !utils.IsValidZipCode(req.ZipCode) {
		return nil, errors.New("invalid zip code format")
	}

	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := models.User{
		Email:        utils.SanitizeString(req.Email),
		PasswordHash: hashedPassword,
		FirstName:    utils.SanitizeString(req.FirstName),
		LastName:     utils.SanitizeString(req.LastName),
		Phone:        utils.SanitizeString(req.Phone),
		Address:      utils.SanitizeString(req.Address),
		City:         utils.SanitizeString(req.City),
		State:        utils.SanitizeString(req.State),
		ZipCode:      utils.SanitizeString(req.ZipCode),
		IsActive:     true,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if user.Address != "" {
		geocodioService := NewGeocodioService()
		if err := geocodioService.GeocodeUser(&user); err != nil {
			fmt.Printf("Warning: Failed to geocode user address: %v\n", err)
		} else {
			s.db.Save(&user)
		}
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *UserService) Login(req LoginRequest) (*LoginResponse, error) {
	if !utils.IsValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	var user models.User
	if err := s.db.Where("email = ? AND is_active = true", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *UserService) RefreshToken(refreshToken string) (*LoginResponse, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	var user models.User
	if err := s.db.Where("id = ? AND is_active = true", claims.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	accessToken, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = true", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

func (s *UserService) UpdateUser(userID uint, req UpdateUserRequest) (*models.UserResponse, error) {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Phone != "" && !utils.IsValidPhone(req.Phone) {
		return nil, errors.New("invalid phone number format")
	}

	if req.ZipCode != "" && !utils.IsValidZipCode(req.ZipCode) {
		return nil, errors.New("invalid zip code format")
	}

	updates := map[string]interface{}{}
	if req.FirstName != "" {
		updates["first_name"] = utils.SanitizeString(req.FirstName)
	}
	if req.LastName != "" {
		updates["last_name"] = utils.SanitizeString(req.LastName)
	}
	if req.Phone != "" {
		updates["phone"] = utils.SanitizeString(req.Phone)
	}
	if req.Address != "" {
		updates["address"] = utils.SanitizeString(req.Address)
	}
	if req.City != "" {
		updates["city"] = utils.SanitizeString(req.City)
	}
	if req.State != "" {
		updates["state"] = utils.SanitizeString(req.State)
	}
	if req.ZipCode != "" {
		updates["zip_code"] = utils.SanitizeString(req.ZipCode)
	}

	if err := s.db.Model(user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if req.Address != "" {
		geocodioService := NewGeocodioService()
		if err := geocodioService.GeocodeUser(user); err != nil {
			fmt.Printf("Warning: Failed to geocode user address: %v\n", err)
		} else {
			s.db.Save(user)
		}
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *UserService) GetUserReviews(userID uint) ([]models.ReviewResponse, error) {
	var reviews []models.Review
	if err := s.db.Where("reviewed_user_id = ?", userID).
		Preload("Reviewer").
		Order("created_at DESC").
		Find(&reviews).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch reviews: %w", err)
	}

	responses := make([]models.ReviewResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = review.ToResponse()
	}

	return responses, nil
}

func (s *UserService) GetUserPublicProfile(userID uint) (*models.UserPublicProfile, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = true", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	profile := user.ToPublicProfile()
	return &profile, nil
}

func (s *UserService) UploadInsuranceDocument(userID uint, documentURL string) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}

	if err := s.db.Model(user).Updates(map[string]interface{}{
		"insurance_document_url": documentURL,
		"insurance_verified":     false,
	}).Error; err != nil {
		return fmt.Errorf("failed to update insurance document: %w", err)
	}

	return nil
}