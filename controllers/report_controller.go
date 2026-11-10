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

	// Ambil booking yang statusnya 'completed' atau 'confirmed' agar langsung muncul di laporan
	err := config.DB.Preload("User").Preload("Barber").
		Where("LOWER(status) IN ?", []string{"completed", "confirmed"}).
		Find(&bookings).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data laporan keuangan: " + err.Error(),
		})
		return
	}

	// Print log verifikasi ke terminal
	fmt.Printf("[DEBUG] Jumlah transaksi valid terbaca: %d\n", len(bookings))

	// Tarif flat standar per booking (misal Rp 50.000)
	const standardPrice = 50000.0

	// Format data agar setiap item booking membawa informasi nominal pendapatan (amount)
	var formattedBookings []gin.H
	for _, b := range bookings {
		totalRevenue += standardPrice

		// Konversi struct booking ke Map/gin.H agar frontend bisa membaca field .amount atau .price
		formattedBookings = append(formattedBookings, gin.H{
			"id":           b.ID,
			"user":         b.User,
			"barber":       b.Barber,
			"booking_time": b.BookingTime,
			"status":       b.Status,
			"amount":       standardPrice, // <-- INI AGAR KOLOM PENDAPATAN DI TABEL FRONTEND MUNCUL NOMINALNYA
		})
	}

	// Kirim respons JSON ke frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil laporan keuangan",
		"data": gin.H{
			"total_revenue":      totalRevenue,
			"total_transactions": len(bookings),
			"transactions":       formattedBookings,
		},
	})
}