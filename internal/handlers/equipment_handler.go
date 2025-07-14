package handlers

import (
	"net/http"
	"strconv"
	"time"

	"mowsy-api/internal/models"
	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type EquipmentHandler struct {
	equipmentService *services.EquipmentService
}

func NewEquipmentHandler() *EquipmentHandler {
	return &EquipmentHandler{
		equipmentService: services.NewEquipmentService(),
	}
}

func (h *EquipmentHandler) GetEquipment(c *gin.Context) {
	var filters services.EquipmentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	// Check if user is authenticated (optional for public endpoint)
	var userID *uint
	if userIDValue, exists := c.Get("user_id"); exists {
		if uid, ok := userIDValue.(uint); ok {
			userID = &uid
		}
	}

	equipment, err := h.equipmentService.GetEquipmentWithUser(filters, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, equipment)
}

func (h *EquipmentHandler) CreateEquipment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req services.CreateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	equipment, err := h.equipmentService.CreateEquipment(userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, equipment)
}

func (h *EquipmentHandler) GetEquipmentByID(c *gin.Context) {
	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	equipment, err := h.equipmentService.GetEquipmentByID(uint(equipmentID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, equipment)
}

func (h *EquipmentHandler) UpdateEquipment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	var req services.UpdateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	equipment, err := h.equipmentService.UpdateEquipment(uint(equipmentID), userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, equipment)
}

func (h *EquipmentHandler) DeleteEquipment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	err = h.equipmentService.DeleteEquipment(uint(equipmentID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Equipment deleted successfully", nil)
}

func (h *EquipmentHandler) RequestRental(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	var req struct {
		StartDate time.Time `json:"start_date" binding:"required"`
		EndDate   time.Time `json:"end_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	rental, err := h.equipmentService.RequestRental(uint(equipmentID), userID.(uint), req.StartDate, req.EndDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, rental)
}

func (h *EquipmentHandler) GetEquipmentRentals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	rentals, err := h.equipmentService.GetEquipmentRentals(uint(equipmentID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, rentals)
}

func (h *EquipmentHandler) UpdateRentalStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	rentalIDStr := c.Param("rental_id")
	rentalID, err := strconv.ParseUint(rentalIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid rental ID")
		return
	}

	var req struct {
		Status models.RentalStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.equipmentService.UpdateRentalStatus(uint(equipmentID), uint(rentalID), userID.(uint), req.Status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rental status updated successfully", nil)
}

func (h *EquipmentHandler) CompleteRental(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	rentalIDStr := c.Param("rental_id")
	rentalID, err := strconv.ParseUint(rentalIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid rental ID")
		return
	}

	var req struct {
		ReturnNotes string `json:"return_notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.equipmentService.CompleteRental(uint(rentalID), userID.(uint), req.ReturnNotes)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Rental completed successfully", nil)
}

// GetMyEquipment godoc
// @Summary Get current user's posted equipment
// @Description Get all equipment posted by the currently authenticated user
// @Tags equipment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category query string false "Filter by equipment category"
// @Param is_available query bool false "Filter by availability status"
// @Param fuel_type query string false "Filter by fuel type"
// @Param power_type query string false "Filter by power type"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Success 200 {array} models.EquipmentResponse "User's posted equipment retrieved successfully"
// @Failure 401 {object} utils.ErrorResponseModel "User not authenticated"
// @Failure 500 {object} utils.ErrorResponseModel "Internal server error"
// @Router /equipment/my [get]
func (h *EquipmentHandler) GetMyEquipment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var filters services.EquipmentFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	equipment, err := h.equipmentService.GetEquipmentByUserID(userID.(uint), filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, equipment)
}