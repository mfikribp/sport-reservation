package routes

import (
	"sport-reservation/controllers"
	"sport-reservation/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Public Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// Public
		api.GET("/lapangan", controllers.GetLapangan)
		api.GET("/lapangan/:id", controllers.GetLapanganByID)
		api.GET("/jadwal", controllers.GetJadwal)

		// Protected User
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/reservasi", controllers.CreateReservasi)
			protected.GET("/reservasi/user", controllers.GetReservasiUser)
			protected.POST("/upload-bukti", controllers.UploadBukti)
			protected.PUT("/reservasi/:id/cancel", controllers.CancelReservasi)
			protected.DELETE("/pembayaran/:id/cancel", controllers.CancelBukti)
		}

		// Protected Admin
		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware())
		adminGroup.Use(middleware.AdminOnly())
		{
			// Dashboard stats
			adminGroup.GET("/stats", controllers.GetDashboardStats)

			// Lapangan management
			adminGroup.POST("/lapangan", controllers.CreateLapangan)
			adminGroup.PUT("/lapangan/:id", controllers.UpdateLapangan)
			adminGroup.DELETE("/lapangan/:id", controllers.DeleteLapangan)
			adminGroup.PATCH("/lapangan/:id/toggle-status", controllers.ToggleStatusLapangan)

			// Jadwal management
			adminGroup.GET("/jadwal", controllers.GetAllJadwalAdmin)
			adminGroup.POST("/jadwal", controllers.CreateJadwal)
			adminGroup.POST("/jadwal/bulk", controllers.CreateJadwalBulk)
			adminGroup.PUT("/jadwal/:id", controllers.UpdateJadwal)
			adminGroup.DELETE("/jadwal/:id", controllers.DeleteJadwal)

			// Pembayaran
			adminGroup.POST("/verifikasi", controllers.VerifikasiPembayaran)
			adminGroup.PUT("/verifikasi/:id/tolak", controllers.TolakPembayaran)
			adminGroup.GET("/pembayaran", controllers.GetAllPembayaran)

			// Reservasi
			adminGroup.GET("/reservasi", controllers.GetAllReservasi)
			adminGroup.PUT("/reservasi/:id/cancel", controllers.AdminCancelReservasi)

			// User management
			adminGroup.GET("/users", controllers.GetAllUsers)
			adminGroup.PUT("/users/:id/role", controllers.UpdateUserRole)
		}
	}
}
