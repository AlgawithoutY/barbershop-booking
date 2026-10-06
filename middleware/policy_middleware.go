package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole adalah middleware untuk memvalidasi apakah user memiliki role yang diizinkan
func RequireRole(allowedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Asumsi role user disimpan di context oleh AuthMiddleware sebelumnya
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: role not found"})
			c.Abort()
			return
		}

		// Cek apakah role user cocok dengan role yang diizinkan
		if userRole != allowedRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: you do not have permission to access this resource"})
			c.Abort()
			return
		}

		c.Next()
	}
}