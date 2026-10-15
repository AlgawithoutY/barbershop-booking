package controllers

import (
	"net/http"
	"barbershop-backend/config" // Sesuaikan dengan lokasi setup DB kalian (misal: config.DB)

	"github.com/gin-gonic/gin"
)

func GetDashboardStats(c *gin.Context) {
	db := config.DB

	var totalPelanggan int64
	var totalBooking int64
	var totalLayanan int64
	var totalPendapatan float64

	// 1. Hitung total pelanggan (role = 'customer')
	db.Table("users").Where("role = ? AND deleted_at IS NULL", "customer").Count(&totalPelanggan)

	// 2. Hitung total booking
	db.Table("bookings").Where("deleted_at IS NULL").Count(&totalBooking)

	// 3. Hitung total layanan
	db.Table("services").Where("deleted_at IS NULL").Count(&totalLayanan)

	// 4. Hitung total pendapatan dari booking yang statusnya 'completed'
	// Menggunakan JOIN ke tabel services karena harga (price) ada di tabel services
	db.Table("bookings").
		Joins("JOIN services ON services.id = bookings.service_id").
		Where("bookings.status = ? AND bookings.deleted_at IS NULL", "completed").
		Select("COALESCE(SUM(services.price), 0)").
		Scan(&totalPendapatan)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"total_pelanggan":  totalPelanggan,
			"total_booking":    totalBooking,
			"total_layanan":    totalLayanan,
			"total_pendapatan": totalPendapatan,
		},
	})
}