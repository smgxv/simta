package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user_service/models"
	"user_service/utils"
)

func DosenDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userIDStr := r.Context().Value("user_id").(string)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Initialize DosenModel
	dosenModel, err := models.NewDosenModel()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to initialize dosen model")
		return
	}

	// Get dosen data from database
	dosen, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get dosen data")
		return
	}

	// Create response
	response := struct {
		Status string      `json:"status"`
		Data   interface{} `json:"data"`
	}{
		Status: "success",
		Data:   dosen,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
