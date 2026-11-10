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

// 2. Model Barber (Tukang Cukur) - Diperbarui untuk model Barber-Centric
type Barber struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Name             string         `gorm:"type:varchar(100);not null" json:"name"`
	Specialty        string         `gorm:"type:varchar(100)" json:"specialty"`
	Description      string         `gorm:"type:text" json:"description"` // Deskripsi/keahlian barber untuk klien
	Rating           float64        `gorm:"type:decimal(3,2);default:0.00" json:"rating"` // Rata-rata rating barber
	IsActive         bool           `gorm:"default:true" json:"is_active"`
	WorkDays         string         `gorm:"type:varchar(100);default:'Senin - Sabtu'" json:"work_days"`
	OperationalHours string         `gorm:"type:varchar(50);default:'08:00 - 22:00 WIB'" json:"operational_hours"`
	Avatar           string         `gorm:"type:varchar(255)" json:"avatar"` // Foto profil barber
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

// 3. Model Service (Opsional / bisa diabaikan jika tidak dipakai lagi)
type Service struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Price       float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	DurationMin int            `gorm:"not null" json:"duration_min"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// 4. Model Booking (Reservasi) - Status diubah ke varchar(20) agar mendukung rejected
type Booking struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user"`
	BarberID    uint           `gorm:"not null" json:"barber_id"`
	Barber      Barber         `gorm:"foreignKey:BarberID" json:"barber"`
	BookingTime time.Time      `gorm:"not null" json:"booking_time"`
	Status      string         `json:"status" gorm:"type:varchar(20);default:'pending'"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// 5. Model Review (Ulasan Pelanggan)
type Review struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BookingID uint           `gorm:"not null;unique" json:"booking_id"` // 1 booking 1 review
	Booking   Booking        `gorm:"foreignKey:BookingID" json:"booking"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"user"`
	BarberID  uint           `gorm:"not null" json:"barber_id"`
	Barber    Barber         `gorm:"foreignKey:BarberID" json:"barber"`
	Rating    int            `gorm:"not null" json:"rating"` // 1 - 5
	Comment   string         `gorm:"type:text" json:"comment"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 6. Model BarberSchedule (Jadwal Kerja Rutin Mingguan Barber)
type BarberSchedule struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BarberID  uint           `gorm:"not null" json:"barber_id"`
	Barber    Barber         `gorm:"foreignKey:BarberID" json:"-"`
	DayOfWeek int            `gorm:"not null" json:"day_of_week"` // 0 = Minggu, 1 = Senin, ..., 6 = Sabtu
	IsOff     bool           `gorm:"default:false" json:"is_off"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 7. Model BarberLeave (Cuti / Libur Spesifik Barber)
type BarberLeave struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	BarberID  uint           `gorm:"not null" json:"barber_id"`
	Barber    Barber         `gorm:"foreignKey:BarberID" json:"-"`
	LeaveDate time.Time      `gorm:"type:date;not null" json:"leave_date"`
	Reason    string         `gorm:"type:varchar(255)" json:"reason"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}