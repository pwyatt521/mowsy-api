package handlers

import (
	"net/http"
	"strconv"

	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(),
	}
}

// GetCurrentUser godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse "User profile retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "User not authenticated"
// @Failure 404 {object} utils.ErrorResponseModel "User not found"
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, user.ToResponse())
}

// UpdateCurrentUser godoc
// @Summary Update current user profile
// @Description Update the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body services.UpdateUserRequest true "User update details"
// @Success 200 {object} models.UserResponse "User profile updated successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid request body or validation error"
// @Failure 401 {object} utils.ErrorResponseModel "User not authenticated"
// @Router /users/me [put]
func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.userService.UpdateUser(userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, response)
}

// GetUserReviews godoc
// @Summary Get user reviews
// @Description Get all reviews for a specific user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} models.ReviewResponse "User reviews retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid user ID"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /users/{id}/reviews [get]
func (h *UserHandler) GetUserReviews(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	reviews, err := h.userService.GetUserReviews(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, reviews)
}

// GetUserPublicProfile godoc
// @Summary Get user public profile
// @Description Get the public profile of a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.UserPublicProfile "User public profile retrieved successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid user ID"
// @Failure 404 {object} utils.ErrorResponseModel "User not found"
// @Router /users/{id}/profile [get]
func (h *UserHandler) GetUserPublicProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	profile, err := h.userService.GetUserPublicProfile(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, profile)
}

// UploadInsuranceDocumentRequest represents the request body for insurance document upload
type UploadInsuranceDocumentRequest struct {
	DocumentURL string `json:"document_url" binding:"required"`
}

// UploadInsuranceDocument godoc
// @Summary Upload insurance document
// @Description Upload an insurance document for the current user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param document body UploadInsuranceDocumentRequest true "Insurance document details"
// @Success 200 {object} utils.SuccessResponseModel "Insurance document uploaded successfully"
// @Failure 400 {object} utils.ErrorResponseModel "Invalid request body"
// @Failure 401 {object} utils.ErrorResponseModel "User not authenticated"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /users/me/insurance [post]
func (h *UserHandler) UploadInsuranceDocument(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req UploadInsuranceDocumentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.userService.UploadInsuranceDocument(userID.(uint), req.DocumentURL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Insurance document uploaded successfully", nil)
}