package models

import "time"

type Pembayaran struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ReservasiID   uint      `json:"reservasi_id"`
	Reservasi     Reservasi `gorm:"foreignKey:ReservasiID" json:"Reservasi"`
	BuktiTransfer string    `json:"bukti_transfer"`
	NoWhatsapp    string    `json:"no_whatsapp"`
	Status        string    `gorm:"default:verifikasi" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
