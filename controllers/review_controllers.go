package controllers

import (
	"net/http"
	"barbershop-backend/config" // Sesuaikan dengan nama modul di go.mod Anda
	"barbershop-backend/models"
	"github.com/gin-gonic/gin"
)

// GetReviews untuk mengambil semua daftar ulasan (untuk ditampilkan di Admin Laporan & Ulasan)
func GetReviews(c *gin.Context) {
	var reviews []models.Review
	
	// Mengambil data ulasan beserta relasi User dan Booking (termasuk Barber & Service)
	if err := config.DB.Preload("User").Preload("Booking.Barber").Preload("Booking.Service").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data ulasan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data ulasan",
		"data":    reviews,
	})
}

// CreateReview untuk menambahkan ulasan baru dari pelanggan
func CreateReview(c *gin.Context) {
	var input struct {
		BookingID uint   `json:"booking_id" binding:"required"`
		UserID    uint   `json:"user_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah booking ada dan statusnya sudah 'completed'
	var booking models.Booking
	if err := config.DB.First(&booking, input.BookingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data booking tidak ditemukan"})
		return
	}

	if booking.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ulasan hanya dapat diberikan jika status booking sudah 'completed'"})
		return
	}

	// Buat review baru
	review := models.Review{
		BookingID: input.BookingID,
		UserID:    input.UserID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}

	if err := config.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan ulasan (kemungkinan booking ini sudah diulas)"})
		return
	}

	config.DB.Preload("User").Preload("Booking.Barber").Preload("Booking.Service").First(&review, review.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ulasan berhasil dikirim",
		"data":    review,
	})
}