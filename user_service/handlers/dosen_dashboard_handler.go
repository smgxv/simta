package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user_service/models"
	"user_service/utils"
)

// DosenDashboardResponseData merepresentasikan struktur data yang dikirim ke frontend.
type DosenDashboardResponseData struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	Nama    string `json:"nama"`
	Jurusan string `json:"jurusan"`
}

func DosenDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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
	dosenEntity, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get dosen data")
		return
	}

	// Buat data respons dengan tag JSON yang diinginkan frontend
	responseData := DosenDashboardResponseData{
		ID:      dosenEntity.ID,
		UserID:  dosenEntity.UserID,
		Nama:    dosenEntity.NamaLengkap,
		Jurusan: dosenEntity.Jurusan,
	}

	// Create response
	response := struct {
		Status string                     `json:"status"`
		Data   DosenDashboardResponseData `json:"data"`
	}{
		Status: "success",
		Data:   responseData,
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
