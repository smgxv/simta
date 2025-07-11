package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"notification_service/config"
	"notification_service/models"

	"github.com/google/uuid"
)

func BroadcastNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	// Max 20MB file
	err := r.ParseMultipartForm(20 << 20)
	if err != nil {
		http.Error(w, "Gagal parsing form data", http.StatusBadRequest)
		return
	}

	judul := r.FormValue("judul")
	deskripsi := r.FormValue("deskripsi")
	targets := r.Form["target"] // checkbox → array
	files := r.MultipartForm.File["files"]

	var fileURLs []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Gagal membuka file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fileName := uuid.New().String() + "_" + fileHeader.Filename
		path := "./uploads/" + fileName

		dst, err := os.Create(path)
		if err != nil {
			http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		io.Copy(dst, file)
		fileURLs = append(fileURLs, "/uploads/"+fileName)
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	notif := models.Notification{
		ID:        uuid.New().String(),
		Judul:     judul,
		Deskripsi: deskripsi,
		Target:    strings.Join(targets, ","), // e.g., "Taruna,Dosen"
		CreatedAt: time.Now(),
	}

	// Handle optional file
	if len(fileURLs) > 0 {
		fileJSON, _ := json.Marshal(fileURLs)
		notif.FileURLs = string(fileJSON)
	}

	// Insert ke database
	query := `INSERT INTO notifications (id, judul, deskripsi, target, file_urls, created_at)
	          VALUES (?, ?, ?, ?, ?, ?)`
	_, err = db.Exec(query, notif.ID, notif.Judul, notif.Deskripsi, notif.Target, notif.FileURLs, notif.CreatedAt)
	if err != nil {
		http.Error(w, "Gagal menyimpan notifikasi", http.StatusInternalServerError)
		fmt.Println("DB error:", err)
		return
	}

	// Sukses
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notifikasi berhasil dikirim",
		"id":      notif.ID,
	})
}
