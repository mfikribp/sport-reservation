package controllers

import (
	"net/http"
	"sport-reservation/config"
	"sport-reservation/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GET /api/admin/pembayaran — semua pembayaran
func GetAllPembayaran(c *gin.Context) {
	var pembayarans []models.Pembayaran
	config.DB.
		Preload("Reservasi").
		Preload("Reservasi.User").
		Preload("Reservasi.Jadwal").
		Preload("Reservasi.Jadwal.Lapangan").
		Order("created_at DESC").
		Find(&pembayarans)
	c.JSON(http.StatusOK, gin.H{"data": pembayarans})
}

// GET /api/admin/reservasi — semua reservasi
func GetAllReservasi(c *gin.Context) {
	var reservasis []models.Reservasi
	config.DB.
		Preload("User").
		Preload("Jadwal").
		Preload("Jadwal.Lapangan").
		Order("created_at DESC").
		Find(&reservasis)
	c.JSON(http.StatusOK, gin.H{"data": reservasis})
}

// GET /api/admin/users — semua user terdaftar
func GetAllUsers(c *gin.Context) {
	var users []models.User
	config.DB.Order("created_at DESC").Find(&users)
	c.JSON(http.StatusOK, gin.H{"data": users})
}

// PUT /api/admin/users/:id/role — ubah role user (user ↔ admin)
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Role != "user" && body.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role tidak valid. Gunakan 'user' atau 'admin'"})
		return
	}
	result := config.DB.Model(&models.User{}).Where("id = ?", id).Update("role", body.Role)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Role user berhasil diperbarui"})
}

// GET /api/admin/stats — ringkasan statistik
func GetDashboardStats(c *gin.Context) {
	var totalReservasi, totalLapangan, totalUser int64
	var pendingPembayaran int64

	config.DB.Model(&models.Reservasi{}).Count(&totalReservasi)
	config.DB.Model(&models.Lapangan{}).Where("status = ?", "active").Count(&totalLapangan)
	config.DB.Model(&models.User{}).Where("role = ?", "user").Count(&totalUser)
	config.DB.Model(&models.Pembayaran{}).Where("status = ?", "verifikasi").Count(&pendingPembayaran)

	// Revenue bulan ini
	var revenue struct{ Total float64 }
	config.DB.Raw(`
		SELECT COALESCE(SUM(j.harga_per_jam * TIMESTAMPDIFF(HOUR, CONCAT(jd.tanggal, ' ', jd.jam_mulai), CONCAT(jd.tanggal, ' ', jd.jam_selesai))), 0) as total
		FROM reservasis r
		JOIN jadwals jd ON r.jadwal_id = jd.id
		JOIN lapangans j ON jd.lapangan_id = j.id
		WHERE r.status = 'lunas'
		AND MONTH(r.created_at) = MONTH(NOW())
		AND YEAR(r.created_at) = YEAR(NOW())
	`).Scan(&revenue)

	c.JSON(http.StatusOK, gin.H{
		"total_reservasi":    totalReservasi,
		"lapangan_aktif":     totalLapangan,
		"total_user":         totalUser,
		"pending_pembayaran": pendingPembayaran,
		"revenue_bulan_ini":  revenue.Total,
	})
}

// PUT /api/admin/pembayaran/:id/tolak — tolak pembayaran, user bisa upload ulang
func TolakPembayaran(c *gin.Context) {
	id := c.Param("id")

	var pembayaran models.Pembayaran
	if err := config.DB.First(&pembayaran, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pembayaran tidak ditemukan"})
		return
	}

	config.DB.Transaction(func(tx *gorm.DB) error {
		tx.Model(&pembayaran).Update("status", "ditolak")
		// Kembalikan status reservasi ke pending agar user bisa upload ulang bukti
		tx.Model(&models.Reservasi{}).Where("id = ?", pembayaran.ReservasiID).Update("status", "pending")
		return nil
	})

	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran ditolak. User dapat mengupload ulang bukti pembayaran."})
}
