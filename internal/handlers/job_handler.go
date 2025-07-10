package handlers

import (
	"net/http"
	"strconv"

	"mowsy-api/internal/models"
	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService *services.JobService
}

func NewJobHandler() *JobHandler {
	return &JobHandler{
		jobService: services.NewJobService(),
	}
}

func (h *JobHandler) GetJobs(c *gin.Context) {
	var filters services.JobFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	jobs, err := h.jobService.GetJobs(filters)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, jobs)
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req services.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	job, err := h.jobService.CreateJob(userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, job)
}

func (h *JobHandler) GetJobByID(c *gin.Context) {
	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := h.jobService.GetJobByID(uint(jobID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, job)
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	var req services.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	job, err := h.jobService.UpdateJob(uint(jobID), userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, job)
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	err = h.jobService.DeleteJob(uint(jobID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Job deleted successfully", nil)
}

func (h *JobHandler) ApplyForJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	application, err := h.jobService.ApplyForJob(uint(jobID), userID.(uint), req.Message)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, application)
}

func (h *JobHandler) GetJobApplications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	applications, err := h.jobService.GetJobApplications(uint(jobID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusForbidden, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, applications)
}

func (h *JobHandler) UpdateApplicationStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	appIDStr := c.Param("app_id")
	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid application ID")
		return
	}

	var req struct {
		Status models.ApplicationStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.jobService.UpdateApplicationStatus(uint(jobID), uint(appID), userID.(uint), req.Status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Application status updated successfully", nil)
}

func (h *JobHandler) CompleteJob(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobIDStr := c.Param("id")
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid job ID")
		return
	}

	var req struct {
		ImageUrls []string `json:"image_urls" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.jobService.CompleteJob(uint(jobID), userID.(uint), req.ImageUrls)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Job completed successfully", nil)
}