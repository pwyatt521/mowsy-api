package handlers

import (
	"net/http"
	"strconv"

	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		paymentService: services.NewPaymentService(),
	}
}

func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req services.CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.paymentService.CreatePaymentIntent(userID.(uint), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusCreated, response)
}

func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		PaymentID uint `json:"payment_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.paymentService.ConfirmPayment(req.PaymentID, userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, response)
}

func (h *PaymentHandler) GetPaymentHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	payments, err := h.paymentService.GetPaymentHistory(userID.(uint), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, payments)
}

func (h *PaymentHandler) GetPaymentByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid payment ID")
		return
	}

	payment, err := h.paymentService.GetPaymentByID(uint(paymentID), userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	utils.DataResponse(c, http.StatusOK, payment)
}