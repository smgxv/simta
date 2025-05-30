package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"user_service/config"
	"user_service/entities"
	"user_service/models"
)

func AssignDosbingProposal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload entities.DosbingProposal
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	model, err := models.NewDosbingModel()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := model.AssignPembimbing(&payload); err != nil {
		http.Error(w, "Failed to assign pembimbing", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Dosen pembimbing berhasil disimpan",
	})
}

func GetTarunaWithDosbing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db, err := config.ConnectDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	query := `
		SELECT t.id, t.nama_lengkap, t.jurusan, t.kelas, d.nama_lengkap AS dosen_pembimbing
		FROM taruna t
		JOIN users u ON t.user_id = u.id
		LEFT JOIN dosbing_proposal dp ON dp.user_id = u.id
		LEFT JOIN users d ON dp.dosen_id = d.id
		WHERE u.role = 'Taruna'
	`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []map[string]interface{}
	for rows.Next() {
		var id int
		var nama, jurusan, kelas, dosen sql.NullString

		if err := rows.Scan(&id, &nama, &jurusan, &kelas, &dosen); err != nil {
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}

		result = append(result, map[string]interface{}{
			"id":               id,
			"nama_lengkap":     nama.String,
			"jurusan":          jurusan.String,
			"kelas":            kelas.String,
			"dosen_pembimbing": dosen.String,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   result,
	})
}
