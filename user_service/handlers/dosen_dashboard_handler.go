package handlers

import (
	"encoding/json"
	"net/http"
	"user_service/models"
	"user_service/utils"
)

func DosenDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID := r.Context().Value("user_id").(string)

	// Get dosen data from database
	dosen, err := models.GetDosenByUserID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get dosen data")
		return
	}

	// Create response
	response := struct {
		Status string       `json:"status"`
		Data   models.Dosen `json:"data"`
	}{
		Status: "success",
		Data:   dosen,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
