package middleware

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// RequireRoles membatasi akses berdasarkan role user
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Role tidak ditemukan"})
            c.Abort()
            return
        }

        userRole := role.(string)
        
        // Cek apakah role user ada di dalam daftar yang diizinkan
        allowed := false
        for _, r := range allowedRoles {
            if userRole == r {
                allowed = true
                break
            }
        }

        if !allowed {
            c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Kamu tidak memiliki hak akses untuk endpoint ini"})
            c.Abort()
            return
        }

        c.Next()
    }
}