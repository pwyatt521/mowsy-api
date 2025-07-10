package handlers

import (
	"net/http"
	"strconv"

	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		adminService: services.NewAdminService(),
	}
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	var filters services.AdminUserListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	users, err := h.adminService.GetUsers(filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, users)
}

func (h *AdminHandler) DeactivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.adminService.DeactivateUser(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User deactivated successfully", nil)
}

func (h *AdminHandler) ActivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.adminService.ActivateUser(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User activated successfully", nil)
}

func (h *AdminHandler) VerifyInsurance(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = h.adminService.VerifyInsurance(uint(userID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Insurance verified successfully", nil)
}

func (h *AdminHandler) RemoveJob(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	err = h.adminService.RemoveJob(uint(jobID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Job removed successfully", nil)
}

func (h *AdminHandler) RemoveEquipment(c *gin.Context) {
	equipmentIDStr := c.Param("id")
	equipmentID, err := strconv.ParseUint(equipmentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid equipment ID")
		return
	}

	err = h.adminService.RemoveEquipment(uint(equipmentID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Equipment removed successfully", nil)
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, stats)
}