<<<<<<< HEAD
package main

import (
	"log"
	"sport-reservation/config"
	"sport-reservation/models"
	"sport-reservation/routes"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func checkExpired() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now()

		// 1. Expire pending reservations whose payment window has closed
		var reservasi []models.Reservasi
		config.DB.Where("status = 'pending' AND expired_at < ?", now).Find(&reservasi)
		for _, r := range reservasi {
			config.DB.Transaction(func(tx *gorm.DB) error {
				tx.Model(&r).Update("status", "expired")
				tx.Model(&models.Jadwal{}).Where("id = ?", r.JadwalID).Update("status", "tersedia")
				return nil
			})
		}

		// 2. Release jadwal whose scheduled end time has passed (masa sewa habis)
		nowStr := now.Format("2006-01-02 15:04:05")
		config.DB.Model(&models.Jadwal{}).
			Where("status = 'dipesan' AND CONCAT(tanggal, ' ', jam_selesai) < ?", nowStr).
			Update("status", "tersedia")
	}
}

func main() {
	config.LoadEnv()
	config.ConnectDB()

	// Background auto-cancel
	go checkExpired()

	r := gin.Default()

	// Serve Static Frontend
	r.Static("/public", "./public")
	r.Static("/img", "./img")
	r.Static("/uploads", "./uploads")
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})
	r.GET("/admin", func(c *gin.Context) {
		c.File("./public/admin.html")
	})

	routes.SetupRoutes(r)

	port := config.GetEnv("PORT", config.GetEnv("APP_PORT", "8080"))
	log.Printf("Server berjalan di http://localhost:%s", port)
	r.Run(":" + port)
}
=======
package main

import (
	"log"
	"sport-reservation/config"
	"sport-reservation/models"
	"sport-reservation/routes"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func checkExpired() {
	for {
		time.Sleep(1 * time.Minute)
		now := time.Now()

		// 1. Expire pending reservations whose payment window has closed
		var reservasi []models.Reservasi
		config.DB.Where("status = 'pending' AND expired_at < ?", now).Find(&reservasi)
		for _, r := range reservasi {
			config.DB.Transaction(func(tx *gorm.DB) error {
				tx.Model(&r).Update("status", "expired")
				tx.Model(&models.Jadwal{}).Where("id = ?", r.JadwalID).Update("status", "tersedia")
				return nil
			})
		}

		// 2. Release jadwal whose scheduled end time has passed (masa sewa habis)
		nowStr := now.Format("2006-01-02 15:04:05")
		config.DB.Model(&models.Jadwal{}).
			Where("status = 'dipesan' AND CONCAT(tanggal, ' ', jam_selesai) < ?", nowStr).
			Update("status", "tersedia")
	}
}

func main() {
	config.LoadEnv()
	config.ConnectDB()

	// Background auto-cancel
	go checkExpired()

	r := gin.Default()

	// Serve Static Frontend
	r.Static("/public", "./public")
	r.Static("/img", "./img")
	r.Static("/uploads", "./uploads")
	r.GET("/", func(c *gin.Context) {
		c.File("./public/index.html")
	})
	r.GET("/admin", func(c *gin.Context) {
		c.File("./public/admin.html")
	})

	routes.SetupRoutes(r)

	port := config.GetEnv("PORT", config.GetEnv("APP_PORT", "8080"))
	log.Printf("Server berjalan di http://localhost:%s", port)
	r.Run(":" + port)
}
>>>>>>> 7bd88b99f9e4c3b8a5475a073afd7b8759926b98
