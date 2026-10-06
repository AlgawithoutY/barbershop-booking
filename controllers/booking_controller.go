package controllers

import (
	"net/http"
	"time"

	"barbershop-backend/config"
	"barbershop-backend/models"

	"github.com/gin-gonic/gin"
)

// GET /api/bookings - Admin melihat semua booking dengan opsi filter (date, status, barber_id)
func GetBookings(c *gin.Context) {
	var bookings []models.Booking
	query := config.DB.Preload("User").Preload("Barber").Preload("Service")

	// Ambil parameter query dari URL
	dateFilter := c.Query("date")      // Format: YYYY-MM-DD
	statusFilter := c.Query("status")  // Format: pending, confirmed, completed, cancelled
	barberIDFilter := c.Query("barber_id") // Format: ID angka

	// Terapkan filter tanggal jika diisi
	if dateFilter != "" {
		startDate := dateFilter + " 00:00:00"
		endDate := dateFilter + " 23:59:59"
		query = query.Where("booking_time BETWEEN ? AND ?", startDate, endDate)
	}

	// Terapkan filter status jika diisi
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	// Terapkan filter barber_id jika diisi
	if barberIDFilter != "" {
		query = query.Where("barber_id = ?", barberIDFilter)
	}

	// Eksekusi query dengan urutan booking terbaru
	if err := query.Order("booking_time DESC").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil daftar booking",
		"total":   len(bookings),
		"data":    bookings,
	})
}

// POST /api/bookings - Buat booking baru dengan validasi anti-bentrok & jam operasional
func CreateBooking(c *gin.Context) {
	var input models.Booking
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil user_id dari JWT token yang disimpan oleh AuthMiddleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}

	// Set UserID dari token yang sedang login (agar tidak perlu manual input user_id di body)
	input.UserID = userID.(uint)

	// 1. Validasi Jam Operasional (Buka pukul 09:00 sampai 21:00 WIB)
	// Gunakan FixedZone UTC+7 (WIB) tanpa manipulasi integer manual
	loc := time.FixedZone("WIB", 7*3600)
	bookingWIB := input.BookingTime.In(loc)

	hour := bookingWIB.Hour()
	if hour < 9 || hour >= 21 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Di luar jam operasional toko (Buka pukul 09:00 - 21:00 WIB)",
		})
		return
	}

	// 2. Validasi Tanggal Masa Lalu (Tidak boleh booking di waktu yang sudah lewat)
	if input.BookingTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Tidak dapat memesan untuk waktu yang sudah berlalu",
		})
		return
	}

	// 3. Cek Double-Booking (Apakah barber sudah dibooking di jam yang sama?)
	var existingBooking models.Booking
	err := config.DB.Where("barber_id = ? AND booking_time = ?", input.BarberID, input.BookingTime).First(&existingBooking).Error
	if err == nil {
		// Jika err == nil, berarti data ditemukan (artinya jadwal sudah terisi)
		c.JSON(http.StatusConflict, gin.H{
			"error": "Jadwal bentrok: Barber ini sudah memiliki jadwal booking di waktu tersebut. Silakan pilih jam lain.",
		})
		return
	}

	// 4. Jika semua validasi lolos, simpan ke database
	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat booking: " + err.Error()})
		return
	}

	// Ambil relasi Preload untuk ditampilkan di response
	config.DB.Preload("User").Preload("Barber").Preload("Service").First(&input, input.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Booking berhasil dibuat",
		"data":    input,
	})
}

// GET /api/bookings/:id - Ambil detail 1 booking
func GetBookingByID(c *gin.Context) {
	var booking models.Booking
	id := c.Param("id")

	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": booking})
}

// PUT /api/bookings/:id - Update status atau data booking
func UpdateBooking(c *gin.Context) {
	var booking models.Booking
	id := c.Param("id")

	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	var input models.Booking
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&booking).Updates(input)
	config.DB.Preload("User").Preload("Barber").Preload("Service").First(&booking, booking.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Booking berhasil diupdate", "data": booking})
}

// DELETE /api/bookings/:id - Batalkan/Hapus booking
func DeleteBooking(c *gin.Context) {
	var booking models.Booking
	id := c.Param("id")

	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	config.DB.Delete(&booking)
	c.JSON(http.StatusOK, gin.H{"message": "Booking berhasil dibatalkan"})
}

// PUT /api/bookings/:id/cancel - Batalkan booking dengan validasi kepemilikan & waktu
func CancelBooking(c *gin.Context) {
	id := c.Param("id")

	// 1. Ambil user_id dan role dari JWT claims
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}

	userRole, _ := c.Get("role") // "admin" atau "customer"

	// 2. Cari data booking beserta relasinya
	var booking models.Booking
	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	// 3. Validasi Kepemilikan (Admin bebas, customer hanya bisa milik sendiri)
	if userRole != "admin" && booking.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak memiliki izin untuk membatalkan booking ini"})
		return
	}

	// 4. Cek status booking
	if booking.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking ini sudah dibatalkan sebelumnya"})
		return
	}
	if booking.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking yang sudah selesai tidak dapat dibatalkan"})
		return
	}

	// 5. Validasi Batasan Waktu (Minimal pembatalan 1 jam sebelum jadwal)
	loc := time.FixedZone("WIB", 7*3600)
	bookingTimeWIB := booking.BookingTime.In(loc)
	nowWIB := time.Now().In(loc)

	if nowWIB.After(bookingTimeWIB.Add(-1 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pembatalan ditolak: Batas waktu pembatalan adalah minimal 1 jam sebelum jadwal booking.",
		})
		return
	}

	// 6. Update status menjadi 'cancelled' (Alih-alih menghapus data permanen)
	booking.Status = "cancelled"
	if err := config.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membatalkan booking: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Booking berhasil dibatalkan",
		"data":    booking,
	})
}

// GET /api/my-bookings - Ambil daftar booking milik user yang sedang login
func GetMyBookings(c *gin.Context) {
	// 1. Ambil user_id dari klaim JWT yang sedang aktif
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}

	// 2. Cari semua booking yang user_id-nya sesuai dengan user yang login
	var bookings []models.Booking
	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").Where("user_id = ?", userID).Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data booking: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil riwayat booking Anda",
		"data":    bookings,
	})
}

// PUT /api/bookings/:id/status - Admin mengubah status booking (misal: confirmed, completed)
func UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id")

	// Struct khusus untuk menangkap input status baru
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid: " + err.Error()})
		return
	}

	// Validasi status yang diperbolehkan
	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"completed": true,
		"cancelled": true,
	}

	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status tidak valid. Status yang diizinkan: pending, confirmed, completed, cancelled",
		})
		return
	}

	// Cari data booking berdasarkan ID
	var booking models.Booking
	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	// Update status booking
	booking.Status = input.Status
	if err := config.DB.Save(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui status booking: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Status booking berhasil diperbarui menjadi '" + input.Status + "'",
		"data":    booking,
	})
}

// GET /api/barbers/:id/slots?date=2026-09-07 - Cek slot jam yang tersedia untuk barber pada tanggal tertentu
func GetAvailableSlots(c *gin.Context) {
	barberID := c.Param("id")
	dateStr := c.Query("date") // Format: YYYY-MM-DD

	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter date wajib diisi (Format: YYYY-MM-DD)"})
		return
	}

	// 1. Definisikan jam operasional toko (09:00 sampai 21:00)
	operationalHours := []string{
		"09:00", "10:00", "11:00", "12:00", "13:00", "14:00",
		"15:00", "16:00", "17:00", "18:00", "19:00", "20:00",
	}

	// 2. Ambil semua booking yang aktif (tidak cancelled) untuk barber dan tanggal tersebut
	var existingBookings []models.Booking
	queryDateStart := dateStr + " 00:00:00"
	queryDateEnd := dateStr + " 23:59:59"

	if err := config.DB.Where("barber_id = ? AND status != ? AND booking_time BETWEEN ? AND ?", 
		barberID, "cancelled", queryDateStart, queryDateEnd).Find(&existingBookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengecek ketersediaan slot"})
		return
	}

	// Buat map untuk menampung jam yang sudah dibooking
	bookedTimes := make(map[string]bool)
	for _, b := range existingBookings {
		tStr := b.BookingTime.Format("15:00")
		bookedTimes[tStr] = true
	}

	// 3. Petakan jam operasional ke status ketersediaan
	type Slot struct {
		Time      string `json:"time"`
		Available bool   `json:"available"`
	}

	var slots []Slot
	for _, hour := range operationalHours {
		isBooked := bookedTimes[hour]
		slots = append(slots, Slot{
			Time:      hour,
			Available: !isBooked,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Berhasil mengambil ketersediaan slot",
		"date":      dateStr,
		"barber_id": barberID,
		"data":      slots,
	})
}