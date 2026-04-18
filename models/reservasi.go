package models

import "time"

type Reservasi struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `json:"user_id"`
	User       User       `gorm:"foreignKey:UserID" json:"User"`
	JadwalID   uint       `json:"jadwal_id"`
	Jadwal     Jadwal     `gorm:"foreignKey:JadwalID" json:"Jadwal"`
	Pembayaran *Pembayaran `gorm:"foreignKey:ReservasiID" json:"Pembayaran"`
	Status     string     `gorm:"default:pending" json:"status"`
	ExpiredAt  time.Time  `json:"expired_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
