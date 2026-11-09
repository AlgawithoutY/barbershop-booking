package controllers

import (
	"net/http"

	"barbershop-backend/config"
	"barbershop-backend/models"

	"github.com/gin-gonic/gin"
)

// POST /api/reviews - Customer memberikan ulasan untuk booking yang sudah selesai
func CreateReview(c *gin.Context) {
	var input struct {
		BookingID uint   `json:"booking_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid: " + err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}

	// 1. Cek apakah booking ada dan milik user yang sedang login
	var booking models.Booking
	if err := config.DB.First(&booking, input.BookingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data booking tidak ditemukan"})
		return
	}

	if booking.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak berhak mengulas booking ini"})
		return
	}

	// 2. Pastikan status booking sudah completed (selesai)
	if booking.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ulasan hanya dapat diberikan jika status booking sudah selesai (completed)"})
		return
	}

	// 3. Cek apakah booking ini sudah pernah direview sebelumnya
	var existingReview models.Review
	if err := config.DB.Where("booking_id = ?", input.BookingID).First(&existingReview).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda sudah memberikan ulasan untuk booking ini"})
		return
	}

	// 4. Simpan review baru ke database
	review := models.Review{
		BookingID: input.BookingID,
		UserID:    userID.(uint),
		BarberID:  booking.BarberID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}

	if err := config.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan ulasan: " + err.Error()})
		return
	}

	// Load relasi user dan barber untuk response
	config.DB.Preload("User").Preload("Barber").First(&review, review.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ulasan berhasil dikirim",
		"data":    review,
	})
}

// GET /api/barbers/:id/reviews - Mengambil daftar ulasan dan rata-rata rating barber tertentu
func GetBarberReviews(c *gin.Context) {
	barberID := c.Param("id")

	var reviews []models.Review
	// Preload Booking saja tanpa .Service karena sudah Barber-Centric
	if err := config.DB.Preload("User").Preload("Booking").Where("barber_id = ?", barberID).Order("created_at DESC").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data ulasan"})
		return
	}

	// Hitung rata-rata rating
	var totalRating int
	count := len(reviews)
	for _, r := range reviews {
		totalRating += r.Rating
	}

	var averageRating float64
	if count > 0 {
		averageRating = float64(totalRating) / float64(count)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Berhasil mengambil ulasan barber",
		"average_rating": averageRating,
		"total_reviews":  count,
		"data":           reviews,
	})
}

// GET /api/admin/reviews - Admin mengambil seluruh daftar ulasan pelanggan
func GetAllReviewsForAdmin(c *gin.Context) {
	var reviews []models.Review
	// Preload Booking saja tanpa .Service
	if err := config.DB.Preload("User").Preload("Barber").Preload("Booking").Order("created_at DESC").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data ulasan untuk admin: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil seluruh data ulasan",
		"data":    reviews,
	})
}