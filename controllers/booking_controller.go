package controllers

import (
	"net/http"
	"time"

	"barbershop-backend/config"
	"barbershop-backend/models"

	"github.com/gin-gonic/gin"
)

// GET /api/bookings - Admin melihat semua booking dengan opsi filter
func GetBookings(c *gin.Context) {
	var bookings []models.Booking
	query := config.DB.Preload("User").Preload("Barber").Preload("Service")

	dateFilter := c.Query("date")
	statusFilter := c.Query("status")
	barberIDFilter := c.Query("barber_id")

	if dateFilter != "" {
		startDate := dateFilter + " 00:00:00"
		endDate := dateFilter + " 23:59:59"
		query = query.Where("booking_time BETWEEN ? AND ?", startDate, endDate)
	}

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	if barberIDFilter != "" {
		query = query.Where("barber_id = ?", barberIDFilter)
	}

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

// POST /api/bookings - Buat booking baru dengan validasi operasional, status barber, cuti, & bentrok
func CreateBooking(c *gin.Context) {
	var input models.Booking
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}
	input.UserID = userID.(uint)

	// Set Timezone WIB (+7)
	loc := time.FixedZone("WIB", 7*3600)
	nowWIB := time.Now().In(loc)
	bookingWIB := input.BookingTime.In(loc)

	// 1. Cek Status Aktif Barber
	var barber models.Barber
	if err := config.DB.First(&barber, input.BarberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}
	if !barber.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Barber sedang tidak aktif atau sedang libur.",
		})
		return
	}

	// 2. Validasi Jam Operasional (08:00 - 22:00 WIB)
	hour := bookingWIB.Hour()
	if hour < 8 || hour >= 22 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Di luar jam operasional toko (08:00 - 22:00 WIB)",
		})
		return
	}

	// 3. Validasi Tanggal Masa Lalu (Perbandingan Waktu WIB)
	if bookingWIB.Before(nowWIB) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Tidak dapat memesan untuk waktu yang sudah berlalu",
		})
		return
	}

	// 4. Validasi Cuti Spesifik Barber
	dateOnly := bookingWIB.Format("2006-01-02")
	var leaveCount int64
	config.DB.Model(&models.BarberLeave{}).
		Where("barber_id = ? AND leave_date = ?", input.BarberID, dateOnly).
		Count(&leaveCount)

	if leaveCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Barber sedang mengambil cuti pada tanggal tersebut.",
		})
		return
	}

	// 5. Validasi Hari Libur Rutin Barber
	dayOfWeek := int(bookingWIB.Weekday())
	var schedule models.BarberSchedule
	errSch := config.DB.Where("barber_id = ? AND day_of_week = ?", input.BarberID, dayOfWeek).First(&schedule).Error
	if errSch == nil && schedule.IsOff {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Booking ditolak: Barber libur rutin pada hari tersebut.",
		})
		return
	}

	// 6. Cek Double-Booking
	var existingBooking models.Booking
	err := config.DB.Where("barber_id = ? AND booking_time = ? AND status != ?", input.BarberID, input.BookingTime, "cancelled").First(&existingBooking).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Jadwal bentrok: Barber ini sudah memiliki jadwal booking di waktu tersebut.",
		})
		return
	}

	// Default status booking baru
	input.Status = "pending"

	// 7. Simpan Booking Ke Database
	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat booking: " + err.Error()})
		return
	}

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

// PUT /api/bookings/:id - Update data booking
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

// DELETE /api/bookings/:id - Hapus booking
func DeleteBooking(c *gin.Context) {
	var booking models.Booking
	id := c.Param("id")

	if err := config.DB.First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	config.DB.Delete(&booking)
	c.JSON(http.StatusOK, gin.H{"message": "Booking berhasil dihapus"})
}

// PUT /api/bookings/:id/cancel - Pembatalan booking oleh pelanggan/admin
func CancelBooking(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}

	userRole, _ := c.Get("role")

	var booking models.Booking
	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	if userRole != "admin" && booking.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak: Anda tidak memiliki izin membatalkan booking ini"})
		return
	}

	if booking.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking ini sudah dibatalkan sebelumnya"})
		return
	}
	if booking.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking yang sudah selesai tidak dapat dibatalkan"})
		return
	}

	loc := time.FixedZone("WIB", 7*3600)
	bookingTimeWIB := booking.BookingTime.In(loc)
	nowWIB := time.Now().In(loc)

	if nowWIB.After(bookingTimeWIB.Add(-1 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Pembatalan ditolak: Batas waktu pembatalan minimal 1 jam sebelum jadwal.",
		})
		return
	}

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

// GET /api/my-bookings - Ambil riwayat booking user
func GetMyBookings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Sesi login tidak valid"})
		return
	}

	var bookings []models.Booking
	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").Where("user_id = ?", userID).Order("booking_time DESC").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data booking: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil riwayat booking Anda",
		"data":    bookings,
	})
}

// PUT /api/bookings/:id/status - Update status booking oleh Admin
func UpdateBookingStatus(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid: " + err.Error()})
		return
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"confirmed": true,
		"completed": true,
		"cancelled": true,
	}

	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status tidak valid. Gunakan: pending, confirmed, completed, cancelled",
		})
		return
	}

	var booking models.Booking
	if err := config.DB.Preload("User").Preload("Barber").Preload("Service").First(&booking, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

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

// GET /api/barbers/:id/slots?date=YYYY-MM-DD - Cek ketersediaan slot jam berdasarkan status barber, jadwal, & cuti
func GetAvailableSlots(c *gin.Context) {
	barberID := c.Param("id")
	dateStr := c.Query("date")

	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter date wajib diisi (Format: YYYY-MM-DD)"})
		return
	}

	// 1. Cek Status Aktif/Libur Barber secara Utama (Kolom IsActive di tabel barbers)
	var barber models.Barber
	if err := config.DB.First(&barber, barberID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}

	if !barber.IsActive {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Barber sedang tidak aktif / libur",
			"is_off":    true,
			"date":      dateStr,
			"barber_id": barberID,
			"data":      []gin.H{},
		})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal tidak valid. Gunakan YYYY-MM-DD"})
		return
	}

	// 2. Cek Cuti Spesifik Barber pada Tanggal Tersebut
	var leaveCount int64
	config.DB.Model(&models.BarberLeave{}).
		Where("barber_id = ? AND leave_date = ?", barberID, dateStr).
		Count(&leaveCount)

	if leaveCount > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Barber sedang cuti/libur pada tanggal ini",
			"is_off":    true,
			"date":      dateStr,
			"barber_id": barberID,
			"data":      []gin.H{},
		})
		return
	}

	// 3. Cek Hari Libur Rutin Mingguan Barber
	dayOfWeek := int(parsedDate.Weekday())
	var schedule models.BarberSchedule
	errSchedule := config.DB.Where("barber_id = ? AND day_of_week = ?", barberID, dayOfWeek).First(&schedule).Error

	if errSchedule == nil && schedule.IsOff {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Barber libur rutin pada hari ini",
			"is_off":    true,
			"date":      dateStr,
			"barber_id": barberID,
			"data":      []gin.H{},
		})
		return
	}

	// 4. Ambil Slot Jam Operasional (08:00 - 21:00)
	operationalHours := []string{
		"08:00", "09:00", "10:00", "11:00", "12:00", "13:00",
		"14:00", "15:00", "16:00", "17:00", "18:00", "19:00",
		"20:00", "21:00",
	}

	var existingBookings []models.Booking
	queryDateStart := dateStr + " 00:00:00"
	queryDateEnd := dateStr + " 23:59:59"

	if err := config.DB.Where("barber_id = ? AND status != ? AND booking_time BETWEEN ? AND ?",
		barberID, "cancelled", queryDateStart, queryDateEnd).Find(&existingBookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengecek ketersediaan slot"})
		return
	}

	loc := time.FixedZone("WIB", 7*3600)
	bookedTimes := make(map[string]bool)
	for _, b := range existingBookings {
		tStr := b.BookingTime.In(loc).Format("15:04")
		bookedTimes[tStr] = true
	}

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
		"is_off":    false,
		"date":      dateStr,
		"barber_id": barberID,
		"data":      slots,
	})
}