package handlers

import (
	"database/sql"
	"document_service/config"
	"document_service/utils/filemanager"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Laporan70Data struct {
	FinalLaporan70ID int    `json:"final_laporan70_id"`
	UserID           int    `json:"user_id"`
	TopikPenelitian  string `json:"topik_penelitian"`
	FilePath         string `json:"file_path"`
	TarunaNama       string `json:"taruna_nama"`
}

// GetSeminarLaporan70ByDosenHandler menangani request untuk mendapatkan data seminar laporan70 berdasarkan ID dosen
func GetSeminarLaporan70ByDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	dosenID := r.URL.Query().Get("dosen_id")
	if dosenID == "" {
		http.Error(w, "Dosen ID is required", http.StatusBadRequest)
		return
	}

	dosenIDInt, err := strconv.Atoi(dosenID)
	if err != nil {
		http.Error(w, "Invalid dosen ID", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT fl.id, fl.user_id, fl.topik_penelitian, fl.file_path, u.nama_lengkap
		FROM final_laporan70 fl
		JOIN users u ON fl.user_id = u.id
		JOIN penguji_laporan70 pl ON fl.id = pl.final_laporan70_id
		WHERE pl.penguji_1_id = ? OR pl.penguji_2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Laporan70Data struct {
		FinalLaporan70ID int    `json:"final_laporan70_id"`
		UserID           int    `json:"user_id"`
		TopikPenelitian  string `json:"topik_penelitian"`
		FilePath         string `json:"file_path"`
		TarunaNama       string `json:"taruna_nama"`
	}

	var laporan70s []Laporan70Data
	for rows.Next() {
		var p Laporan70Data
		err := rows.Scan(&p.FinalLaporan70ID, &p.UserID, &p.TopikPenelitian, &p.FilePath, &p.TarunaNama)
		if err != nil {
			http.Error(w, "Error scanning rows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		laporan70s = append(laporan70s, p)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   laporan70s,
	})
}

// GetTarunaListForDosenHandler menangani request untuk mendapatkan daftar taruna yang belum memiliki final laporan70
func GetSeminarLaporan70TarunaListForDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	dosenID := r.URL.Query().Get("dosen_id")
	if dosenID == "" {
		http.Error(w, "Dosen ID is required", http.StatusBadRequest)
		return
	}

	dosenIDInt, err := strconv.Atoi(dosenID)
	if err != nil {
		http.Error(w, "Invalid dosen ID", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT DISTINCT 
			pp.user_id, 
			t.nama_lengkap, 
			fp.topik_penelitian,
			fp.id AS final_laporan70_id
		FROM penguji_laporan70 pp
		JOIN taruna t ON pp.user_id = t.user_id
		LEFT JOIN final_laporan70 fp ON fp.user_id = t.user_id
		WHERE 
			pp.penguji_1_id = ? 
			OR pp.penguji_2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TarunaData struct {
		UserID           int    `json:"user_id"`
		NamaLengkap      string `json:"nama_lengkap"`
		TopikPenelitian  string `json:"topik_penelitian"`
		FinalLaporan70ID int    `json:"final_laporan70_id"`
	}

	var tarunaList []TarunaData
	for rows.Next() {
		var t TarunaData
		err := rows.Scan(&t.UserID, &t.NamaLengkap, &t.TopikPenelitian, &t.FinalLaporan70ID)
		if err != nil {
			http.Error(w, "Error reading rows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tarunaList = append(tarunaList, t)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   tarunaList,
	})
}

// PenilaianLaporan70Handler menangani request untuk menyimpan penilaian laporan70
func PenilaianLaporan70Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Batas maksimum 2 file x 15MB
	if r.ContentLength > 2*filemanager.MaxFileSize {
		http.Error(w, "Total ukuran file terlalu besar. Maksimal 30MB", http.StatusRequestEntityTooLarge)
		return
	}

	if err := r.ParseMultipartForm(2 * filemanager.MaxFileSize); err != nil {
		http.Error(w, "Gagal parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	dosenID := r.FormValue("dosen_id")
	finalLaporan70ID := r.FormValue("final_laporan70_id")

	if userID == "" || dosenID == "" || finalLaporan70ID == "" {
		http.Error(w, "user_id, dosen_id, dan final_laporan70_id wajib diisi", http.StatusBadRequest)
		return
	}

	// === Upload file penilaian ===
	penilaianFile, penilaianHeader, err := r.FormFile("penilaian_file")
	if err != nil {
		http.Error(w, "Gagal membaca file penilaian: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer penilaianFile.Close()

	if err := filemanager.ValidateFileType(penilaianFile, penilaianHeader.Filename); err != nil {
		http.Error(w, "File penilaian tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}
	penilaianFile.Seek(0, 0)
	penilaianName := fmt.Sprintf("PENILAIAN_%s_%s_%s", userID, time.Now().Format("20060102150405"), filemanager.ValidateFileName(penilaianHeader.Filename))
	penilaianPath, err := filemanager.SaveUploadedFile(penilaianFile, penilaianHeader, "uploads/penilaian_laporan70", penilaianName)
	if err != nil {
		http.Error(w, "Gagal menyimpan file penilaian: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// === Upload file berita acara ===
	beritaFile, beritaHeader, err := r.FormFile("berita_acara_file")
	if err != nil {
		os.Remove(penilaianPath)
		http.Error(w, "Gagal membaca file berita acara: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer beritaFile.Close()

	if err := filemanager.ValidateFileType(beritaFile, beritaHeader.Filename); err != nil {
		os.Remove(penilaianPath)
		http.Error(w, "File berita acara tidak valid: "+err.Error(), http.StatusBadRequest)
		return
	}
	beritaFile.Seek(0, 0)
	beritaName := fmt.Sprintf("BERITAACARA_%s_%s_%s", userID, time.Now().Format("20060102150405"), filemanager.ValidateFileName(beritaHeader.Filename))
	beritaPath, err := filemanager.SaveUploadedFile(beritaFile, beritaHeader, "uploads/berita_acara_laporan70", beritaName)
	if err != nil {
		os.Remove(penilaianPath)
		http.Error(w, "Gagal menyimpan file berita acara: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// === Simpan ke database ===
	db, err := config.GetDB()
	if err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaPath)
		http.Error(w, "Koneksi database gagal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaPath)
		http.Error(w, "Gagal memulai transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM seminar_laporan70_penilaian 
			WHERE user_id = ? AND dosen_id = ? AND final_laporan70_id = ?
		)`, userID, dosenID, finalLaporan70ID).Scan(&exists)

	if err != nil {
		tx.Rollback()
		os.Remove(penilaianPath)
		os.Remove(beritaPath)
		http.Error(w, "Gagal cek data lama: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		_, err = tx.Exec(`
			UPDATE seminar_laporan70_penilaian 
			SET file_penilaian_path = ?, file_berita_acara_path = ?, 
				submitted_at = NOW(), status_pengumpulan = 'sudah'
			WHERE user_id = ? AND dosen_id = ? AND final_laporan70_id = ?
		`, penilaianPath, beritaPath, userID, dosenID, finalLaporan70ID)
	} else {
		_, err = tx.Exec(`
			INSERT INTO seminar_laporan70_penilaian (
				user_id, final_laporan70_id, dosen_id,
				file_penilaian_path, file_berita_acara_path,
				status_pengumpulan, submitted_at
			) VALUES (?, ?, ?, ?, ?, 'sudah', NOW())
		`, userID, finalLaporan70ID, dosenID, penilaianPath, beritaPath)
	}

	if err != nil {
		tx.Rollback()
		os.Remove(penilaianPath)
		os.Remove(beritaPath)
		http.Error(w, "Gagal menyimpan ke database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		os.Remove(penilaianPath)
		os.Remove(beritaPath)
		http.Error(w, "Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Penilaian laporan 70% berhasil diunggah",
		"data": map[string]interface{}{
			"penilaian_path":    penilaianPath,
			"berita_acara_path": beritaPath,
		},
	})
}

// DownloadFilePenilaianLaporan70Handler digunakan untuk mengunduh file penilaian atau berita acara laporan 70%
func DownloadFilePenilaianLaporan70Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil nama file dari query
	rawPath := r.URL.Query().Get("path")
	if rawPath == "" {
		http.Error(w, "Parameter 'path' wajib diisi", http.StatusBadRequest)
		return
	}

	fileName := filepath.Base(rawPath) // Mencegah path traversal

	// Tentukan direktori berdasarkan jenis file
	var baseDir string
	if strings.HasPrefix(fileName, "PENILAIAN_") {
		baseDir = "uploads/penilaian_laporan70"
	} else if strings.HasPrefix(fileName, "BERITAACARA_") {
		baseDir = "uploads/berita_acara_laporan70"
	} else {
		http.Error(w, "Prefix nama file tidak valid", http.StatusForbidden)
		return
	}

	// Bangun path absolut
	joinedPath := filepath.Join(baseDir, fileName)
	absPath, err := filepath.Abs(joinedPath)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil || !strings.HasPrefix(absPath, baseAbs) {
		http.Error(w, "Unauthorized file path", http.StatusForbidden)
		return
	}

	// Buka file
	file, err := os.Open(absPath)
	if err != nil {
		http.Error(w, "File tidak ditemukan", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set header untuk download
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/pdf")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Gagal mengirim file", http.StatusInternalServerError)
		return
	}
}

// Helper function untuk menyimpan file
func saveFile(src io.Reader, destPath string) error {
	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func GetMonitoringPenilaianLaporan70Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT
			fp.id AS final_laporan70_id,
			u.nama_lengkap AS nama_taruna,
			t.jurusan,
			fp.topik_penelitian,

			d1.nama_lengkap AS penguji1,
			d2.nama_lengkap AS penguji2,

			-- Hitung kelengkapan berkas
			CASE
				WHEN COUNT(CASE WHEN spp.file_penilaian_path IS NOT NULL AND spp.file_berita_acara_path IS NOT NULL THEN 1 END) = 2 THEN 'Lengkap'
				ELSE 'Belum Lengkap'
			END AS status_kelengkapan

		FROM penguji_laporan70 pp
		JOIN final_laporan70 fp ON fp.id = pp.final_laporan70_id
		JOIN users u ON u.id = pp.user_id
		JOIN taruna t ON t.user_id = u.id

		JOIN dosen d1 ON d1.id = pp.penguji_1_id
		JOIN dosen d2 ON d2.id = pp.penguji_2_id

		LEFT JOIN seminar_laporan70_penilaian spp
			ON spp.final_laporan70_id = pp.final_laporan70_id

		GROUP BY
			fp.id, u.nama_lengkap, t.jurusan, fp.topik_penelitian,
			d1.nama_lengkap, d2.nama_lengkap
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error querying database",
		})
		return
	}
	defer rows.Close()

	type MonitoringData struct {
		FinalLaporan70ID  int    `json:"final_laporan70_id"`
		NamaTaruna        string `json:"nama_taruna"`
		Jurusan           string `json:"jurusan"`
		TopikPenelitian   string `json:"topik_penelitian"`
		Penguji1          string `json:"penguji1"`
		Penguji2          string `json:"penguji2"`
		StatusKelengkapan string `json:"status_kelengkapan"`
	}

	var result []MonitoringData
	for rows.Next() {
		var m MonitoringData
		err := rows.Scan(
			&m.FinalLaporan70ID,
			&m.NamaTaruna,
			&m.Jurusan,
			&m.TopikPenelitian,
			&m.Penguji1,
			&m.Penguji2,
			&m.StatusKelengkapan,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		result = append(result, m)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error reading data",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   result,
	})
}

func GetFinalLaporan70DetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	finalLaporan70ID := vars["id"]
	if finalLaporan70ID == "" {
		http.Error(w, "Final Laporan70 ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query laporan70 dan taruna
	queryLaporan70 := `
		SELECT fp.id, u.nama_lengkap, u.jurusan, u.kelas, fp.topik_penelitian
		FROM final_laporan70 fp
		JOIN users u ON u.id = fp.user_id
		WHERE fp.id = ?
	`
	var (
		idLaporan70     int
		namaTaruna      string
		jurusan         string
		kelas           string
		topikPenelitian string
	)
	err = db.QueryRow(queryLaporan70, finalLaporan70ID).Scan(&idLaporan70, &namaTaruna, &jurusan, &kelas, &topikPenelitian)
	if err != nil {
		http.Error(w, "Laporan70 tidak ditemukan: "+err.Error(), http.StatusNotFound)
		return
	}

	// Query penguji
	queryPenguji := `
		SELECT penguji_1_id, penguji_2_id
		FROM penguji_laporan70
		WHERE final_laporan70_id = ?
		LIMIT 1
	`
	var penguji1ID, penguji2ID int
	err = db.QueryRow(queryPenguji, finalLaporan70ID).Scan(&penguji1ID, &penguji2ID)
	if err != nil {
		http.Error(w, "Data penguji tidak ditemukan: "+err.Error(), http.StatusNotFound)
		return
	}

	// Fungsi ambil penilaian
	getPenilaian := func(dosenID int) map[string]string {
		var namaDosen string
		err := db.QueryRow(`SELECT nama_lengkap FROM dosen WHERE id = ?`, dosenID).Scan(&namaDosen)
		if err != nil {
			namaDosen = "-"
		}

		var (
			status, filePenilaian, fileBA string
		)

		err = db.QueryRow(`
			SELECT status_pengumpulan, file_penilaian_path, file_berita_acara_path
			FROM seminar_laporan70_penilaian
			WHERE final_laporan70_id = ? AND dosen_id = ?
			LIMIT 1
		`, finalLaporan70ID, dosenID).Scan(&status, &filePenilaian, &fileBA)

		if err != nil {
			status = "belum"
			filePenilaian = ""
			fileBA = ""
		}

		return map[string]string{
			"nama_dosen":         namaDosen,
			"status_pengumpulan": status,
			"file_penilaian":     filePenilaian,
			"file_berita_acara":  fileBA,
		}
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"nama_taruna":      namaTaruna,
			"jurusan":          jurusan,
			"kelas":            kelas,
			"topik_penelitian": topikPenelitian,
			"penguji1":         getPenilaian(penguji1ID),
			"penguji2":         getPenilaian(penguji2ID),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function untuk mendapatkan data penilaian dosen
func getPengujiLaporan70Data(db *sql.DB, laporan70ID string, dosenID int) map[string]interface{} {
	query := `
		SELECT 
			d.nama_lengkap,
			spp.file_penilaian_path,
			spp.file_berita_acara_path,
			spp.status_pengumpulan
		FROM dosen d
		LEFT JOIN seminar_laporan70_penilaian spp ON spp.dosen_id = d.id 
			AND spp.seminar_laporan70_id = ?
		WHERE d.id = ?
	`

	var data struct {
		NamaLengkap       string
		FilePenilaianPath sql.NullString
		BeritaAcaraPath   sql.NullString
		StatusPengumpulan sql.NullString
	}

	err := db.QueryRow(query, laporan70ID, dosenID).Scan(
		&data.NamaLengkap,
		&data.FilePenilaianPath,
		&data.BeritaAcaraPath,
		&data.StatusPengumpulan,
	)

	if err != nil {
		return map[string]interface{}{
			"nama_lengkap":           "-",
			"file_penilaian_path":    "",
			"file_berita_acara_path": "",
			"status_pengumpulan":     "belum",
		}
	}

	return map[string]interface{}{
		"nama_lengkap":           data.NamaLengkap,
		"file_penilaian_path":    data.FilePenilaianPath.String,
		"file_berita_acara_path": data.BeritaAcaraPath.String,
		"status_pengumpulan":     data.StatusPengumpulan.String,
	}
}

func GetCatatanPerbaikanTarunaLaporan70Handler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil user_id dari query
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User ID is required",
		})
		return
	}

	// Koneksi ke DB
	db, err := config.GetDB()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database connection error: " + err.Error(),
		})
		return
	}
	defer db.Close()

	// Query ambil catatan perbaikan
	query := `
		SELECT spp.id, d.nama_lengkap, fp.topik_penelitian, spp.file_penilaian_path, spp.submitted_at
		FROM seminar_laporan70_penilaian spp
		JOIN dosen d ON spp.dosen_id = d.id
		LEFT JOIN final_laporan70 fp ON spp.final_laporan70_id = fp.id
		WHERE spp.user_id = ? AND spp.status_pengumpulan = 'sudah'
		ORDER BY spp.submitted_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Query error: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var results []CatatanPerbaikan
	for rows.Next() {
		var c CatatanPerbaikan
		err := rows.Scan(&c.ID, &c.NamaDosen, &c.TopikPenelitian, &c.FilePath, &c.SubmittedAt)
		if err != nil {
			continue // skip baris yang gagal diparse
		}
		results = append(results, c)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   results,
	})
}
