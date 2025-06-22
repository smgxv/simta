package handlers

import (
	"database/sql"
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

	"github.com/gorilla/mux"
)

type CatatanPerbaikan struct {
	ID              int    `json:"id"`
	NamaDosen       string `json:"nama_dosen"`
	TopikPenelitian string `json:"topik_penelitian"`
	FilePath        string `json:"file_path"`
	SubmittedAt     string `json:"submitted_at"`
}

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

// GetTarunaListForDosenHandler menangani request untuk mendapatkan daftar taruna yang belum memiliki final proposal
func GetTarunaListForDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
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
			fp.id AS final_proposal_id
		FROM penguji_proposal pp
		JOIN taruna t ON pp.user_id = t.user_id
		LEFT JOIN final_proposal fp ON fp.user_id = t.user_id
		WHERE 
			pp.ketua_penguji_id = ? 
			OR pp.penguji_1_id = ? 
			OR pp.penguji_2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TarunaData struct {
		UserID          int    `json:"user_id"`
		NamaLengkap     string `json:"nama_lengkap"`
		TopikPenelitian string `json:"topik_penelitian"`
		FinalProposalID int    `json:"final_proposal_id"`
	}

	var tarunaList []TarunaData
	for rows.Next() {
		var t TarunaData
		err := rows.Scan(&t.UserID, &t.NamaLengkap, &t.TopikPenelitian, &t.FinalProposalID)
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

// PenilaianProposalHandler menangani request untuk menyimpan penilaian proposal
func PenilaianProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set header JSON di awal fungsi
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Fungsi helper untuk mengirim error response
	sendErrorResponse := func(message string, statusCode int) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": message,
		})
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		sendErrorResponse("Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validasi input
	userID := r.FormValue("user_id")
	dosenID := r.FormValue("dosen_id")
	finalProposalID := r.FormValue("final_proposal_id")

	if userID == "" || dosenID == "" || finalProposalID == "" {
		sendErrorResponse("user_id, dosen_id, dan final_proposal_id harus diisi", http.StatusBadRequest)
		return
	}

	// Get files
	penilaianFile, penilaianHeader, err := r.FormFile("penilaian_file")
	if err != nil {
		sendErrorResponse("Error retrieving penilaian file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer penilaianFile.Close()

	beritaAcaraFile, beritaAcaraHeader, err := r.FormFile("berita_acara_file")
	if err != nil {
		sendErrorResponse("Error retrieving berita acara file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer beritaAcaraFile.Close()

	// Create directories
	penilaianDir := "uploads/penilaian_proposal"
	beritaAcaraDir := "uploads/berita_acara_proposal"

	if err := os.MkdirAll(penilaianDir, 0777); err != nil {
		sendErrorResponse("Error creating directories: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll(beritaAcaraDir, 0777); err != nil {
		sendErrorResponse("Error creating directories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate filenames
	timestamp := time.Now().Format("20060102150405")
	penilaianFilename := fmt.Sprintf("Penilaian_%s_%s_%s", userID, timestamp, penilaianHeader.Filename)
	beritaAcaraFilename := fmt.Sprintf("BeritaAcara_%s_%s_%s", userID, timestamp, beritaAcaraHeader.Filename)

	penilaianPath := filepath.Join(penilaianDir, penilaianFilename)
	beritaAcaraPath := filepath.Join(beritaAcaraDir, beritaAcaraFilename)

	// Save files
	if err := saveFile(penilaianFile, penilaianPath); err != nil {
		sendErrorResponse("Error saving penilaian file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := saveFile(beritaAcaraFile, beritaAcaraPath); err != nil {
		os.Remove(penilaianPath) // cleanup
		sendErrorResponse("Error saving berita acara file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Database operation
	db, err := config.GetDB()
	if err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		sendErrorResponse("Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Use transaction for database operations
	tx, err := db.Begin()
	if err != nil {
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		sendErrorResponse("Transaction error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if record exists
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM seminar_proposal_penilaian WHERE user_id = ? AND dosen_id = ? AND final_proposal_id = ?)",
		userID, dosenID, finalProposalID).Scan(&exists)

	if err != nil {
		tx.Rollback()
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		sendErrorResponse("Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var query string
	if exists {
		query = `
			UPDATE seminar_proposal_penilaian 
			SET file_penilaian_path = ?,
				file_berita_acara_path = ?,
				submitted_at = NOW(),
				status_pengumpulan = 'sudah'
			WHERE user_id = ? AND dosen_id = ? AND final_proposal_id = ?
        `
		_, err = tx.Exec(query, penilaianPath, beritaAcaraPath, userID, dosenID, finalProposalID)
	} else {
		query = `
			INSERT INTO seminar_proposal_penilaian (
				user_id, final_proposal_id, dosen_id,
				file_penilaian_path, file_berita_acara_path,
				status_pengumpulan, submitted_at
			) VALUES (?, ?, ?, ?, ?, 'sudah', NOW())
        `
		_, err = tx.Exec(query, userID, finalProposalID, dosenID, penilaianPath, beritaAcaraPath)
	}

	if err != nil {
		tx.Rollback()
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		sendErrorResponse("Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		os.Remove(penilaianPath)
		os.Remove(beritaAcaraPath)
		sendErrorResponse("Transaction commit error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Penilaian proposal berhasil disimpan",
		"data": map[string]interface{}{
			"penilaian_path":    penilaianPath,
			"berita_acara_path": beritaAcaraPath,
		},
	})
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

func GetMonitoringPenilaianProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
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
			fp.id AS final_proposal_id,
			u.nama_lengkap AS nama_taruna,
			t.jurusan,
			fp.topik_penelitian,

			d_ketua.nama_lengkap AS ketua_penguji,
			d1.nama_lengkap AS penguji1,
			d2.nama_lengkap AS penguji2,

			-- Hitung kelengkapan berkas
			CASE
				WHEN COUNT(CASE WHEN spp.file_penilaian_path IS NOT NULL AND spp.file_berita_acara_path IS NOT NULL THEN 1 END) = 3 THEN 'Lengkap'
				ELSE 'Belum Lengkap'
			END AS status_kelengkapan

		FROM penguji_proposal pp
		JOIN final_proposal fp ON fp.id = pp.final_proposal_id
		JOIN users u ON u.id = pp.user_id
		JOIN taruna t ON t.user_id = u.id

		JOIN dosen d_ketua ON d_ketua.id = pp.ketua_penguji_id
		JOIN dosen d1 ON d1.id = pp.penguji_1_id
		JOIN dosen d2 ON d2.id = pp.penguji_2_id

		LEFT JOIN seminar_proposal_penilaian spp
			ON spp.final_proposal_id = pp.final_proposal_id

		GROUP BY
			fp.id, u.nama_lengkap, t.jurusan, fp.topik_penelitian,
			d_ketua.nama_lengkap, d1.nama_lengkap, d2.nama_lengkap
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
		FinalProposalID   int    `json:"final_proposal_id"`
		NamaTaruna        string `json:"nama_taruna"`
		Jurusan           string `json:"jurusan"`
		TopikPenelitian   string `json:"topik_penelitian"`
		KetuaPenguji      string `json:"ketua_penguji"`
		Penguji1          string `json:"penguji1"`
		Penguji2          string `json:"penguji2"`
		StatusKelengkapan string `json:"status_kelengkapan"`
	}

	var result []MonitoringData
	for rows.Next() {
		var m MonitoringData
		err := rows.Scan(
			&m.FinalProposalID,
			&m.NamaTaruna,
			&m.Jurusan,
			&m.TopikPenelitian,
			&m.KetuaPenguji,
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

// Helper function to handle null strings
func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func GetFinalProposalDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	finalProposalID := vars["id"]
	if finalProposalID == "" {
		http.Error(w, "Final Proposal ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query proposal dan taruna
	queryProposal := `
		SELECT fp.id, u.nama_lengkap, u.jurusan, u.kelas, fp.topik_penelitian
		FROM final_proposal fp
		JOIN users u ON u.id = fp.user_id
		WHERE fp.id = ?
	`
	var (
		idProposal      int
		namaTaruna      string
		jurusan         string
		kelas           string
		topikPenelitian string
	)
	err = db.QueryRow(queryProposal, finalProposalID).Scan(&idProposal, &namaTaruna, &jurusan, &kelas, &topikPenelitian)
	if err != nil {
		http.Error(w, "Proposal tidak ditemukan: "+err.Error(), http.StatusNotFound)
		return
	}

	// Query penguji
	queryPenguji := `
		SELECT ketua_penguji_id, penguji_1_id, penguji_2_id
		FROM penguji_proposal
		WHERE final_proposal_id = ?
		LIMIT 1
	`
	var ketuaID, penguji1ID, penguji2ID int
	err = db.QueryRow(queryPenguji, finalProposalID).Scan(&ketuaID, &penguji1ID, &penguji2ID)
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
			FROM seminar_proposal_penilaian
			WHERE final_proposal_id = ? AND dosen_id = ?
			LIMIT 1
		`, finalProposalID, dosenID).Scan(&status, &filePenilaian, &fileBA)

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
			"ketua_penguji":    getPenilaian(ketuaID),
			"penguji1":         getPenilaian(penguji1ID),
			"penguji2":         getPenilaian(penguji2ID),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Helper function untuk mendapatkan data penilaian dosen
func getPengujiData(db *sql.DB, proposalID string, dosenID int) map[string]interface{} {
	query := `
		SELECT 
			d.nama_lengkap,
			spp.file_penilaian_path,
			spp.file_berita_acara_path,
			spp.status_pengumpulan
		FROM dosen d
		LEFT JOIN seminar_proposal_penilaian spp ON spp.dosen_id = d.id 
			AND spp.seminar_proposal_id = ?
		WHERE d.id = ?
	`

	var data struct {
		NamaLengkap       string
		FilePenilaianPath sql.NullString
		BeritaAcaraPath   sql.NullString
		StatusPengumpulan sql.NullString
	}

	err := db.QueryRow(query, proposalID, dosenID).Scan(
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

func GetCatatanPerbaikanTarunaHandler(w http.ResponseWriter, r *http.Request) {
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
		FROM seminar_proposal_penilaian spp
		JOIN dosen d ON spp.dosen_id = d.id
		LEFT JOIN final_proposal fp ON spp.final_proposal_id = fp.id
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
