package handlers

import (
	"encoding/json"
	"net/http"
	"user_service/models"
)

func GetAllTaruna(w http.ResponseWriter, r *http.Request) {
	// Set header CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
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

	tarunaModel, err := models.NewTarunaModel()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	tarunas, err := tarunaModel.GetAllTaruna()
	if err != nil {
		http.Error(w, "Failed to fetch taruna data", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   tarunas,
	})
}

// Edit user taruna
// s
