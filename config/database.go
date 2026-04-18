package config

import (
	"fmt"
	"log"
	"time"

	"sport-reservation/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		GetEnv("DB_USER", "root"),
		GetEnv("DB_PASSWORD", ""),
		GetEnv("DB_HOST", "localhost"),
		GetEnv("DB_PORT", "3306"),
		GetEnv("DB_NAME", "sport_reservation"),
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database:", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Lapangan{},
		&models.Jadwal{},
		&models.Reservasi{},
		&models.Pembayaran{},
	)
	if err != nil {
		log.Fatal("Gagal migrate:", err)
	}

	log.Println("✅ Database connected & migrated")

	seedData(DB)
	seedAdmin(DB)
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&models.Lapangan{}).Count(&count)
	if count == 0 {
		lapangans := []models.Lapangan{
			{Nama: "Lapang Mini Soccer Andromeda", Kategori: "Mini Soccer", Lokasi: "Jl. Sudirman No 1", HargaPerJam: 150000, Deskripsi: "Fasilitas lengkap: Kamar ganti, Tribun penonton, Parkir luas, dan Cafe.", Status: "active"},
			{Nama: "Lapang Futsal Galaxy", Kategori: "Futsal", Lokasi: "Jl. Merdeka No 45", HargaPerJam: 100000, Deskripsi: "Lapang indoor menggunakan rumput sintetis standar internasional.", Status: "active"},
			{Nama: "Lapang Tenis 76", Kategori: "Tenis", Lokasi: "Jl. Kopo No 59", HargaPerJam: 100000, Deskripsi: "Lapangan indoor yang masih baru.", Status: "active"},
			{Nama: "Lapang Basket Sangkuriang", Kategori: "Basket", Lokasi: "Jl. Budiawan No 7", HargaPerJam: 120000, Deskripsi: "Nyaman dan fasilitas lengkap.", Status: "active"},
			{Nama: "Lapang Padel Pro", Kategori: "Padel", Lokasi: "Jl. Padel No 1", HargaPerJam: 150000, Deskripsi: "Kaca tempered standar internasional.", Status: "active"},
		}

		for i := range lapangans {
			db.Create(&lapangans[i])
			besok := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			jadwals := []models.Jadwal{
				{LapanganID: lapangans[i].ID, Tanggal: besok, JamMulai: "08:00:00", JamSelesai: "09:00:00", Status: "tersedia"},
				{LapanganID: lapangans[i].ID, Tanggal: besok, JamMulai: "10:00:00", JamSelesai: "11:00:00", Status: "tersedia"},
				{LapanganID: lapangans[i].ID, Tanggal: besok, JamMulai: "15:00:00", JamSelesai: "16:00:00", Status: "tersedia"},
			}
			for _, j := range jadwals {
				db.Create(&j)
			}
		}
		log.Println("✅ Seed Data Lapangan & Jadwal berhasil dimasukkan")
	}
}

func seedAdmin(db *gorm.DB) {
	var count int64

	db.Model(&models.User{}).Where("email = ?", "admin@sportify.com").Count(&count)
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 14)
		admin1 := models.User{
			Name:     "Admin Sportify",
			Email:    "admin@sportify.com",
			Password: string(hashed),
			Role:     "admin",
		}
		db.Create(&admin1)
		log.Println("✅ Akun admin default dibuat: admin@sportify.com / admin123")
	}

	db.Model(&models.User{}).Where("email = ?", "fikri@sportify.com").Count(&count)
	if count == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("maspik"), 14)
		admin2 := models.User{
			Name:     "Maspik",
			Email:    "fikri@sportify.com",
			Password: string(hashed),
			Role:     "admin",
		}
		db.Create(&admin2)
		log.Println("✅ Akun admin default dibuat: fikri@sportify.com / maspik")
	}
}
