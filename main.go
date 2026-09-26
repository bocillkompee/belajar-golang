package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"belajar-golang/config"
	"belajar-golang/controllers"
	"belajar-golang/models"
)

func main() {
	if err := config.ConnectDB(); err != nil {
		log.Fatalf("koneksi database gagal: %v", err)
	}
	defer config.DB.Close()

	if err := models.Inisialisasi(); err != nil {
		log.Fatalf("inisialisasi tabel gagal: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else {
		port = ":" + port
	}

	server := &http.Server{
		Addr:         port,
		Handler:      controllers.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("server berjalan di http://localhost%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server gagal: %v", err)
	}
}
