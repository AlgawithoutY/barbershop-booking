package controllers

import (
	"fmt"
	"net/http"

	"barbershop-backend/config"
	"barbershop-backend/models"

	"github.com/gin-gonic/gin"
)

func GetFinancialReport(c *gin.Context) {
	var bookings []models.Booking
	var totalRevenue float64

	// Filter status completed tanpa peka huruf besar/kecil (LOWER)
	err := config.DB.Preload("User").Preload("Barber").Preload("Service").
		Where("LOWER(status) = ?", "completed").
		Find(&bookings).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data laporan keuangan: " + err.Error(),
		})
		return
	}

	// Print log verifikasi ke terminal
	fmt.Printf("[DEBUG] Jumlah transaksi completed terbaca: %d\n", len(bookings))

	// Akumulasi total pendapatan
	for _, b := range bookings {
		totalRevenue += b.Service.Price
	}

	// Kirim respons JSON ke frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil laporan keuangan",
		"data": gin.H{
			"total_revenue":      totalRevenue,
			"total_transactions": len(bookings),
			"transactions":       bookings,
		},
	})
}