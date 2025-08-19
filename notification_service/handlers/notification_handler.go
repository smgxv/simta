package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"notification_service/config"
	"notification_service/models"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	MaxUploadBytes = 15 << 20 // 15 MB
	uploadDir      = "./uploads"
)

// ekstensi yang diizinkan
var allowedExt = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

// MIME yang diizinkan (hasil sniffing / Content-Type)
var allowedMIME = map[string]bool{
	"application/pdf":    true,
	"application/msword": true, // .doc
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
	"application/vnd.ms-excel": true, // .xls
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func validateFileHeader(h *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(h.Filename))
	if !allowedExt[ext] {
		return fmt.Errorf("jenis file tidak diizinkan (hanya PDF, DOC/DOCX, XLS/XLSX)")
	}
	// periksa ukuran yang dilaporkan header (tidak selalu akurat, tapi cepat)
	if h.Size > MaxUploadBytes {
		return fmt.Errorf("ukuran file melebihi 15MB")
	}
	return nil
}

func sniffMIME(rc io.ReadSeeker) (string, error) {
	// baca 512 byte untuk deteksi tipe konten
	buf := make([]byte, 512)
	n, err := rc.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	// kembalikan cursor ke awal sebelum dipakai copy
	if _, err := rc.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	ct := http.DetectContentType(buf[:n])
	return ct, nil
}

func BroadcastNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Batasi total ukuran request (hard limit)
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadBytes+1) // +1 untuk deteksi over
	// Parse form multipart (nilai di sini adalah ambang penggunaan RAM untuk parts)
	if err := r.ParseMultipartForm(MaxUploadBytes + (1 << 20)); err != nil {
		http.Error(w, `Gagal parsing form data atau ukuran melebihi 15MB`, http.StatusBadRequest)
		return
	}

	judul := r.FormValue("judul")
	deskripsi := r.FormValue("deskripsi")
	targets := r.Form["target"] // checkbox → array
	files := r.MultipartForm.File["files"]

	if deskripsi == "" {
		deskripsi = "-"
	}

	// Pastikan folder ada
	if err := os.MkdirAll(uploadDir, 0o750); err != nil {
		http.Error(w, `{"error":"Gagal membuat direktori upload"}`, http.StatusInternalServerError)
		return
	}

	var fileURLs []string

	for _, fh := range files {
		// Validasi awal berdasar header (ekstensi & size dari header)
		if err := validateFileHeader(fh); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}

		src, err := fh.Open()
		if err != nil {
			http.Error(w, `{"error":"Gagal membuka file"}`, http.StatusInternalServerError)
			return
		}
		// Pastikan bisa seek untuk sniff MIME
		rs, ok := src.(io.ReadSeeker)
		if !ok {
			// jika bukan ReadSeeker, bungkus ke memori sementara (ukuran max sudah dibatasi)
			var buf strings.Builder
			if _, err := io.CopyN(&buf, src, MaxUploadBytes+1); err != nil && err != io.EOF {
				src.Close()
				http.Error(w, `{"error":"Gagal membaca file"}`, http.StatusInternalServerError)
				return
			}
			src.Close()
			// ukuran runtime check
			if buf.Len() > MaxUploadBytes {
				http.Error(w, `Ukuran file melebihi 15MB`, http.StatusBadRequest)
				return
			}
			// buat ReadSeeker dari bytes
			rs = strings.NewReader(buf.String())
		}
		defer func(c io.Closer) {
			_ = c.Close()
		}(io.NopCloser(rs))

		// Sniff MIME
		ct, err := sniffMIME(rs)
		if err != nil {
			http.Error(w, `{"error":"Gagal mendeteksi tipe file"}`, http.StatusBadRequest)
			return
		}
		if !allowedMIME[ct] {
			http.Error(w, `{"error":"Tipe file tidak diizinkan (PDF/DOC/DOCX/XLS/XLSX)"}`, http.StatusBadRequest)
			return
		}

		// Siapkan nama file yang aman + timestamp
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		base := strings.TrimSuffix(fh.Filename, ext)
		base = sanitizeFilename(base)
		fileName := fmt.Sprintf("%s_%s%s", base, time.Now().Format("20060102_150405"), ext)
		finalPath := filepath.Join(uploadDir, fileName)

		// Buat file tujuan dengan O_EXCL (hindari overwrite)
		dst, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o640)
		if err != nil {
			http.Error(w, `{"error":"Gagal menyimpan file"}`, http.StatusInternalServerError)
			return
		}

		// Copy dengan limiter (<= 15MB). Jika lebih, hapus file dan error.
		written, err := io.Copy(dst, io.LimitReader(rs, MaxUploadBytes+1))
		_ = dst.Close()
		if err != nil {
			_ = os.Remove(finalPath)
			http.Error(w, `{"error":"Gagal menyimpan file"}`, http.StatusInternalServerError)
			return
		}
		if written > MaxUploadBytes {
			_ = os.Remove(finalPath)
			http.Error(w, `{"error":"Ukuran file melebihi 15MB"}`, http.StatusBadRequest)
			return
		}

		fileURLs = append(fileURLs, "/uploads/"+fileName)
	}

	// Hilangkan target duplikat
	unique := map[string]struct{}{}
	var filteredTargets []string
	for _, t := range targets {
		if _, ok := unique[t]; !ok {
			unique[t] = struct{}{}
			filteredTargets = append(filteredTargets, t)
		}
	}

	// Simpan notifikasi ke DB
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	type Notification struct {
		Judul     string
		Deskripsi string
		Target    string
		FileURLs  string
		CreatedAt time.Time
	}
	notif := Notification{
		Judul:     judul,
		Deskripsi: deskripsi,
		Target:    strings.Join(filteredTargets, ","),
		CreatedAt: time.Now(),
	}
	if len(fileURLs) > 0 {
		b, _ := json.Marshal(fileURLs)
		notif.FileURLs = string(b)
	}

	query := `INSERT INTO notifications (judul, deskripsi, target, file_urls, created_at) VALUES (?, ?, ?, ?, ?)`
	res, err := db.Exec(query, notif.Judul, notif.Deskripsi, notif.Target, notif.FileURLs, notif.CreatedAt)
	if err != nil {
		log.Printf("❌ INSERT error: %+v", err)
		http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Notifikasi berhasil dikirim",
		"id":      id,
		"files":   fileURLs,
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

	// Ambil role dari query parameter
	role := r.URL.Query().Get("role")
	if role == "" {
		role = "Taruna" // default fallback
	}

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
		// Filter sesuai role
		if strings.Contains(n.Target, role) {
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
	role := r.URL.Query().Get("role") // Ambil role dari query parameter (e.g. "Dosen" atau "Taruna")

	if role != "Taruna" && role != "Dosen" {
		http.Error(w, "Role tidak valid", http.StatusBadRequest)
		return
	}

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

	// Filter berdasarkan target
	if !strings.Contains(notif.Target, role) {
		http.Error(w, "Notifikasi tidak ditujukan untuk pengguna ini", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(notif)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	fileName := vars["filename"]

	// Lokasi file
	filePath := filepath.Join("uploads", fileName)

	// Cek apakah file ada
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File tidak ditemukan", http.StatusNotFound)
		return
	}

	// Set header untuk download
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")

	// Kirim file
	http.ServeFile(w, r, filePath)
}
