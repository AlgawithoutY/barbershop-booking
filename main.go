package main

import (
	"barbershop-backend/config"
	"barbershop-backend/routes"
	"fmt"
	"os"
)

func main() {
	// 1. Inisialisasi Database
	config.ConnectDatabase()

	// 2. Setup Router
	r := routes.SetupRouter()

	// 3. Jalankan Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server running di http://localhost:%s\n", port)
	r.Run(":" + port)
}