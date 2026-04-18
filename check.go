package main

import (
	"fmt"
	"sport-reservation/config"
	"sport-reservation/models"
)

func main() {
	config.LoadEnv()
	config.ConnectDB()

	var j []models.Jadwal
	config.DB.Find(&j)
	for _, x := range j {
		fmt.Printf("ID: %v, Tanggal: %v, Mulai: %v, Selesai: %v\n", x.ID, x.Tanggal, x.JamMulai, x.JamSelesai)
	}
}
