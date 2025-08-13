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

type CatatanPerbaikan struct {
	ID              int    `json:"id"`
	NamaDosen       string `json:"nama_dosen"`
	TopikPenelitian string `json:"topik_penelitian"`
	FilePath        string `json:"file_path"`
	SubmittedAt     string `json:"submitted_at"`
}

// GetSeminarProposalByDosenHandler menangani request untuk mendapatkan data seminar proposal berdasarkan ID dosen
func GetSeminarProposalByDosenHandler(w http.ResponseWriter, r *http.Request) {
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

	// Ambil final proposal yang diuji oleh dosen_id, sertakan file_pendukung_path
	query := `
		SELECT 
			fp.id, 
			fp.user_id, 
			fp.topik_penelitian, 
			fp.file_path, 
			COALESCE(fp.file_pendukung_path, '') AS file_pendukung_path,
			u.nama_lengkap
		FROM final_proposal fp
		JOIN users u ON fp.user_id = u.id
		JOIN penguji_proposal pp ON fp.id = pp.final_proposal_id
		WHERE pp.ketua_penguji_id = ? 
		   OR pp.penguji_1_id = ? 
		   OR pp.penguji_2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FinalProposalData struct {
		FinalProposalID   int    `json:"final_proposal_id"`
		UserID            int    `json:"user_id"`
		TopikPenelitian   string `json:"topik_penelitian"`
		FilePath          string `json:"file_path"`
		FilePendukungPath string `json:"file_pendukung_path"` // tambahkan field ini
		TarunaNama        string `json:"taruna_nama"`
	}

	var data []FinalProposalData
	for rows.Next() {
		var item FinalProposalData
		err := rows.Scan(
			&item.FinalProposalID,
			&item.UserID,
			&item.TopikPenelitian,
			&item.FilePath,
			&item.FilePendukungPath,
			&item.TarunaNama,
		)
		if err != nil {
			http.Error(w, "Error scanning data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data = append(data, item)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// GetTarunaListForDosenHandler menangani request untuk mendapatkan daftar taruna yang belum memiliki final proposal
func GetSeminarProposalTarunaListForDosenHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	sendError := func(message string, statusCode int) {
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": message,
		})
	}

	if r.ContentLength > filemanager.MaxFileSize {
		sendError("Ukuran file terlalu besar. Maksimal 15MB.", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(filemanager.MaxFileSize); err != nil {
		sendError("Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	dosenID := r.FormValue("dosen_id")
	finalProposalID := r.FormValue("final_proposal_id")

	if userID == "" || dosenID == "" || finalProposalID == "" {
		sendError("user_id, dosen_id, dan final_proposal_id harus diisi", http.StatusBadRequest)
		return
	}

	// === Penilaian File ===
	catatanperbaikanFile, catatanperbaikanHeader, err := r.FormFile("catatanperbaikan_file")
	if err != nil {
		sendError("Gagal mengambil file Catatan Perbaikan: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer catatanperbaikanFile.Close()

	if err := filemanager.ValidateFileType(catatanperbaikanFile, catatanperbaikanHeader.Filename); err != nil {
		sendError(err.Error(), http.StatusBadRequest)
		return
	}
	catatanperbaikanFile.Seek(0, 0)
	catatanperbaikanFilename := fmt.Sprintf("Catatan_Perbaikan_%s_%s_%s",
		userID,
		time.Now().Format("20060102150405"),
		filemanager.ValidateFileName(catatanperbaikanHeader.Filename))
	catatanperbaikanPath, err := filemanager.SaveUploadedFile(catatanperbaikanFile, catatanperbaikanHeader, "uploads/catatanperbaikan_proposal", catatanperbaikanFilename)
	if err != nil {
		sendError("Gagal menyimpan file catatan perbaikan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// === Penilaian File ===
	penilaianFile, penilaianHeader, err := r.FormFile("penilaian_file")
	if err != nil {
		os.Remove(catatanperbaikanPath)
		sendError("Gagal mengambil file penilaian: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer penilaianFile.Close()

	if err := filemanager.ValidateFileType(penilaianFile, penilaianHeader.Filename); err != nil {
		os.Remove(catatanperbaikanPath)
		sendError(err.Error(), http.StatusBadRequest)
		return
	}
	penilaianFile.Seek(0, 0)
	penilaianFilename := fmt.Sprintf("File_Penilaian_%s_%s_%s",
		userID,
		time.Now().Format("20060102150405"),
		filemanager.ValidateFileName(penilaianHeader.Filename))
	penilaianPath, err := filemanager.SaveUploadedFile(penilaianFile, penilaianHeader, "uploads/file_penilaian_proposal", penilaianFilename)
	if err != nil {
		os.Remove(catatanperbaikanPath)
		sendError("Gagal menyimpan file penilaian: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// === Database ===
	db, err := config.GetDB()
	if err != nil {
		os.Remove(catatanperbaikanPath)
		os.Remove(penilaianPath)
		sendError("Koneksi database gagal: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		os.Remove(catatanperbaikanPath)
		os.Remove(penilaianPath)
		sendError("Gagal memulai transaksi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(
		SELECT 1 FROM seminar_proposal_penilaian 
		WHERE user_id = ? AND dosen_id = ? AND final_proposal_id = ?
	)`, userID, dosenID, finalProposalID).Scan(&exists)
	if err != nil {
		tx.Rollback()
		os.Remove(catatanperbaikanPath)
		os.Remove(penilaianPath)
		sendError("Gagal cek data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if exists {
		_, err = tx.Exec(`UPDATE seminar_proposal_penilaian 
			SET file_catatanperbaikan_path = ?, file_penilaian_path = ?, 
				submitted_at = NOW(), status_pengumpulan = 'sudah' 
			WHERE user_id = ? AND dosen_id = ? AND final_proposal_id = ?`,
			catatanperbaikanPath, penilaianPath, userID, dosenID, finalProposalID)
	} else {
		_, err = tx.Exec(`INSERT INTO seminar_proposal_penilaian (
			user_id, final_proposal_id, dosen_id,
			file_catatanperbaikan_path, file_penilaian_path,
			status_pengumpulan, submitted_at
		) VALUES (?, ?, ?, ?, ?, 'sudah', NOW())`,
			userID, finalProposalID, dosenID, catatanperbaikanPath, penilaianPath)
	}
	if err != nil {
		tx.Rollback()
		os.Remove(catatanperbaikanPath)
		os.Remove(penilaianPath)
		sendError("Gagal menyimpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		os.Remove(catatanperbaikanPath)
		os.Remove(penilaianPath)
		sendError("Gagal commit data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Penilaian proposal berhasil disimpan",
		"data": map[string]interface{}{
			"catatanperbaikan_path": catatanperbaikanPath,
			"penilaian_path":        penilaianPath,
		},
	})
}

// DownloadFileCatatanPerbaikanProposalHandler digunakan untuk mengunduh file penilaian atau berita acara laporan 70%
func DownloadFileCatatanPerbaikanProposalHandler(w http.ResponseWriter, r *http.Request) {
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
	if strings.HasPrefix(fileName, "Catatan_Perbaikan_") {
		baseDir = "uploads/catatanperbaikan_proposal"
	} else if strings.HasPrefix(fileName, "File_Penilaian_") {
		baseDir = "uploads/penilaian_proposal"
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
func saveProposal(src io.Reader, destPath string) error {
	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func GetMonitoringPenilaianProposalHandler(w http.ResponseWriter, r *http.Request) {
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
			fp.id AS final_proposal_id,
			u.nama_lengkap AS nama_taruna,
			t.jurusan,
			fp.topik_penelitian,

			d_ketua.nama_lengkap AS ketua_penguji,
			d1.nama_lengkap AS penguji1,
			d2.nama_lengkap AS penguji2,

			-- Hitung kelengkapan berkas
			CASE
				WHEN COUNT(CASE WHEN spp.file_catatanperbaikan_path IS NOT NULL AND spp.file_penilaian_path IS NOT NULL THEN 1 END) = 3 THEN 'Lengkap'
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

func GetFinalProposalDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
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
			SELECT status_pengumpulan, file_catatanperbaikan_path, file_penilaian_path
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
			spp.file_catatanperbaikan_path,
			spp.file_penilaian_path,
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
			"nama_lengkap":               "-",
			"file_catatanperbaikan_path": "",
			"file_penilaian_path":        "",
			"status_pengumpulan":         "belum",
		}
	}

	return map[string]interface{}{
		"nama_lengkap":               data.NamaLengkap,
		"file_catatanperbaikan_path": data.FilePenilaianPath.String,
		"file_penilaian_path":        data.BeritaAcaraPath.String,
		"status_pengumpulan":         data.StatusPengumpulan.String,
	}
}

func GetCatatanPerbaikanTarunaProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "https://securesimta.my.id")
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
		SELECT spp.id, d.nama_lengkap, fp.topik_penelitian, spp.file_catatanperbaikan_path, spp.submitted_at
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
