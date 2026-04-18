package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sport-reservation/config"
	"sport-reservation/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UploadBukti(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	reservasiIDStr := c.PostForm("reservasi_id")
	reservasiID, _ := strconv.Atoi(reservasiIDStr)

	noWhatsapp := c.PostForm("no_whatsapp")
	if noWhatsapp == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor WhatsApp wajib diisi"})
		return
	}

	var reservasi models.Reservasi
	if err := config.DB.First(&reservasi, reservasiID).Error; err != nil || reservasi.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Reservasi tidak ditemukan atau bukan milik Anda"})
		return
	}

	if reservasi.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Hanya reservasi berstatus pending yang boleh upload bukti"})
		return
	}

	// Cek apakah sudah ada pembayaran aktif
	var existing models.Pembayaran
	if err := config.DB.Where("reservasi_id = ?", reservasiID).First(&existing).Error; err == nil {
		if existing.Status == "verifikasi" || existing.Status == "sudah_bayar" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Sudah ada bukti pembayaran yang sedang diproses"})
			return
		}
	}

	file, err := c.FormFile("bukti")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File bukti wajib diupload"})
		return
	}

	uploadDir := "uploads"
	os.MkdirAll(uploadDir, os.ModePerm)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	dst := filepath.Join(uploadDir, filename)
	c.SaveUploadedFile(file, dst)

	urlPath := uploadDir + "/" + filename

	pembayaran := models.Pembayaran{
		ReservasiID:   uint(reservasiID),
		BuktiTransfer: urlPath,
		NoWhatsapp:    noWhatsapp,
		Status:        "verifikasi",
	}
	config.DB.Create(&pembayaran)

	c.JSON(http.StatusOK, gin.H{"message": "Bukti pembayaran berhasil diupload, menunggu verifikasi admin"})
}

// CancelBukti tetap sama (sudah bagus)
func CancelBukti(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	pembayaranID := c.Param("id")

	var pembayaran models.Pembayaran
	if err := config.DB.Preload("Reservasi").First(&pembayaran, pembayaranID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bukti tidak ditemukan"})
		return
	}

	if pembayaran.Reservasi.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bukan milik Anda"})
		return
	}

	if pembayaran.Status != "verifikasi" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak dapat membatalkan bukti yang sudah diverifikasi"})
		return
	}

	config.DB.Delete(&pembayaran)
	c.JSON(http.StatusOK, gin.H{"message": "Pengiriman bukti berhasil dibatalkan"})
}

func VerifikasiPembayaran(c *gin.Context) {
	var input struct {
		PembayaranID uint `json:"pembayaran_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var pembayaran models.Pembayaran
	if err := config.DB.First(&pembayaran, input.PembayaranID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pembayaran tidak ditemukan"})
		return
	}

	config.DB.Transaction(func(tx *gorm.DB) error {
		tx.Model(&pembayaran).Update("status", "sudah_bayar")
		tx.Model(&models.Reservasi{}).Where("id = ?", pembayaran.ReservasiID).Update("status", "lunas")
		return nil
	})

	c.JSON(http.StatusOK, gin.H{"message": "Pembayaran berhasil diverifikasi"})
}
