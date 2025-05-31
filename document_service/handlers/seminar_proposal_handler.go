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

// GetTarunaListForDosenHandler menangani request untuk mendapatkan list taruna untuk dropdown penilaian proposal
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
			fp.topik_penelitian
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
	}

	var tarunaList []TarunaData
	for rows.Next() {
		var t TarunaData
		err := rows.Scan(&t.UserID, &t.NamaLengkap, &t.TopikPenelitian)
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
			user_id, dosen_id,
			file_penilaian_path, file_berita_acara_path,
			status_pengumpulan, submitted_at
		) VALUES (?, ?, ?, ?, 'belum', NOW())
	`
	_, err = db.Exec(query,
		userID,
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

// GetMonitoringPenilaianProposalHandler menangani request untuk mendapatkan data monitoring penilaian proposal
func GetMonitoringPenilaianProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Buka koneksi database
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query untuk mendapatkan data monitoring
	query := `
		SELECT 
			sp.id as seminar_proposal_id,
			u.nama_lengkap AS nama_taruna,
			t.jurusan,
			sp.topik_penelitian,
			
			d_ketua.nama_lengkap AS ketua_penguji,
			d1.nama_lengkap AS penguji1,
			d2.nama_lengkap AS penguji2,

			-- Status kelengkapan berkas
			CASE
				WHEN COUNT(CASE WHEN spp.status_pengumpulan = 'sudah' THEN 1 END) = 3 THEN 'Lengkap'
				ELSE 'Belum Lengkap'
			END AS status_kelengkapan

		FROM seminar_proposal sp

		JOIN users u ON u.id = sp.user_id
		JOIN taruna t ON t.user_id = u.id

		JOIN dosen d_ketua ON d_ketua.id = sp.ketua_penguji_id
		JOIN dosen d1 ON d1.id = sp.penguji1_id
		JOIN dosen d2 ON d2.id = sp.penguji2_id

		LEFT JOIN seminar_proposal_penilaian spp ON spp.seminar_proposal_id = sp.id

		GROUP BY 
			sp.id, u.nama_lengkap, t.jurusan, sp.topik_penelitian,
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
		SeminarProposalID int    `json:"seminar_proposal_id"`
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
			&m.SeminarProposalID,
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

// GetSeminarProposalDetailHandler menangani request untuk mendapatkan detail berkas seminar proposal
func GetSeminarProposalDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	proposalID := vars["id"]

	if proposalID == "" {
		http.Error(w, "Proposal ID is required", http.StatusBadRequest)
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
			sp.id, sp.user_id, t.nama_lengkap, t.jurusan, t.kelas,
			sp.topik_penelitian,
			sp.ketua_penguji_id, d_ketua.nama_lengkap,
			sp.penguji1_id, d1.nama_lengkap,
			sp.penguji2_id, d2.nama_lengkap
		FROM seminar_proposal sp
		JOIN taruna t ON t.user_id = sp.user_id
		JOIN dosen d_ketua ON d_ketua.id = sp.ketua_penguji_id
		JOIN dosen d1 ON d1.id = sp.penguji1_id
		JOIN dosen d2 ON d2.id = sp.penguji2_id
		WHERE sp.id = ?
	`

	var proposal struct {
		ID              int    `json:"id"`
		UserID          int    `json:"user_id"`
		NamaTaruna      string `json:"nama_taruna"`
		Jurusan         string `json:"jurusan"`
		Kelas           string `json:"kelas"`
		TopikPenelitian string `json:"topik_penelitian"`
		KetuaID         int    `json:"ketua_penguji_id"`
		KetuaNama       string `json:"ketua_penguji_nama"`
		Penguji1ID      int    `json:"penguji1_id"`
		Penguji1Nama    string `json:"penguji1_nama"`
		Penguji2ID      int    `json:"penguji2_id"`
		Penguji2Nama    string `json:"penguji2_nama"`
	}

	err = db.QueryRow(query, proposalID).Scan(
		&proposal.ID, &proposal.UserID, &proposal.NamaTaruna, &proposal.Jurusan, &proposal.Kelas,
		&proposal.TopikPenelitian,
		&proposal.KetuaID, &proposal.KetuaNama,
		&proposal.Penguji1ID, &proposal.Penguji1Nama,
		&proposal.Penguji2ID, &proposal.Penguji2Nama,
	)
	if err != nil {
		http.Error(w, "Error getting proposal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type Penilaian struct {
		NamaDosen         string `json:"nama_dosen"`
		StatusPengumpulan string `json:"status_pengumpulan"`
		FilePenilaian     string `json:"file_penilaian"`
		FileBeritaAcara   string `json:"file_berita_acara"`
	}

	// Ambil penilaian dari masing-masing penguji
	getPenilaian := func(dosenID int) Penilaian {
		var p Penilaian
		query := `
			SELECT d.nama_lengkap, spp.status_pengumpulan, spp.file_penilaian_path, spp.file_berita_acara_path
			FROM seminar_proposal_penilaian spp
			JOIN dosen d ON d.id = spp.dosen_id
			WHERE spp.seminar_proposal_id = ? AND spp.dosen_id = ?
			LIMIT 1
		`
		err := db.QueryRow(query, proposalID, dosenID).Scan(
			&p.NamaDosen, &p.StatusPengumpulan, &p.FilePenilaian, &p.FileBeritaAcara,
		)
		if err != nil {
			// jika belum ada data, tetap kembalikan nama dosen
			p.NamaDosen = "-"
			p.StatusPengumpulan = "belum"
			p.FilePenilaian = ""
			p.FileBeritaAcara = ""
		}
		return p
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"nama_taruna":      proposal.NamaTaruna,
			"jurusan":          proposal.Jurusan,
			"kelas":            proposal.Kelas,
			"topik_penelitian": proposal.TopikPenelitian,
			"ketua_penguji":    getPenilaian(proposal.KetuaID),
			"penguji1":         getPenilaian(proposal.Penguji1ID),
			"penguji2":         getPenilaian(proposal.Penguji2ID),
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
