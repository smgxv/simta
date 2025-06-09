package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"user_service/models"
)

type TarunaDashboardResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func TarunaDashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIdStr := r.URL.Query().Get("user_id")
	if userIdStr == "" {
		json.NewEncoder(w).Encode(TarunaDashboardResponse{
			Status: "error",
			Error:  "user_id is required",
		})
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(TarunaDashboardResponse{
			Status: "error",
			Error:  "user_id must be integer",
		})
		return
	}

	tarunaModel, err := models.NewTarunaModel()
	if err != nil {
		json.NewEncoder(w).Encode(TarunaDashboardResponse{
			Status: "error",
			Error:  "Database connection error",
		})
		return
	}

	data, err := tarunaModel.GetTarunaByUserID(userId)
	if err != nil {
		json.NewEncoder(w).Encode(TarunaDashboardResponse{
			Status: "error",
			Error:  "Taruna not found",
		})
		return
	}

	// Ambil status ICP dari tabel final_icp
	db := tarunaModel.GetDB() // tambahkan method GetDB() di TarunaModel jika belum ada
	var statusICP string
	err = db.QueryRow("SELECT status FROM final_icp WHERE user_id = ? ORDER BY created_at DESC LIMIT 1", userId).Scan(&statusICP)
	if err == sql.ErrNoRows {
		statusICP = "Belum mengumpulkan ICP"
	} else if err != nil {
		statusICP = "Error mengambil status ICP"
	}
	data["status_icp"] = statusICP

	json.NewEncoder(w).Encode(TarunaDashboardResponse{
		Status: "success",
		Data:   data,
	})
}
