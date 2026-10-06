package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"barbershop-backend/config"
	"barbershop-backend/models"
	"barbershop-backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// POST /api/v1/auth/register - Pendaftaran user baru
func Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone" binding:"required,min=10"` // Field Nomor Telepon
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	// Set role default ke 'customer' jika kosong
	role := input.Role
	if role == "" {
		role = "customer"
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau nomor telepon sudah terdaftar"})
		return
	}

	// Sembunyikan password sebelum dikirim sebagai response
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{
		"message": "Registrasi berhasil",
		"data":     user,
	})
}

// POST /api/v1/auth/login - Login user dengan Email & Password
func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	// Bandingkan password input dengan password hash di database
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	// Generate Token JWT
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   token,
		"data":    user,
	})
}

// POST /api/v1/auth/send-otp - Kirim kode OTP ke Nomor Telepon
func SendOTP(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor telepon wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", input.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nomor telepon belum terdaftar"})
		return
	}

	// Generate 6 digit angka acak (contoh: 482910)
	rand.Seed(time.Now().UnixNano())
	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiresAt := time.Now().Add(5 * time.Minute) // Berlaku 5 menit

	// Simpan OTP & Waktu Expiry ke Database
	user.OtpCode = otpCode
	user.OtpExpiresAt = &expiresAt
	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Kode OTP berhasil dikirim!",
		"debug_otp": otpCode, // Ditampilkan di JSON untuk kemudahan testing di Postman
	})
}

// POST /api/v1/auth/verify-otp - Verifikasi OTP & Return JWT Token
func VerifyOTP(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
		Otp   string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor telepon dan OTP wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", input.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Nomor telepon tidak ditemukan"})
		return
	}

	// Cek apakah kode OTP cocok
	if user.OtpCode != input.Otp {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode OTP salah"})
		return
	}

	// Cek apakah OTP sudah kadaluarsa
	if user.OtpExpiresAt == nil || time.Now().After(*user.OtpExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode OTP sudah kadaluarsa"})
		return
	}

	// Reset OTP di database agar tidak bisa digunakan kembali
	user.OtpCode = ""
	user.OtpExpiresAt = nil
	config.DB.Save(&user)

	// Buat Token JWT
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token JWT"})
		return
	}

	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Login OTP berhasil!",
		"token":   token,
		"data":    user,
	})
}

// ForgotPassword menangani permintaan pemulihan kata sandi
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, instruksi pemulihan telah dikirim."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instruksi reset kata sandi telah dikirim ke email."})
}