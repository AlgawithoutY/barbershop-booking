package controllers

import (
	"barbershop-backend/config"
	"barbershop-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/services - Ambil semua daftar layanan
func GetServices(c *gin.Context) {
	var services []models.Service
	config.DB.Find(&services)
	c.JSON(http.StatusOK, gin.H{"data": services})
}

// POST /api/services - Tambah layanan baru
func CreateService(c *gin.Context) {
	var input models.Service
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&input)
	c.JSON(http.StatusCreated, gin.H{"message": "Layanan berhasil ditambahkan", "data": input})
}

// GET /api/services/:id - Ambil detail 1 layanan
func GetServiceByID(c *gin.Context) {
	var service models.Service
	id := c.Param("id")

	if err := config.DB.First(&service, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": service})
}

// PUT /api/services/:id - Update data layanan
func UpdateService(c *gin.Context) {
	var service models.Service
	id := c.Param("id")

	if err := config.DB.First(&service, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan"})
		return
	}

	var input models.Service
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&service).Updates(input)
	c.JSON(http.StatusOK, gin.H{"message": "Layanan berhasil diupdate", "data": service})
}

// DELETE /api/services/:id - Hapus layanan
func DeleteService(c *gin.Context) {
	var service models.Service
	id := c.Param("id")

	if err := config.DB.First(&service, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Layanan tidak ditemukan"})
		return
	}

	config.DB.Delete(&service)
	c.JSON(http.StatusOK, gin.H{"message": "Layanan berhasil dihapus"})
}