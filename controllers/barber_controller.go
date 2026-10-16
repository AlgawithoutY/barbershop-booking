package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"barbershop-backend/config"
	"barbershop-backend/models"

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

// PUT /api/barbers/:id - Update data barber, status aktif, hari & jam kerja
func UpdateBarber(c *gin.Context) {
	var barber models.Barber
	id := c.Param("id")

	if err := config.DB.First(&barber, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}

	var input struct {
		Name             string `json:"name"`
		Specialty        string `json:"specialty"`
		IsActive         *bool  `json:"is_active"`
		WorkDays         string `json:"work_days"`
		OperationalHours string `json:"operational_hours"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update field jika disediakan
	if input.Name != "" {
		barber.Name = input.Name
	}
	if input.Specialty != "" {
		barber.Specialty = input.Specialty
	}
	if input.WorkDays != "" {
		barber.WorkDays = input.WorkDays
	}
	if input.OperationalHours != "" {
		barber.OperationalHours = input.OperationalHours
	}
	
	isOff := false
	if input.IsActive != nil {
		barber.IsActive = *input.IsActive
		isOff = !barber.IsActive // Jika is_active = false, maka is_off = true (Libur)
	}

	if err := config.DB.Save(&barber).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengupdate data barber"})
		return
	}

	// Otomatis sinkronkan jadwal rutin mingguan (0-6) agar ikut libur/masuk
	if input.IsActive != nil {
		for day := 0; day <= 6; day++ {
			var schedule models.BarberSchedule
			err := config.DB.Where("barber_id = ? AND day_of_week = ?", barber.ID, day).First(&schedule).Error
			if err != nil {
				// Buat baru jika belum ada
				newSch := models.BarberSchedule{
					BarberID:  barber.ID,
					DayOfWeek: day,
					IsOff:     isOff,
				}
				config.DB.Create(&newSch)
			} else {
				// Update jika sudah ada
				schedule.IsOff = isOff
				config.DB.Save(&schedule)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data barber berhasil diupdate", "data": barber})
}

// PUT /api/barbers/:id/avatar - Upload atau update foto profil barber
func UploadBarberAvatar(c *gin.Context) {
	id := c.Param("id")

	var barber models.Barber
	if err := config.DB.First(&barber, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Barber tidak ditemukan"})
		return
	}

	// Tangkap file dari form-data dengan key "avatar"
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File foto wajib diunggah"})
		return
	}

	// Validasi ekstensi file
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file harus .jpg, .jpeg, atau .png"})
		return
	}

	// Buat nama file unik
	filename := fmt.Sprintf("barber_%s_%d%s", id, time.Now().Unix(), ext)
	dst := "./uploads/" + filename

	// Simpan file ke folder lokal server
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan file foto ke server"})
		return
	}

	// Update path avatar di database
	barber.Avatar = "/uploads/" + filename
	if err := config.DB.Save(&barber).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui database barber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Foto profil barber berhasil diunggah",
		"data":    barber,
	})
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

// ==========================================
// MANAJEMEN JADWAL & CUTI BARBER (ADMIN)
// ==========================================

// GET /api/barbers/:id/schedules - Ambil jadwal rutin mingguan barber
func GetBarberSchedules(c *gin.Context) {
	barberID := c.Param("id")
	var schedules []models.BarberSchedule

	if err := config.DB.Where("barber_id = ?", barberID).Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil jadwal barber"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

// POST /api/barbers/schedules - Simpan/Update jadwal rutin mingguan
func SaveBarberSchedule(c *gin.Context) {
	var input struct {
		BarberID  uint `json:"barber_id" binding:"required"`
		DayOfWeek int  `json:"day_of_week"` // 0 = Minggu, 1 = Senin, ..., 6 = Sabtu
		IsOff     bool `json:"is_off"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid"})
		return
	}

	var schedule models.BarberSchedule
	err := config.DB.Where("barber_id = ? AND day_of_week = ?", input.BarberID, input.DayOfWeek).First(&schedule).Error

	if err != nil {
		// Buat jadwal baru jika belum ada
		newSchedule := models.BarberSchedule{
			BarberID:  input.BarberID,
			DayOfWeek: input.DayOfWeek,
			IsOff:     input.IsOff,
		}
		if err := config.DB.Create(&newSchedule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan jadwal"})
			return
		}
	} else {
		// Update status libur jika sudah ada
		schedule.IsOff = input.IsOff
		config.DB.Save(&schedule)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Jadwal barber berhasil diperbarui"})
}

// GET /api/barbers/:id/leaves - Ambil daftar cuti barber
func GetBarberLeaves(c *gin.Context) {
	barberID := c.Param("id")
	var leaves []models.BarberLeave

	if err := config.DB.Where("barber_id = ?", barberID).Order("leave_date DESC").Find(&leaves).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar cuti"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": leaves})
}

// POST /api/barbers/leaves - Tambah data cuti/libur spesifik tanggal
func CreateBarberLeave(c *gin.Context) {
	var input struct {
		BarberID  uint   `json:"barber_id" binding:"required"`
		LeaveDate string `json:"leave_date" binding:"required"` // Format: YYYY-MM-DD
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input tidak valid"})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", input.LeaveDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format tanggal salah. Gunakan YYYY-MM-DD"})
		return
	}

	leave := models.BarberLeave{
		BarberID:  input.BarberID,
		LeaveDate: parsedDate,
		Reason:    input.Reason,
	}

	if err := config.DB.Create(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencatat cuti barber"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Cuti barber berhasil ditambahkan", "data": leave})
}

// DELETE /api/barbers/leaves/:id - Hapus status cuti
func DeleteBarberLeave(c *gin.Context) {
	id := c.Param("id")
	var leave models.BarberLeave

	if err := config.DB.First(&leave, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data cuti tidak ditemukan"})
		return
	}

	config.DB.Delete(&leave)
	c.JSON(http.StatusOK, gin.H{"message": "Status cuti berhasil dihapus"})
}