package handlers

import (
	"document_service/config"
	"document_service/utils"
	"encoding/json"
	"net/http"
)

// GetTarunaTopicsHandler handles fetching taruna names and their research topics
func GetTarunaTopicsHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCors(&w)
	if r.Method == "OPTIONS" {
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Error connecting to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var results []struct {
		UserID          int    `json:"user_id"`
		NamaLengkap     string `json:"nama_lengkap"`
		TopikPenelitian string `json:"topik_penelitian"`
	}

	query := `SELECT 
		t.user_id,
		t.nama_lengkap,
		COALESCE(f.topik_penelitian, '') as topik_penelitian
	FROM taruna t
	LEFT JOIN final_proposal f ON t.user_id = f.user_id
	ORDER BY t.nama_lengkap ASC`

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Error fetching data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result struct {
			UserID          int    `json:"user_id"`
			NamaLengkap     string `json:"nama_lengkap"`
			TopikPenelitian string `json:"topik_penelitian"`
		}
		if err := rows.Scan(&result.UserID, &result.NamaLengkap, &result.TopikPenelitian); err != nil {
			http.Error(w, "Error scanning data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating rows: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
