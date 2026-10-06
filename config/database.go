package config

import (
	"barbershop-backend/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment variable default")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal terkoneksi ke database: ", err)
	}

	// Auto Migrate Tabel
	err = database.AutoMigrate(
		&models.User{},
		&models.Barber{},
		&models.Service{},
		&models.Booking{},
		&models.Review{}, // <-- Model Review sudah ditambahkan di sini
	)
	if err != nil {
		log.Fatal("Gagal melakukan auto migrate: ", err)
	}

	DB = database
	fmt.Println("Koneksi ke database MariaDB & Auto Migrate berhasil!")
}