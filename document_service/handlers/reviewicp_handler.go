package handlers

import (
	"document_service/config"
	"document_service/entities"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// Tambahkan handler baru
func GetICPByDosenIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	// Ambil dosen ID dari path parameter
	vars := mux.Vars(r)
	dosenID := vars["dosen_id"]
	if dosenID == "" {
		http.Error(w, "Dosen ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT 
			i.id, i.user_id, i.dosen_id, i.topik_penelitian,
			i.keterangan, i.file_path, i.status, i.created_at,
			i.updated_at, d.nama_lengkap as dosen_nama,
			t.nama_lengkap as nama_taruna, t.kelas
		FROM icp i
		LEFT JOIN dosen d ON i.dosen_id = d.id
		LEFT JOIN taruna t ON i.user_id = t.id
		WHERE i.dosen_id = ?
	`
	rows, err := db.Query(query, dosenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var icps []entities.ICP
	for rows.Next() {
		var icp entities.ICP
		err := rows.Scan(
			&icp.ID,
			&icp.UserID,
			&icp.DosenID,
			&icp.TopikPenelitian,
			&icp.Keterangan,
			&icp.FilePath,
			&icp.Status,
			&icp.CreatedAt,
			&icp.UpdatedAt,
			&icp.DosenNama,
			&icp.NamaTaruna,
			&icp.Kelas,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		icps = append(icps, icp)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   icps,
	})
}
