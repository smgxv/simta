package handlers

import (
	"database/sql"
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"document_service/utils/filemanager"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Handler untuk mengupload final Proposal + file pendukung (wajib)
func UploadFinalProposalHandler(w http.ResponseWriter, r *http.Request) {
	// ===== CORS =====
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// ===== Batas total payload (opsional) =====
	if r.ContentLength > filemanager.MaxFileSize*4 { // mis. 15MB * 4 = 60MB
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Total file terlalu besar. Maksimal 60MB",
		})
		return
	}

	// ===== Parse form =====
	if err := r.ParseMultipartForm(filemanager.MaxFileSize * 4); err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Gagal parsing form data",
		})
		return
	}

	// ===== Ambil field =====
	userID := r.FormValue("user_id")
	namaLengkap := r.FormValue("nama_lengkap")
	jurusan := r.FormValue("jurusan")
	kelas := r.FormValue("kelas")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	if userID == "" || namaLengkap == "" || jurusan == "" || kelas == "" || topikPenelitian == "" {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Missing required fields",
		})
		return
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "User ID tidak valid",
		})
		return
	}

	// ===== Konfigurasi validasi file =====
	maxSize := int64(15 * 1024 * 1024) // 15MB
	allowedFinalExt := map[string]bool{".pdf": true}
	allowedSupportExt := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	}
	hasAllowedExt := func(name string, allowed map[string]bool) bool {
		ext := strings.ToLower(filepath.Ext(name))
		return allowed[ext]
	}

	// ===== FINAL Proposal (wajib, PDF) =====
	file, handler, err := r.FormFile("file")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "File final wajib diunggah: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Ukuran
	if handler.Size > maxSize {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Ukuran file final melebihi 15MB",
		})
		return
	}
	// Ekstensi
	if !hasAllowedExt(handler.Filename, allowedFinalExt) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Tipe file final Proposal tidak diizinkan (hanya PDF)",
		})
		return
	}
	// (Opsional) Content sniffing khusus PDF (pakai util kamu)
	if err := filemanager.ValidateFileType(file, handler.Filename); err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	_, _ = file.Seek(0, 0)

	// Simpan final Proposal
	safeFilename := filemanager.ValidateFileName(handler.Filename)
	finalName := fmt.Sprintf("FINAL_Proposal_%s_%s_%s", userID, time.Now().Format("20060102150405"), safeFilename)
	finalDir := "uploads/finalproposal"
	filePath, err := filemanager.SaveUploadedFile(file, handler, finalDir, finalName)
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	// Kumpulan path untuk cleanup jika gagal
	savedPaths := []string{filePath}

	// ===== FILE PENDUKUNG (wajib minimal 1) =====
	supportFiles := r.MultipartForm.File["support_files[]"]
	if len(supportFiles) == 0 {
		_ = os.Remove(filePath)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Minimal 1 file pendukung wajib diunggah",
		})
		return
	}

	var supportPaths []string
	supportDir := "uploads/pendukungproposal"

	for _, fh := range supportFiles {
		f, err := fh.Open()
		if err != nil {
			_ = os.Remove(filePath)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  "error",
				"message": "Gagal membuka file pendukung: " + err.Error(),
			})
			return
		}

		// Ukuran
		if fh.Size > maxSize {
			f.Close()
			_ = os.Remove(filePath)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  "error",
				"message": fmt.Sprintf("Ukuran file pendukung melebihi 15MB (%s)", fh.Filename),
			})
			return
		}
		// Ekstensi
		if !hasAllowedExt(fh.Filename, allowedSupportExt) {
			f.Close()
			_ = os.Remove(filePath)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  "error",
				"message": fmt.Sprintf("Tipe file pendukung tidak diizinkan (%s). Hanya PDF, DOC, DOCX, XLS, XLSX.", fh.Filename),
			})
			return
		}
		// (Opsional) Sniffing — jika filemanager.ValidateFileType hanya untuk PDF,
		// biarkan ext check untuk DOC/XLS; tetap aman karena kita sanitize nama & batasi size.

		_, _ = f.Seek(0, 0)
		safeName := filemanager.ValidateFileName(fh.Filename)
		// Pakai timestamp high-res agar unik walau banyak file per detik
		supportName := fmt.Sprintf("%d_%s", time.Now().Unix(), safeName)

		outPath, err := filemanager.SaveUploadedFile(f, fh, supportDir, supportName)
		f.Close()
		if err != nil {
			_ = os.Remove(filePath)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		savedPaths = append(savedPaths, outPath)
		supportPaths = append(supportPaths, outPath)
	}

	// ===== SIMPAN DB =====
	db, err := config.GetDB()
	if err != nil {
		for _, p := range savedPaths {
			_ = os.Remove(p)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Error connecting to database: " + err.Error(),
		})
		return
	}
	defer db.Close()

	finalProposalModel := models.NewFinalProposalModel(db)
	finalProposal := &entities.FinalProposal{
		UserID:          userIDInt,
		NamaLengkap:     namaLengkap,
		Jurusan:         jurusan,
		Kelas:           kelas,
		TopikPenelitian: topikPenelitian,
		FilePath:        filePath, // simpan path relatif/absolut sesuai kebijakanmu
		Keterangan:      keterangan,
	}

	// Set JSON array path pendukung
	if err := finalProposal.SetProposalSupportingFiles(supportPaths); err != nil {
		for _, p := range savedPaths {
			_ = os.Remove(p)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Gagal encode file pendukung: " + err.Error(),
		})
		return
	}

	if err := finalProposalModel.Create(finalProposal); err != nil {
		for _, p := range savedPaths {
			_ = os.Remove(p)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Error saving to database: " + err.Error(),
		})
		return
	}

	// ===== RESPONSE =====
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Final Proposal dan file pendukung berhasil diunggah",
		"data": map[string]any{
			"id":                  finalProposal.ID,
			"file_path":           filePath,
			"file_pendukung_path": supportPaths, // array path (bukan JSON string) untuk kenyamanan klien
		},
	})
}

// Handler untuk mengambil daftar final proposal berdasarkan user_id
func GetFinalProposalHandler(w http.ResponseWriter, r *http.Request) {
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

	finalProposalModel := models.NewFinalProposalModel(db)
	finalProposals, err := finalProposalModel.GetByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mapped := []map[string]interface{}{}
	for _, proposal := range finalProposals {
		// Parse daftar path pendukung dari kolom JSON (aman jika kosong/invalid)
		var supportPaths []string
		if paths, err := proposal.GetProposalSupportingFiles(); err == nil && len(paths) > 0 {
			supportPaths = paths
		}

		// Buat URL download per index
		filePendukungURL := make([]string, 0, len(supportPaths))
		supportFiles := make([]map[string]interface{}, 0, len(supportPaths))
		for i, p := range supportPaths {
			url := fmt.Sprintf("/api/document/finalproposal/download/%d?type=support&index=%d", proposal.ID, i)
			filePendukungURL = append(filePendukungURL, url)

			name := filepath.Base(p)
			supportFiles = append(supportFiles, map[string]interface{}{
				"index": i,
				"name":  name,
				"url":   url,
			})
		}

		mapped = append(mapped, map[string]interface{}{
			"taruna_id":          proposal.UserID,
			"nama_lengkap":       proposal.NamaLengkap,
			"jurusan":            proposal.Jurusan,
			"kelas":              proposal.Kelas,
			"topik_penelitian":   proposal.TopikPenelitian,
			"status":             proposal.Status,
			"final_proposal_id":  proposal.ID, // penting untuk download by id
			"final_download_url": fmt.Sprintf("/api/document/finalproposal/download/%d?type=final", proposal.ID),

			// Kompatibilitas + kemudahan frontend:
			"file_pendukung_path": proposal.FilePendukungPath, // string JSON asli dari DB (jika ingin)
			"file_pendukung_url":  filePendukungURL,           // array URL download
			"support_files":       supportFiles,               // array object {index,name,url}
		})
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   mapped,
	})
}

// Handler untuk mengambil data gabungan taruna dan final proposal
func GetAllFinalProposalWithTarunaHandler(w http.ResponseWriter, r *http.Request) {
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
			COALESCE(f.id, 0) as final_proposal_id
		FROM taruna t
		LEFT JOIN final_proposal f ON t.user_id = f.user_id
		ORDER BY t.nama_lengkap ASC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TarunaProposal struct {
		TarunaID        int    `json:"taruna_id"`
		NamaLengkap     string `json:"nama_lengkap"`
		Jurusan         string `json:"jurusan"`
		Kelas           string `json:"kelas"`
		TopikPenelitian string `json:"topik_penelitian"`
		Status          string `json:"status"`
		FinalProposalID int    `json:"final_proposal_id"`
	}

	var results []TarunaProposal
	for rows.Next() {
		var data TarunaProposal
		err := rows.Scan(
			&data.TarunaID,
			&data.NamaLengkap,
			&data.Jurusan,
			&data.Kelas,
			&data.TopikPenelitian,
			&data.Status,
			&data.FinalProposalID,
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
func UpdateFinalProposalStatusHandler(w http.ResponseWriter, r *http.Request) {
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

	query := "UPDATE final_proposal SET status = ? WHERE id = ?"
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

// Handler untuk download file Final Proposal atau file pendukung
func DownloadFinalProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	proposalID := vars["id"]

	fileType := r.URL.Query().Get("type")    // "final" atau "support"
	supportIdx := r.URL.Query().Get("index") // index untuk file pendukung

	if fileType == "" {
		fileType = "final"
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var filePath string

	if fileType == "final" {
		// Ambil path file final
		query := "SELECT file_path FROM final_proposal WHERE id = ?"
		err = db.QueryRow(query, proposalID).Scan(&filePath)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	} else if fileType == "support" {
		// Ambil JSON file pendukung
		var filePendukungJSON string
		query := "SELECT file_pendukung_path FROM final_proposal WHERE id = ?"
		err = db.QueryRow(query, proposalID).Scan(&filePendukungJSON)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "File pendukung tidak ditemukan", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Decode JSON ke slice
		var paths []string
		if err := json.Unmarshal([]byte(filePendukungJSON), &paths); err != nil {
			http.Error(w, "Gagal membaca data file pendukung", http.StatusInternalServerError)
			return
		}

		// Ambil index
		idx, err := strconv.Atoi(supportIdx)
		if err != nil || idx < 0 || idx >= len(paths) {
			http.Error(w, "Index file pendukung tidak valid", http.StatusBadRequest)
			return
		}

		filePath = paths[idx]
	} else {
		http.Error(w, "Tipe file tidak valid", http.StatusBadRequest)
		return
	}

	// Pastikan file ada
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	fileName := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

func DownloadFinalProposalDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil final_proposal_id dari parameter
	vars := mux.Vars(r)
	proposalID := vars["id"] // /finalproposal/dosen/download/{id}

	if proposalID == "" {
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
	err = db.QueryRow("SELECT file_path FROM final_proposal WHERE id = ?", proposalID).Scan(&filePath)
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
