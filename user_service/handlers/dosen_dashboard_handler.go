package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user_service/models"
)

type DosenDashboardResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// Pastikan dosenModel sudah diinisialisasi di main.go dan diimport/diakses di sini
var DosenModelInstance *models.DosenModel

func DosenDashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		json.NewEncoder(w).Encode(DosenDashboardResponse{
			Status: "error",
			Error:  "Unauthorized",
		})
		return
	}
	userID, err := strconv.Atoi(userIDVal.(string))
	if err != nil {
		json.NewEncoder(w).Encode(DosenDashboardResponse{
			Status: "error",
			Error:  "Invalid user_id",
		})
		return
	}

	dosen, err := DosenModelInstance.GetDosenByUserID(userID)
	if err != nil {
		json.NewEncoder(w).Encode(DosenDashboardResponse{
			Status: "error",
			Error:  "Dosen not found",
		})
		return
	}

	resp := map[string]interface{}{
		"nama":    dosen.NamaLengkap,
		"jurusan": dosen.Jurusan,
	}

	json.NewEncoder(w).Encode(DosenDashboardResponse{
		Status: "success",
		Data:   resp,
	})
}
