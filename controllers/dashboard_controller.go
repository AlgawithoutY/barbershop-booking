package controllers

import (
	"net/http"

	"barbershop-backend/config"

	"github.com/gin-gonic/gin"
)

func GetDashboardStats(c *gin.Context) {
	db := config.DB

	var totalPelanggan int64
	var totalBooking int64
	var totalTransaksiSelesai int64
	var totalPendapatan float64

	// 1. Hitung total pelanggan (role = 'customer')
	db.Table("users").Where("role = ? AND deleted_at IS NULL", "customer").Count(&totalPelanggan)

	// 2. Hitung total booking
	db.Table("bookings").Where("deleted_at IS NULL").Count(&totalBooking)

	// 3. Hitung total transaksi selesai (status 'completed' atau 'selesai')
	db.Table("bookings").Where("LOWER(status) IN ? AND deleted_at IS NULL", []string{"completed", "selesai", "done"}).Count(&totalTransaksiSelesai)

	// 4. Hitung total pendapatan berdasarkan jumlah transaksi selesai dikali tarif (misal Rp 50.000)
	const tarifPerCukur = 50000.0 
	totalPendapatan = float64(totalTransaksiSelesai) * tarifPerCukur

	// Mengembalikan data ke level utama agar sinkron dengan frontend
	// Kita sediakan 'total_transaksi_selesai' dan 'total_layanan' sekaligus (sebagai cadangan)
	c.JSON(http.StatusOK, gin.H{
		"total_pelanggan":         totalPelanggan,
		"total_booking":           totalBooking,
		"total_transaksi_selesai": totalTransaksiSelesai,
		"total_layanan":           totalTransaksiSelesai, // <--- Cadangan agar aman jika frontend masih membaca key ini
		"total_pendapatan":        totalPendapatan,
	})
}