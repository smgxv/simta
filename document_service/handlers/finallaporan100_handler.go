package handlers

import (
	"database/sql"
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"document_service/utils"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

// Handler untuk mengupload final laporan 100%
func UploadFinalLaporan100Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error parsing form: " + err.Error(),
		})
		return
	}

	// Ambil data form
	userID := r.FormValue("user_id")
	namaLengkap := r.FormValue("nama_lengkap")
	jurusan := r.FormValue("jurusan")
	kelas := r.FormValue("kelas")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	// Validasi
	if userID == "" || namaLengkap == "" || topikPenelitian == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// === Upload Final Laporan 100 ===
	finalFilePath, err := utils.HandleFileUpload(r, "final_laporan100_file", userID, "FINAL_LAPORAN100")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal upload final laporan: " + err.Error(),
		})
		return
	}

	// === Upload Form Bimbingan ===
	formBimbinganPath, err := utils.HandleFileUpload(r, "form_bimbingan_file", userID, "FORM_BIMBINGAN100")
	if err != nil {
		os.Remove(finalFilePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal upload form bimbingan: " + err.Error(),
		})
		return
	}

	// === Simpan ke database ===
	db, err := config.GetDB()
	if err != nil {
		os.Remove(finalFilePath)
		os.Remove(formBimbinganPath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal koneksi database: " + err.Error(),
		})
		return
	}
	defer db.Close()

	finalLaporan100Model := models.NewFinalLaporan100Model(db)
	finalLaporan100 := &entities.FinalLaporan100{
		UserID:            utils.ParseInt(userID),
		NamaLengkap:       namaLengkap,
		Jurusan:           jurusan,
		Kelas:             kelas,
		TopikPenelitian:   topikPenelitian,
		FilePath:          finalFilePath,
		FormBimbinganPath: formBimbinganPath,
		Keterangan:        keterangan,
	}

	if err := finalLaporan100Model.Create(finalLaporan100); err != nil {
		os.Remove(finalFilePath)
		os.Remove(formBimbinganPath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal menyimpan ke database: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Final Laporan 100% dan Form Bimbingan berhasil diunggah",
		"data": map[string]interface{}{
			"id":                  finalLaporan100.ID,
			"file_path":           finalFilePath,
			"form_bimbingan_path": formBimbinganPath,
		},
	})
}

// Handler untuk mengambil daftar final proposal berdasarkan user_id
func GetFinalLaporan100Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	finalLaporan100Model := models.NewFinalLaporan100Model(db)
	finalLaporan100s, err := finalLaporan100Model.GetByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   finalLaporan100s,
	})
}

// Handler untuk mengambil data gabungan taruna dan final proposal
func GetAllFinalLaporan100WithTarunaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mengambil data gabungan
	query := `
		SELECT 
			t.user_id as taruna_id,
			t.nama_lengkap,
			t.jurusan,
			t.kelas,
			COALESCE(f.topik_penelitian, '') as topik_penelitian,
			COALESCE(f.status, '') as status,
			COALESCE(f.id, 0) as final_laporan100_id
		FROM taruna t
		LEFT JOIN final_laporan100 f ON t.user_id = f.user_id
		ORDER BY t.nama_lengkap ASC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TarunaLaporan100 struct {
		TarunaID          int    `json:"taruna_id"`
		NamaLengkap       string `json:"nama_lengkap"`
		Jurusan           string `json:"jurusan"`
		Kelas             string `json:"kelas"`
		TopikPenelitian   string `json:"topik_penelitian"`
		Status            string `json:"status"`
		FinalLaporan100ID int    `json:"final_laporan100_id"`
	}

	var results []TarunaLaporan100
	for rows.Next() {
		var data TarunaLaporan100
		err := rows.Scan(
			&data.TarunaID,
			&data.NamaLengkap,
			&data.Jurusan,
			&data.Kelas,
			&data.TopikPenelitian,
			&data.Status,
			&data.FinalLaporan100ID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, data)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   results,
	})
}

// Handler untuk update status Final Proposal
func UpdateFinalLaporan100StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var requestData struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := "UPDATE final_laporan100 SET status = ? WHERE id = ?"
	_, err = db.Exec(query, requestData.Status, requestData.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Status berhasil diupdate",
	})
}

// Handler untuk download file Final Laporan 100% atau Form Bimbingan
func DownloadFinalLaporan100Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil ID dan tipe file dari parameter
	vars := mux.Vars(r)
	laporanID := vars["id"]               // /final_laporan100/download/{id}
	fileType := r.URL.Query().Get("type") // ?type=laporan100 atau ?type=form_bimbingan

	if laporanID == "" || fileType == "" {
		http.Error(w, "Parameter 'id' dan 'type' wajib disediakan", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var filePath string
	switch fileType {
	case "laporan100":
		err = db.QueryRow("SELECT file_path FROM final_laporan100 WHERE id = ?", laporanID).Scan(&filePath)
	case "form_bimbingan":
		err = db.QueryRow("SELECT form_bimbingan_path FROM final_laporan100 WHERE id = ?", laporanID).Scan(&filePath)
	default:
		http.Error(w, "Parameter 'type' tidak valid. Gunakan 'laporan100' atau 'form_bimbingan'", http.StatusBadRequest)
		return
	}

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "File tidak ditemukan", http.StatusNotFound)
		} else {
			http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Gagal membuka file", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/pdf")
	http.ServeFile(w, r, filePath)
}
