package models

import "time"

type Jadwal struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	LapanganID uint      `json:"lapangan_id"`
	Lapangan   Lapangan  `gorm:"foreignKey:LapanganID" json:"Lapangan"`
	Tanggal    string    `json:"tanggal"`
	JamMulai   string    `json:"jam_mulai"`
	JamSelesai string    `json:"jam_selesai"`
	Status     string    `gorm:"default:tersedia" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
