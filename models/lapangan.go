package models

import "time"

type Lapangan struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Nama        string    `json:"nama"`
	Kategori    string    `json:"kategori"`    // futsal, badminton, basket, tenis, voli, dll
	Lokasi      string    `json:"lokasi"`
	Deskripsi   string    `json:"deskripsi"`
	Fasilitas   string    `json:"fasilitas"`   // deskripsi fasilitas (parkir, toilet, dll)
	HargaPerJam float64   `json:"harga_per_jam"`
	GambarURL   string    `json:"gambar_url"`  // URL gambar lapangan
	Status      string    `gorm:"default:active" json:"status"` // active / inactive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
