package routes

import (
	"time"

	"barbershop-backend/controllers"
	"barbershop-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// --- PASANG CORS FLEXIBLE ---
	r.Use(cors.New(cors.Config{
		// Mengizinkan localhost & 127.0.0.1 agar tidak ERR_CONNECTION_REFUSED
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Opsional: Komentari dulu RateLimiter jika masih terkendala
	r.Use(middleware.RateLimiterMiddleware())

	// --- ROUTE AUTH (Publik) ---
	r.POST("/api/register", controllers.Register)
	r.POST("/api/login", controllers.Login)
	r.POST("/api/send-otp", controllers.SendOTP)
	r.POST("/api/verify-otp", controllers.VerifyOTP)
	r.POST("/api/forgot-password", controllers.ForgotPassword)

	// Endpoint Ping (Health Check)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong! API Barbershop siap digunakan.",
		})
	})

	// --- ROUTES PUBLIK ---
	r.GET("/api/barbers/:id/slots", controllers.GetAvailableSlots)

	// Grouping Route API yang Diproteksi JWT
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// --- Routes Barber ---
		api.GET("/barbers", controllers.GetBarbers)
		api.GET("/barbers/:id", controllers.GetBarberByID)

		adminBarber := api.Group("/")
		adminBarber.Use(middleware.RequireRoles("admin"))
		{
			adminBarber.POST("/barbers", controllers.CreateBarber)
			adminBarber.PUT("/barbers/:id", controllers.UpdateBarber)
			adminBarber.DELETE("/barbers/:id", controllers.DeleteBarber)
		}

		// --- Routes Service (Layanan) ---
		api.GET("/services", controllers.GetServices)
		api.GET("/services/:id", controllers.GetServiceByID)

		adminService := api.Group("/")
		adminService.Use(middleware.RequireRoles("admin"))
		{
			adminService.POST("/services", controllers.CreateService)
			adminService.PUT("/services/:id", controllers.UpdateService)
			adminService.DELETE("/services/:id", controllers.DeleteService)
		}

		// --- Routes Booking (Pemesanan) ---
		api.GET("/bookings", controllers.GetBookings)
		api.GET("/my-bookings", controllers.GetMyBookings)
		api.POST("/bookings", controllers.CreateBooking)
		api.GET("/bookings/:id", controllers.GetBookingByID)
		api.PUT("/bookings/:id", controllers.UpdateBooking)
		api.DELETE("/bookings/:id", controllers.DeleteBooking)
		api.PUT("/bookings/:id/cancel", controllers.CancelBooking)

		adminBooking := api.Group("/")
		adminBooking.Use(middleware.RequireRoles("admin"))
		{
			adminBooking.PUT("/bookings/:id/status", controllers.UpdateBookingStatus)
		}

		// --- Routes Review (Ulasan Pelanggan) ---
		api.GET("/reviews", controllers.GetReviews)
		api.POST("/reviews", controllers.CreateReview)

		// --- Routes Admin Reports (Laporan Keuangan) ---
		api.GET("/admin/reports/financial", middleware.RequireRoles("admin"), controllers.GetFinancialReport)
	}

	return r
}