package handlers

import (
	"document_service/config"
	"document_service/models"
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

	icpModel := models.NewICPModel(db)
	icps, err := icpModel.GetByDosenID(dosenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   icps,
	})
}
