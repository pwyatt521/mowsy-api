package middleware

import (
	"net/http"

	"mowsy-api/internal/services"
	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

func InsuranceRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
			c.Abort()
			return
		}

		userService := services.NewUserService()
		user, err := userService.GetUserByID(userID.(uint))
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to verify user status")
			c.Abort()
			return
		}

		if !user.InsuranceVerified {
			utils.ErrorResponse(c, http.StatusForbidden, "Insurance verification required for this action")
			c.Abort()
			return
		}

		c.Next()
	}
}