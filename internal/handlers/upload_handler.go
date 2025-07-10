package handlers

import (
	"net/http"

	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	uploadService *services.UploadService
}

func NewUploadHandler() *UploadHandler {
	uploadService, err := services.NewUploadService()
	if err != nil {
		panic("Failed to initialize upload service: " + err.Error())
	}

	return &UploadHandler{
		uploadService: uploadService,
	}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "No file provided")
		return
	}

	category := c.PostForm("category")

	response, err := h.uploadService.UploadImage(userID.(uint), file, category)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, response)
}

func (h *UploadHandler) GetPresignedUploadURL(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		FileName string `json:"file_name" binding:"required"`
		MimeType string `json:"mime_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	url, err := h.uploadService.GetPresignedUploadURL(userID.(uint), req.FileName, req.MimeType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, gin.H{
		"upload_url": url,
	})
}

func (h *UploadHandler) DeleteFile(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		Key string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.uploadService.DeleteFile(req.Key)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "File deleted successfully", nil)
}