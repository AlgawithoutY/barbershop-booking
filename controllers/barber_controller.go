package controllers

import (
	"barbershop-backend/config"
	"barbershop-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /api/barbers - Ambil semua data barber
func GetBarbers(c *gin.Context) {
	var barbers []models.Barber
	config.DB.Find(&barbers)
	c.JSON(http.StatusOK, gin.H{"data": barbers})
}

// POST /api/barbers - Tambah barber baru
func CreateBarber(c *gin.Context) {
	var input models.Barber
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&input)
	c.JSON(http.StatusCreated, gin.H{"message": "Barber berhasil ditambahkan", "data": input})
}

// GET /api/barbers/:id - Ambil detail 1 barber
func GetBarberByID(c *gin.Context) {
	var barber models.Barber
	id := c.Param("id")

	if err := config.DB.First(&barber, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": barber})
}

// PUT /api/barbers/:id - Update data barber
func UpdateBarber(c *gin.Context) {
	var barber models.Barber
	id := c.Param("id")

	if err := config.DB.First(&barber, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}

	var input models.Barber
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&barber).Updates(input)
	c.JSON(http.StatusOK, gin.H{"message": "Data barber berhasil diupdate", "data": barber})
}

// DELETE /api/barbers/:id - Hapus barber
func DeleteBarber(c *gin.Context) {
	var barber models.Barber
	id := c.Param("id")

	if err := config.DB.First(&barber, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}

	config.DB.Delete(&barber)
	c.JSON(http.StatusOK, gin.H{"message": "Barber berhasil dihapus"})
}