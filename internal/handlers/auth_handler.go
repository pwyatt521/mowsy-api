package handlers

import (
	"net/http"

	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		userService: services.NewUserService(),
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body services.RegisterRequest true "User registration details"
// @Success 201 {object} services.LoginResponse "User registered successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid request body or validation error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.userService.Register(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body services.LoginRequest true "User login credentials"
// @Success 200 {object} services.LoginResponse "User logged in successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid request body"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.userService.Login(req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, response)
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh an expired access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param token body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} services.LoginResponse "Token refreshed successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid request body"
// @Failure 401 {object} utils.ErrorResponseModel "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, response)
}

// Logout godoc
// @Summary Logout user
// @Description Logout the current user (client-side token cleanup)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponseModel "User logged out successfully"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	utils.SuccessResponse(c, http.StatusOK, "Successfully logged out", nil)
}