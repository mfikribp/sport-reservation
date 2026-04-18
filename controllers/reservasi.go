package controllers

import (
	"net/http"
	"os"
	"sport-reservation/config"
	"sport-reservation/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookingInput struct {
	JadwalID uint `json:"jadwal_id" binding:"required"`
}

func CreateReservasi(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var input BookingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Cek jadwal tersedia
		var jadwal models.Jadwal
		if err := tx.First(&jadwal, input.JadwalID).Error; err != nil {
			return err
		}
		if jadwal.Status != "tersedia" {
			return gorm.ErrRecordNotFound
		}

		// Anti double-booking
		var count int64
		tx.Model(&models.Reservasi{}).Where("jadwal_id = ? AND status NOT IN ('expired','batal')", input.JadwalID).Count(&count)
		if count > 0 {
			return gorm.ErrDuplicatedKey
		}

		// Buat reservasi
		reservasi := models.Reservasi{
			UserID:    userID,
			JadwalID:  input.JadwalID,
			Status:    "pending",
			ExpiredAt: time.Now().Add(15 * time.Minute),
		}
		if err := tx.Create(&reservasi).Error; err != nil {
			return err
		}

		// Lock jadwal
		tx.Model(&jadwal).Update("status", "dipesan")
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jadwal tidak tersedia atau sudah dipesan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Reservasi berhasil dibuat (pending 15 menit)"})
}

func GetReservasiUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var reservasi []models.Reservasi
	config.DB.
		Preload("Jadwal").
		Preload("Jadwal.Lapangan").
		Preload("Pembayaran").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reservasi)
	c.JSON(http.StatusOK, gin.H{"data": reservasi})
}

func CancelReservasi(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	reservasiID := c.Param("id")

	var reservasi models.Reservasi
	var needRefund bool

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&reservasi, reservasiID).Error; err != nil {
			return err
		}

		if reservasi.UserID != userID {
			return gorm.ErrRecordNotFound // Unauthorized
		}

		if reservasi.Status != "pending" && reservasi.Status != "lunas" {
			return gorm.ErrInvalidData // Cannot cancel expired/already cancelled
		}

		needRefund = reservasi.Status == "lunas"

		// Cancel it
		if err := tx.Model(&reservasi).Update("status", "batal").Error; err != nil {
			return err
		}

		// Make jadwal available again
		if err := tx.Model(&models.Jadwal{}).Where("id = ?", reservasi.JadwalID).Update("status", "tersedia").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membatalkan. Pastikan pesanan masih pending/lunas dan milik Anda."})
		return
	}

	if needRefund {
		wa := os.Getenv("ADMIN_WHATSAPP")
		if wa == "" {
			wa = "6281234567890" // default fallback
		}
		c.JSON(http.StatusOK, gin.H{
			"message":    "Pesanan berhasil dibatalkan. Karena sudah lunas, silakan hubungi admin untuk proses refund.",
			"need_refund": true,
			"whatsapp":   wa,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pesanan berhasil dibatalkan", "need_refund": false})
}

// AdminCancelReservasi — admin can cancel any pending or lunas reservation
func AdminCancelReservasi(c *gin.Context) {
	reservasiID := c.Param("id")

	var reservasi models.Reservasi
	var needRefund bool

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&reservasi, reservasiID).Error; err != nil {
			return err
		}

		if reservasi.Status != "pending" && reservasi.Status != "lunas" {
			return gorm.ErrInvalidData
		}

		needRefund = reservasi.Status == "lunas"

		if err := tx.Model(&reservasi).Update("status", "batal").Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Jadwal{}).Where("id = ?", reservasi.JadwalID).Update("status", "tersedia").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membatalkan reservasi."})
		return
	}

	if needRefund {
		wa := os.Getenv("ADMIN_WHATSAPP")
		if wa == "" {
			wa = "6281234567890"
		}
		c.JSON(http.StatusOK, gin.H{
			"message":    "Reservasi dibatalkan oleh admin. Proses refund diperlukan karena sudah lunas.",
			"need_refund": true,
			"whatsapp":   wa,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservasi berhasil dibatalkan oleh admin", "need_refund": false})
}
