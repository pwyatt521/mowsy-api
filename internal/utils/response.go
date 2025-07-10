package utils

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponseModel represents error response structure
type ErrorResponseModel struct {
	Error   string `json:"error" example:"error"`
	Message string `json:"message,omitempty" example:"Invalid request"`
}

// SuccessResponseModel represents success response structure
type SuccessResponseModel struct {
	Message string      `json:"message" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// DataResponseModel represents data response structure with generic data field
type DataResponseModel struct {
	Data interface{} `json:"data"`
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponseModel{
		Error:   "error",
		Message: message,
	})
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	response := SuccessResponseModel{
		Message: message,
	}
	if data != nil {
		response.Data = data
	}
	c.JSON(statusCode, response)
}

func DataResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}