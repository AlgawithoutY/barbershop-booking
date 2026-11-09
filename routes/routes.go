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
		AllowAllOrigins:  true, // Diizinkan dari IP laptop mana saja di jaringan lokal
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// RateLimiter Middleware
	r.Use(middleware.RateLimiterMiddleware())

	// --- STATIC FOLDER UPLOADS (Agar foto barber bisa diakses publik) ---
	r.Static("/uploads", "./uploads")

	// --- ROUTE AUTH (Publik) ---
	r.POST("/api/register", controllers.Register)
	r.POST("/api/login", controllers.Login)
	r.POST("/api/send-otp", controllers.SendOTP)
	r.POST("/api/verify-otp", controllers.VerifyOTP)
	r.POST("/api/forgot-password", controllers.ForgotPassword)
	r.POST("/api/reset-password", controllers.ResetPassword)

	// Endpoint Ping (Health Check)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong! API Barbershop siap digunakan.",
		})
	})

	// --- ROUTES PUBLIK (Daftar Layanan, Barber, Jadwal, Slot Jam, & Review Barber Publik) ---
	r.GET("/api/barbers", controllers.GetBarbers)
	r.GET("/api/barbers/:id", controllers.GetBarberByID)
	r.GET("/api/barbers/:id/slots", controllers.GetAvailableSlots)
	r.GET("/api/barbers/:id/schedules", controllers.GetBarberSchedules)
	r.GET("/api/barbers/:id/leaves", controllers.GetBarberLeaves)
	r.GET("/api/barbers/:id/reviews", controllers.GetBarberReviews)

	r.GET("/api/services", controllers.GetServices)
	r.GET("/api/services/:id", controllers.GetServiceByID)

	// Pengaturan Barbershop Publik (Bisa diakses tanpa login)
	r.GET("/api/settings", controllers.GetSettings)

	// --- GROUPING ROUTE API YANG DIPROTEKSI JWT ---
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// --- Action CRUD Barber & Pengaturan Jadwal/Cuti (Khusus Admin) ---
		adminBarber := api.Group("/")
		adminBarber.Use(middleware.RequireRoles("admin"))
		{
			adminBarber.POST("/barbers", controllers.CreateBarber)
			adminBarber.PUT("/barbers/:id", controllers.UpdateBarber)
			adminBarber.DELETE("/barbers/:id", controllers.DeleteBarber)
			
			// Endpoint Upload Foto Profil / Avatar Barber (Khusus Admin)
			adminBarber.PUT("/barbers/:id/avatar", controllers.UploadBarberAvatar)

			// Endpoint Kelola Jadwal Rutin & Cuti Barber oleh Admin
			adminBarber.POST("/barbers/schedules", controllers.SaveBarberSchedule)
			adminBarber.POST("/barbers/leaves", controllers.CreateBarberLeave)
			adminBarber.DELETE("/barbers/leaves/:id", controllers.DeleteBarberLeave)
		}

		// --- Action CRUD Service (Khusus Admin) ---
		adminService := api.Group("/")
		adminService.Use(middleware.RequireRoles("admin"))
		{
			adminService.POST("/services", controllers.CreateService)
			adminService.PUT("/services/:id", controllers.UpdateService)
			adminService.DELETE("/services/:id", controllers.DeleteService)
		}

		// --- Routes Booking (Pemesanan - Butuh Auth Customer/Admin) ---
		api.GET("/bookings", controllers.GetBookings)
		
		// RUTE DENGAN ALIAS YANG DIBERSIHKAN
		api.GET("/bookings/my", controllers.GetMyBookings)
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

		// --- Routes Review Customer ---
		api.POST("/reviews", controllers.CreateReview)

		// --- Routes Admin Reports, Settings, Customers, Reviews & Dashboard Stats (Hanya Admin) ---
		adminOnly := api.Group("/")
		adminOnly.Use(middleware.RequireRoles("admin"))
		{
			// Dashboard Stats
			adminOnly.GET("/admin/dashboard-stats", controllers.GetDashboardStats)

			// Data Pelanggan Terdaftar
			adminOnly.GET("/admin/customers", controllers.GetCustomers)

			// Endpoint Notifikasi & List Booking Admin
			adminOnly.GET("/admin/bookings", controllers.GetBookings)

			// Endpoint Kelola Ulasan Admin
			adminOnly.GET("/admin/reviews", controllers.GetAllReviewsForAdmin)

			// Laporan Keuangan
			adminOnly.GET("/admin/reports/financial", controllers.GetFinancialReport)

			// Pengaturan Barbershop Admin (GET & PUT)
			adminOnly.GET("/admin/settings", controllers.GetSettings)
			adminOnly.PUT("/admin/settings", controllers.UpdateSettings)
		}
	}

	return r
}