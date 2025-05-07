package handlers

import (
	"encoding/json"
	"net/http"
	"user_service/models"
)

func GetAllDosen(w http.ResponseWriter, r *http.Request) {
	// Set header CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dosenModel, err := models.NewDosenModel()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	dosens, err := dosenModel.GetAllDosen()
	if err != nil {
		http.Error(w, "Failed to fetch dosen data", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   dosens,
	})
}
