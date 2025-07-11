package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"notification_service/config"
	"notification_service/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
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

		ext := filepath.Ext(fileHeader.Filename)
		nameOnly := strings.TrimSuffix(fileHeader.Filename, ext)
		timestamp := time.Now().Format("20060102_150405")
		fileName := fmt.Sprintf("%s_%s%s", nameOnly, timestamp, ext)

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

	// Validasi fallback deskripsi
	if deskripsi == "" {
		deskripsi = "-"
	}

	// Hilangkan target duplikat
	uniqueTargetMap := map[string]bool{}
	var filteredTargets []string
	for _, t := range targets {
		if !uniqueTargetMap[t] {
			uniqueTargetMap[t] = true
			filteredTargets = append(filteredTargets, t)
		}
	}

	notif := models.Notification{
		Judul:     judul,
		Deskripsi: deskripsi,
		Target:    strings.Join(filteredTargets, ","),
		CreatedAt: time.Now(),
	}

	// Masukkan fileURLs yang telah dikumpulkan
	if len(fileURLs) > 0 {
		fileJSON, _ := json.Marshal(fileURLs)
		notif.FileURLs = string(fileJSON)
	}

	query := `INSERT INTO notifications (judul, deskripsi, target, file_urls, created_at) VALUES (?, ?, ?, ?, ?)`
	res, err := db.Exec(query, notif.Judul, notif.Deskripsi, notif.Target, notif.FileURLs, notif.CreatedAt)

	if err != nil {
		log.Printf("❌ INSERT error: %+v", err)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ambil ID hasil insert
	insertedID, _ := res.LastInsertId()

	// Kirim response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Notifikasi berhasil dikirim",
		"id":      insertedID,
	})

}

func GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, judul, deskripsi, target, file_urls, created_at FROM notifications ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []models.Notification

	for rows.Next() {
		var n models.Notification
		err := rows.Scan(&n.ID, &n.Judul, &n.Deskripsi, &n.Target, &n.FileURLs, &n.CreatedAt)
		if err != nil {
			continue
		}
		// Filter hanya untuk target Taruna
		if strings.Contains(n.Target, "Taruna") {
			results = append(results, n)
		}
	}

	json.NewEncoder(w).Encode(results)
}

func GetNotificationByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var notif models.Notification
	row := db.QueryRow(`
		SELECT id, judul, deskripsi, target, file_urls, created_at 
		FROM notifications 
		WHERE id = ?`, id)

	err = row.Scan(&notif.ID, &notif.Judul, &notif.Deskripsi, &notif.Target, &notif.FileURLs, &notif.CreatedAt)
	if err != nil {
		http.Error(w, "Notifikasi tidak ditemukan", http.StatusNotFound)
		return
	}

	// Opsional: Filter agar hanya target Taruna yang bisa lihat
	if !strings.Contains(notif.Target, "Taruna") {
		http.Error(w, "Notifikasi tidak untuk pengguna ini", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(notif)
}
