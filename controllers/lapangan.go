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
)

func GetLapangan(c *gin.Context) {
	var lapangan []models.Lapangan
	query := config.DB
	if kategori := c.Query("kategori"); kategori != "" {
		query = query.Where("kategori = ?", kategori)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Order("nama ASC").Find(&lapangan)
	c.JSON(http.StatusOK, gin.H{"data": lapangan})
}

func GetLapanganByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var lap models.Lapangan
	if err := config.DB.First(&lap, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lapangan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": lap})
}

func CreateLapangan(c *gin.Context) {
	nama := c.PostForm("nama")
	kategori := c.PostForm("kategori")
	lokasi := c.PostForm("lokasi")
	deskripsi := c.PostForm("deskripsi")
	fasilitas := c.PostForm("fasilitas")
	hargaStr := c.PostForm("harga_per_jam")
	status := c.PostForm("status")

	if nama == "" || hargaStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan Harga wajib diisi"})
		return
	}

	harga, _ := strconv.ParseFloat(hargaStr, 64)
	if status == "" {
		status = "active"
	}

	lap := models.Lapangan{
		Nama:        nama,
		Kategori:    kategori,
		Lokasi:      lokasi,
		Deskripsi:   deskripsi,
		Fasilitas:   fasilitas,
		HargaPerJam: harga,
		Status:      status,
	}

	// Handle file upload
	file, err := c.FormFile("gambar")
	if err == nil {
		uploadDir := "uploads/lapangan"
		os.MkdirAll(uploadDir, os.ModePerm)
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		dst := filepath.Join(uploadDir, filename)
		c.SaveUploadedFile(file, dst)
		lap.GambarURL = "/" + filepath.ToSlash(dst)
	}

	config.DB.Create(&lap)
	c.JSON(http.StatusCreated, gin.H{"message": "Lapangan berhasil ditambahkan", "data": lap})
}

func UpdateLapangan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var lap models.Lapangan
	if err := config.DB.First(&lap, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lapangan tidak ditemukan"})
		return
	}

	nama := c.PostForm("nama")
	kategori := c.PostForm("kategori")
	lokasi := c.PostForm("lokasi")
	deskripsi := c.PostForm("deskripsi")
	fasilitas := c.PostForm("fasilitas")
	hargaStr := c.PostForm("harga_per_jam")
	status := c.PostForm("status")

	updates := map[string]interface{}{}
	if nama != "" {
		updates["nama"] = nama
	}
	if kategori != "" {
		updates["kategori"] = kategori
	}
	if lokasi != "" {
		updates["lokasi"] = lokasi
	}
	if deskripsi != "" {
		updates["deskripsi"] = deskripsi
	}
	if fasilitas != "" {
		updates["fasilitas"] = fasilitas
	}
	if hargaStr != "" {
		harga, _ := strconv.ParseFloat(hargaStr, 64)
		updates["harga_per_jam"] = harga
	}
	if status != "" {
		updates["status"] = status
	}

	// Handle file upload
	file, err := c.FormFile("gambar")
	if err == nil {
		uploadDir := "uploads/lapangan"
		os.MkdirAll(uploadDir, os.ModePerm)
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		dst := filepath.Join(uploadDir, filename)
		c.SaveUploadedFile(file, dst)
		updates["gambar_url"] = "/" + filepath.ToSlash(dst)
	}

	config.DB.Model(&lap).Updates(updates)
	config.DB.First(&lap, id) // reload
	c.JSON(http.StatusOK, gin.H{"message": "Lapangan berhasil diupdate", "data": lap})
}

func DeleteLapangan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var lap models.Lapangan
	if err := config.DB.First(&lap, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lapangan tidak ditemukan"})
		return
	}

	var count int64
	config.DB.Model(&models.Jadwal{}).Where("lapangan_id = ? AND status IN ('tersedia', 'dipesan')", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lapangan masih memiliki jadwal aktif/dipesan. Hapus jadwal terlebih dahulu."})
		return
	}

	config.DB.Delete(&lap)
	c.JSON(http.StatusOK, gin.H{"message": "Lapangan berhasil dihapus"})
}

func ToggleStatusLapangan(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var lap models.Lapangan
	if err := config.DB.First(&lap, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lapangan tidak ditemukan"})
		return
	}
	newStatus := "active"
	if lap.Status == "active" {
		newStatus = "inactive"
	}
	config.DB.Model(&lap).Update("status", newStatus)
	c.JSON(http.StatusOK, gin.H{"message": "Status lapangan diperbarui", "status": newStatus})
}
