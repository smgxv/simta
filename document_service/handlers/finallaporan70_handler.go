package handlers

import (
	"database/sql"
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"document_service/utils"
	"document_service/utils/filemanager"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

// Handler untuk mengupload final laporan 70
func UploadFinalLaporan70Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Cek ukuran total upload
	if r.ContentLength > (2 * filemanager.MaxFileSize) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Ukuran total file terlalu besar. Maksimal 30MB",
		})
		return
	}

	err := r.ParseMultipartForm(2 * filemanager.MaxFileSize)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal parsing form: " + err.Error(),
		})
		return
	}

	// Ambil field dari form
	userID := r.FormValue("user_id")
	namaLengkap := r.FormValue("nama_lengkap")
	jurusan := r.FormValue("jurusan")
	kelas := r.FormValue("kelas")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	if userID == "" || namaLengkap == "" || topikPenelitian == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Field wajib tidak boleh kosong",
		})
		return
	}

	// === Final Laporan 70 ===
	file1, handler1, err := r.FormFile("final_laporan70_file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal membaca file final laporan: " + err.Error(),
		})
		return
	}
	defer file1.Close()

	if err := filemanager.ValidateFileType(file1, handler1.Filename); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	file1.Seek(0, 0)

	safeName1 := filemanager.ValidateFileName(handler1.Filename)
	finalFilename := fmt.Sprintf("FINAL_LAPORAN70_%s_%s_%s", userID, time.Now().Format("20060102150405"), safeName1)
	finalFilePath, err := filemanager.SaveUploadedFile(file1, handler1, "uploads/final_laporan70", finalFilename)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// === Form Bimbingan ===
	file2, handler2, err := r.FormFile("form_bimbingan_file")
	if err != nil {
		os.Remove(finalFilePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Gagal membaca file form bimbingan: " + err.Error(),
		})
		return
	}
	defer file2.Close()

	if err := filemanager.ValidateFileType(file2, handler2.Filename); err != nil {
		os.Remove(finalFilePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	file2.Seek(0, 0)

	safeName2 := filemanager.ValidateFileName(handler2.Filename)
	formBimbinganFilename := fmt.Sprintf("FORM_BIMBINGAN70_%s_%s_%s", userID, time.Now().Format("20060102150405"), safeName2)
	formBimbinganPath, err := filemanager.SaveUploadedFile(file2, handler2, "uploads/form_bimbingan70", formBimbinganFilename)
	if err != nil {
		os.Remove(finalFilePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": err.Error(),
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

	finalLaporanModel := models.NewFinalLaporan70Model(db)
	final := &entities.FinalLaporan70{
		UserID:            utils.ParseInt(userID),
		NamaLengkap:       namaLengkap,
		Jurusan:           jurusan,
		Kelas:             kelas,
		TopikPenelitian:   topikPenelitian,
		FilePath:          finalFilePath,
		FormBimbinganPath: formBimbinganPath,
		Keterangan:        keterangan,
	}

	if err := finalLaporanModel.Create(final); err != nil {
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
		"message": "Final Laporan 70% dan Form Bimbingan berhasil diunggah",
		"data": map[string]interface{}{
			"id":                  final.ID,
			"file_path":           finalFilePath,
			"form_bimbingan_path": formBimbinganPath,
		},
	})
}

// Handler untuk mengambil daftar final laporan 70 berdasarkan user_id
func GetFinalLaporan70Handler(w http.ResponseWriter, r *http.Request) {
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

	finalLaporan70Model := models.NewFinalLaporan70Model(db)
	finalLaporan70s, err := finalLaporan70Model.GetByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   finalLaporan70s,
	})
}

// Handler untuk mengambil data gabungan taruna dan final laporan 70
func GetAllFinalLaporan70WithTarunaHandler(w http.ResponseWriter, r *http.Request) {
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
			COALESCE(f.id, 0) as final_laporan70_id
		FROM taruna t
		LEFT JOIN final_laporan70 f ON t.user_id = f.user_id
		ORDER BY t.nama_lengkap ASC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TarunaLaporan70 struct {
		TarunaID         int    `json:"taruna_id"`
		NamaLengkap      string `json:"nama_lengkap"`
		Jurusan          string `json:"jurusan"`
		Kelas            string `json:"kelas"`
		TopikPenelitian  string `json:"topik_penelitian"`
		Status           string `json:"status"`
		FinalLaporan70ID int    `json:"final_laporan70_id"`
	}

	var results []TarunaLaporan70
	for rows.Next() {
		var data TarunaLaporan70
		err := rows.Scan(
			&data.TarunaID,
			&data.NamaLengkap,
			&data.Jurusan,
			&data.Kelas,
			&data.TopikPenelitian,
			&data.Status,
			&data.FinalLaporan70ID,
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
func UpdateFinalLaporan70StatusHandler(w http.ResponseWriter, r *http.Request) {
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

	query := "UPDATE final_laporan70 SET status = ? WHERE id = ?"
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

// Handler untuk download file Final Laporan 70% atau Form Bimbingan
func DownloadFinalLaporan70Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil ID dan tipe file dari parameter
	vars := mux.Vars(r)
	laporanID := vars["id"]               // /final_laporan70/download/{id}
	fileType := r.URL.Query().Get("type") // ?type=laporan70 atau ?type=form_bimbingan

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
	case "laporan70":
		err = db.QueryRow("SELECT file_path FROM final_laporan70 WHERE id = ?", laporanID).Scan(&filePath)
	case "form_bimbingan":
		err = db.QueryRow("SELECT form_bimbingan_path FROM final_laporan70 WHERE id = ?", laporanID).Scan(&filePath)
	default:
		http.Error(w, "Parameter 'type' tidak valid. Gunakan 'laporan70' atau 'form_bimbingan'", http.StatusBadRequest)
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

func DownloadFinalLaporan70DosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil final_laporan70_id dari parameter
	vars := mux.Vars(r)
	laporanID := vars["id"] // /finallaporan70/dosen/download/{id}

	if laporanID == "" {
		http.Error(w, "Parameter 'id' wajib disediakan", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var filePath string
	err = db.QueryRow("SELECT file_path FROM final_laporan70 WHERE id = ?", laporanID).Scan(&filePath)
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
