package controllers

import (
	"errors"
	"net/http"
	"sport-reservation/config"
	"sport-reservation/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetJadwal(c *gin.Context) {
	lapanganID := c.Query("lapangan_id")
	tanggal := c.Query("tanggal")

	var jadwal []models.Jadwal
	query := config.DB.Preload("Lapangan").Where("CONCAT(tanggal, ' ', jam_selesai) > NOW()")

	if lapanganID != "" {
		query = query.Where("lapangan_id = ?", lapanganID)
	}
	if tanggal != "" {
		query = query.Where("tanggal = ?", tanggal)
	}

	query.Order("tanggal ASC, jam_mulai ASC").Find(&jadwal)
	c.JSON(http.StatusOK, gin.H{"data": jadwal})
}

func GetAllJadwalAdmin(c *gin.Context) {
	lapanganID := c.Query("lapangan_id")
	tanggal := c.Query("tanggal")

	var jadwal []models.Jadwal
	query := config.DB.Preload("Lapangan")

	if lapanganID != "" {
		query = query.Where("lapangan_id = ?", lapanganID)
	}
	if tanggal != "" {
		query = query.Where("tanggal = ?", tanggal)
	}

	query.Order("tanggal DESC, jam_mulai ASC").Find(&jadwal)
	c.JSON(http.StatusOK, gin.H{"data": jadwal})
}

// Helper validasi waktu
func validateTime(j models.Jadwal) error {
	if j.JamMulai >= j.JamSelesai {
		return errors.New("Jam mulai harus lebih kecil dari jam selesai")
	}
	return nil
}

func CreateJadwal(c *gin.Context) {
	var j models.Jadwal
	if err := c.ShouldBindJSON(&j); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if j.Status == "" {
		j.Status = "tersedia"
	}

	if err := validateTime(j); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek overlap jadwal
	var existing models.Jadwal
	err := config.DB.Where(
		"lapangan_id = ? AND tanggal = ? AND status != ? AND "+
			"jam_mulai < ? AND jam_selesai > ?",
		j.LapanganID, j.Tanggal, "expired",
		j.JamSelesai, j.JamMulai,
	).First(&existing).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Jadwal bentrok dengan jadwal lain"})
		return
	}

	config.DB.Create(&j)
	config.DB.Preload("Lapangan").First(&j, j.ID)
	c.JSON(http.StatusCreated, gin.H{"message": "Jadwal berhasil dibuat", "data": j})
}

func UpdateJadwal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var j models.Jadwal
	if err := config.DB.First(&j, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jadwal tidak ditemukan"})
		return
	}
	if j.Status == "dipesan" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jadwal yang sudah dipesan tidak dapat diubah"})
		return
	}

	var input models.Jadwal
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateTime(input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&j).Updates(map[string]interface{}{
		"lapangan_id": input.LapanganID,
		"tanggal":     input.Tanggal,
		"jam_mulai":   input.JamMulai,
		"jam_selesai": input.JamSelesai,
		"status":      input.Status,
	})
	config.DB.Preload("Lapangan").First(&j, id)
	c.JSON(http.StatusOK, gin.H{"message": "Jadwal berhasil diupdate", "data": j})
}

func DeleteJadwal(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var j models.Jadwal
	if err := config.DB.First(&j, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jadwal tidak ditemukan"})
		return
	}
	if j.Status == "dipesan" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jadwal yang sudah dipesan tidak dapat dihapus"})
		return
	}
	config.DB.Delete(&j)
	c.JSON(http.StatusOK, gin.H{"message": "Jadwal berhasil dihapus"})
}

func CreateJadwalBulk(c *gin.Context) {
	var input struct {
		LapanganID uint     `json:"lapangan_id" binding:"required"`
		Tanggals   []string `json:"tanggals" binding:"required"`
		JamMulai   string   `json:"jam_mulai" binding:"required"`
		JamSelesai string   `json:"jam_selesai" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.JamMulai >= input.JamSelesai {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jam mulai harus lebih kecil dari jam selesai"})
		return
	}

	var created, failed int
	for _, tanggal := range input.Tanggals {
		j := models.Jadwal{
			LapanganID: input.LapanganID,
			Tanggal:    tanggal,
			JamMulai:   input.JamMulai,
			JamSelesai: input.JamSelesai,
			Status:     "tersedia",
		}

		// cek overlap
		var existing models.Jadwal
		err := config.DB.Where(
			"lapangan_id = ? AND tanggal = ? AND status != ? AND "+
				"jam_mulai < ? AND jam_selesai > ?",
			j.LapanganID, j.Tanggal, "expired",
			j.JamSelesai, j.JamMulai,
		).First(&existing).Error

		if err == nil {
			failed++
			continue
		}

		if config.DB.Create(&j).Error == nil {
			created++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Bulk jadwal selesai",
		"created": created,
		"failed":  failed,
	})
}
