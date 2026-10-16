package main

import (
	"barbershop-backend/config"
	"barbershop-backend/models"
	"barbershop-backend/routes"
	"fmt"
	"os"
)

func main() {
	// 1. Inisialisasi Database
	config.ConnectDatabase()

	// 2. Jalankan Auto Migrate untuk membuat/memperbarui tabel database
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.Barber{},
		&models.Service{},
		&models.Booking{},
		&models.Review{},
		&models.BarberSchedule{},
		&models.BarberLeave{},
	)

	if err != nil {
		fmt.Println("Gagal melakukan AutoMigrate database:", err)
	} else {
		fmt.Println("AutoMigrate database berhasil!")
	}

	// 3. Setup Router
	r := routes.SetupRouter()

	// 4. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running di http://localhost:%s\n", port)
	r.Run(":" + port)
}