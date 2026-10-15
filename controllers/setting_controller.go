package controllers

import (
	"net/http"

	"barbershop-backend/config"
	"barbershop-backend/models"

	"github.com/gin-gonic/gin"
)

// GET /api/admin/settings
func GetSettings(c *gin.Context) {
	var setting models.Setting
	if err := config.DB.First(&setting).Error; err != nil {
		// Jika data belum ada di DB, kembalikan data default
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": models.Setting{
				Name:    "Barberly",
				Address: "Jl. Merdeka No. 45, Jakarta Selatan",
				Phone:   "081234567890",
				Email:   "admin@barberly.com",
				OpenAt:  "09:00",
				CloseAt: "21:00",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   setting,
	})
}

// PUT /api/admin/settings
func UpdateSettings(c *gin.Context) {
	var input models.Setting
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	var setting models.Setting
	if err := config.DB.First(&setting).Error; err != nil {
		// Buat record baru jika belum ada
		if err := config.DB.Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Gagal menyimpan pengaturan"})
			return
		}
	} else {
		// Update record yang sudah ada
		config.DB.Model(&setting).Updates(input)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pengaturan berhasil diperbarui",
		"data":    input,
	})
}