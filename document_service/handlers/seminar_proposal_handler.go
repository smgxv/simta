package handlers

import (
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func UploadSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ambil nilai form
	userID := r.FormValue("user_id")
	topikPenelitian := r.FormValue("topik_penelitian")
	ketuaPenguji := r.FormValue("ketua_penguji")
	penguji1 := r.FormValue("penguji1")
	penguji2 := r.FormValue("penguji2")

	// Ambil file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Buat direktori uploads jika belum ada
	uploadDir := "uploads/seminar_proposal"
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		http.Error(w, "Error creating upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Simpan file dengan nama unik
	filename := fmt.Sprintf("SeminarProposal_%s_%s_%s", userID, time.Now().Format("20060102150405"), handler.Filename)
	filePath := filepath.Join(uploadDir, filename)

	log.Printf("Saving file to: %s", filePath)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Buka koneksi database
	db, err := config.GetDB()
	if err != nil {
		os.Remove(filePath)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Konversi form string ke int
	userIDInt, _ := strconv.Atoi(userID)
	ketuaPengujiInt, _ := strconv.Atoi(ketuaPenguji)
	penguji1Int, _ := strconv.Atoi(penguji1)
	penguji2Int, _ := strconv.Atoi(penguji2)

	// Simpan entitas ke DB
	proposal := &entities.SeminarProposal{
		UserID:           userIDInt,
		TopikPenelitian:  topikPenelitian,
		FileProposalPath: filePath,
		KetuaPengujiID:   ketuaPengujiInt,
		Penguji1ID:       penguji1Int,
		Penguji2ID:       penguji2Int,
	}

	if err := models.InsertSeminarProposal(db, proposal); err != nil {
		os.Remove(filePath)
		http.Error(w, "Error saving to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Sukses
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Seminar proposal berhasil diunggah",
		"data": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// GetSeminarProposalByDosenHandler menangani request untuk mendapatkan data seminar proposal berdasarkan ID dosen
func GetSeminarProposalByDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil dosen_id dari query parameter
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

	// Buka koneksi database
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan data seminar proposal
	query := `
		SELECT sp.id, sp.user_id, sp.topik_penelitian, sp.file_proposal_path,
			   t.nama_lengkap as taruna_nama
		FROM seminar_proposal sp
		JOIN taruna t ON sp.user_id = t.user_id
		WHERE sp.ketua_penguji_id = ? OR sp.penguji1_id = ? OR sp.penguji2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ProposalData struct {
		ID              int    `json:"id"`
		UserID          int    `json:"user_id"`
		TopikPenelitian string `json:"topik_penelitian"`
		FilePath        string `json:"file_path"`
		TarunaNama      string `json:"taruna_nama"`
	}

	var proposals []ProposalData
	for rows.Next() {
		var p ProposalData
		err := rows.Scan(&p.ID, &p.UserID, &p.TopikPenelitian, &p.FilePath, &p.TarunaNama)
		if err != nil {
			http.Error(w, "Error scanning rows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		proposals = append(proposals, p)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   proposals,
	})
}

// GetTarunaListForDosenHandler menangani request untuk mendapatkan list taruna untuk dropdown
func GetTarunaListForDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil dosen_id dari query parameter
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

	// Buka koneksi database
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan data taruna dan topik penelitian
	query := `
		SELECT DISTINCT sp.user_id, t.nama_lengkap, sp.topik_penelitian
		FROM seminar_proposal sp
		JOIN taruna t ON sp.user_id = t.user_id
		WHERE sp.ketua_penguji_id = ? OR sp.penguji1_id = ? OR sp.penguji2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TarunaData struct {
		UserID          int    `json:"user_id"`
		NamaLengkap     string `json:"nama_lengkap"`
		TopikPenelitian string `json:"topik_penelitian"`
	}

	var tarunaList []TarunaData
	for rows.Next() {
		var t TarunaData
		err := rows.Scan(&t.UserID, &t.NamaLengkap, &t.TopikPenelitian)
		if err != nil {
			http.Error(w, "Error scanning rows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tarunaList = append(tarunaList, t)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   tarunaList,
	})
}

// PenilaianProposalHandler menangani request untuk menyimpan penilaian proposal
func PenilaianProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get form values
	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	seminarProposalID := r.FormValue("seminar_proposal_id")
	if seminarProposalID == "" {
		http.Error(w, "Seminar Proposal ID is required", http.StatusBadRequest)
		return
	}

	dosenID := r.FormValue("dosen_id")
	if dosenID == "" {
		http.Error(w, "Dosen ID is required", http.StatusBadRequest)
		return
	}

	// Get files
	penilaianFile, penilaianHeader, err := r.FormFile("penilaian_file")
	if err != nil {
		http.Error(w, "Error retrieving penilaian file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer penilaianFile.Close()

	beritaAcaraFile, beritaAcaraHeader, err := r.FormFile("berita_acara_file")
	if err != nil {
		http.Error(w, "Error retrieving berita acara file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer beritaAcaraFile.Close()

	// Create upload directories if they don't exist
	penilaianDir := "uploads/penilaian_proposal"
	beritaAcaraDir := "uploads/berita_acara_proposal"

	if err := os.MkdirAll(penilaianDir, 0777); err != nil {
		http.Error(w, "Error creating penilaian directory: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll(beritaAcaraDir, 0777); err != nil {
		http.Error(w, "Error creating berita acara directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate unique filenames
	timestamp := time.Now().Format("20060102150405")
	penilaianFilename := fmt.Sprintf("Penilaian_%s_%s_%s", userID, timestamp, penilaianHeader.Filename)
	beritaAcaraFilename := fmt.Sprintf("BeritaAcara_%s_%s_%s", userID, timestamp, beritaAcaraHeader.Filename)

	penilaianPath := filepath.Join(penilaianDir, penilaianFilename)
	beritaAcaraPath := filepath.Join(beritaAcaraDir, beritaAcaraFilename)

	// Save penilaian file
	penilaianDst, err := os.Create(penilaianPath)
	if err != nil {
		http.Error(w, "Error creating penilaian file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer penilaianDst.Close()

	if _, err = io.Copy(penilaianDst, penilaianFile); err != nil {
		os.Remove(penilaianPath)
		http.Error(w, "Error saving penilaian file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save berita acara file
	beritaAcaraDst, err := os.Create(beritaAcaraPath)
	if err != nil {
		os.Remove(penilaianPath)
		http.Error(w, "Error creating berita acara file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer beritaAcaraDst.Close()

	if _, err = io.Copy(beritaAcaraDst, beritaAcaraFile); err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		http.Error(w, "Error saving berita acara file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Open database connection
	db, err := config.GetDB()
	if err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Insert into database
	query := `
		INSERT INTO seminar_proposal_penilaian (
			user_id, seminar_proposal_id, dosen_id,
			file_penilaian_path, file_berita_acara_path,
			status_pengumpulan, submitted_at
		) VALUES (?, ?, ?, ?, ?, 'belum', NOW())
	`

	_, err = db.Exec(query,
		userID,
		seminarProposalID,
		dosenID,
		penilaianPath,
		beritaAcaraPath,
	)

	if err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		http.Error(w, "Error saving to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Penilaian proposal berhasil disimpan",
		"data": map[string]interface{}{
			"penilaian_path":    penilaianPath,
			"berita_acara_path": beritaAcaraPath,
		},
	})
}
