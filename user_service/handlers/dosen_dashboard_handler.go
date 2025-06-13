package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user_service/models"
)

type DosenDashboardResponse struct {
	NamaLengkap string `json:"nama_lengkap"`
	Jurusan     string `json:"jurusan"`
}

func DosenDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// CORS setup
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil userId dari query param (atau bisa juga pakai header kalau frontend mengirim lewat itu)
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	dosenModel, err := models.NewDosenModel()
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dosen, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		http.Error(w, "Dosen not found", http.StatusNotFound)
		return
	}

	resp := DosenDashboardResponse{
		NamaLengkap: dosen.NamaLengkap,
		Jurusan:     dosen.Jurusan,
	}
	json.NewEncoder(w).Encode(resp)
}
