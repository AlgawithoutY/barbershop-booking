package models

import (
	"time"

	"gorm.io/gorm"
)

// 1. Model User (Admin / Customer)
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	Email        string         `gorm:"type:varchar(100);unique;not null" json:"email"`
	Phone        string         `gorm:"type:varchar(20);unique;not null" json:"phone"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"` // Hidden dari JSON response
	Role         string         `gorm:"type:enum('admin', 'customer');default:'customer'" json:"role"`
	OtpCode      string         `gorm:"type:varchar(6)" json:"-"`            // Kode OTP sementara
	OtpExpiresAt *time.Time     `json:"-"`                                   // Masa berlaku OTP
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// 2. Model Barber (Tukang Cukur)
type Barber struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Specialty string         `gorm:"type:varchar(100)" json:"specialty"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 3. Model Service (Layanan / Paket Cukur)
type Service struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	DurationMin int            `gorm:"not null" json:"duration_min"` // durasi menit
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// 4. Model Booking (Reservasi)
type Booking struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user"`
	BarberID    uint           `gorm:"not null" json:"barber_id"`
	Barber      Barber         `gorm:"foreignKey:BarberID" json:"barber"`
	ServiceID   uint           `gorm:"not null" json:"service_id"`
	Service     Service        `gorm:"foreignKey:ServiceID" json:"service"`
	BookingTime time.Time      `gorm:"not null" json:"booking_time"`
	Status      string         `gorm:"type:enum('pending', 'confirmed', 'completed', 'cancelled');default:'pending'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// 5. Model Review (Ulasan Pelanggan)
type Review struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BookingID uint           `gorm:"not null;unique" json:"booking_id"` // Relasi ke Booking agar 1 booking 1 review
	Booking   Booking        `gorm:"foreignKey:BookingID" json:"booking"`
	UserID    uint           `gorm:"not null" json:"user_id"`           // Pelanggan yang memberi review
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	Rating    int            `gorm:"not null" json:"rating"`            // Nilai bintang (misal: 1 - 5)
	Comment   string         `gorm:"type:text" json:"comment"`          // Komentar ulasan
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}