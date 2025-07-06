package handlers

import (
	"database/sql"
	"document_service/config"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func GetTugasAkhirDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		http.Error(w, "Revisi Laporan100 ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Ambil data dari tabel revisi_laporan100
	query := `
		SELECT 
			nama_lengkap, jurusan, kelas, tahun_akademik, 
			topik_penelitian, abstrak_id, abstrak_en, kata_kunci, link_repo,
			file_path, file_produk_path, file_bap_path
		FROM revisi_laporan100
		WHERE id = ?
	`

	var (
		namaTaruna, jurusan, kelas, tahunAkademik,
		topikPenelitian, abstrakID, abstrakEN, kataKunci, linkRepo,
		filePath, fileProdukPath, fileBapPath sql.NullString
	)

	err = db.QueryRow(query, id).Scan(
		&namaTaruna, &jurusan, &kelas, &tahunAkademik,
		&topikPenelitian, &abstrakID, &abstrakEN, &kataKunci, &linkRepo,
		&filePath, &fileProdukPath, &fileBapPath,
	)

	if err != nil {
		http.Error(w, "Data tidak ditemukan: "+err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"nama_taruna":      nullToDash(namaTaruna),
			"jurusan":          nullToDash(jurusan),
			"kelas":            nullToDash(kelas),
			"tahun_akademik":   nullToDash(tahunAkademik),
			"topik_penelitian": nullToDash(topikPenelitian),
			"abstrak_id":       nullToDash(abstrakID),
			"abstrak_en":       nullToDash(abstrakEN),
			"kata_kunci":       nullToDash(kataKunci),
			"link_repo":        nullToDash(linkRepo),
			"file_dokumen":     nullToDash(filePath),
			"file_produk":      nullToDash(fileProdukPath),
			"file_bap":         nullToDash(fileBapPath),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Helper untuk mengganti NULL menjadi "-"
func nullToDash(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return "-"
}
