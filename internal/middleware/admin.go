package middleware

import (
	"net/http"
	"os"
	"strings"

	"mowsy-api/internal/utils"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminKey := os.Getenv("ADMIN_API_KEY")
		if adminKey == "" {
			utils.ErrorResponse(c, http.StatusServiceUnavailable, "Admin functionality not configured")
			c.Abort()
			return
		}

		authHeader := c.GetHeader("X-Admin-Key")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Admin API key required")
			c.Abort()
			return
		}

		if !strings.EqualFold(authHeader, adminKey) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid admin API key")
			c.Abort()
			return
		}

		c.Next()
	}
}