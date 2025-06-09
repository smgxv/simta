package handlers

import (
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

	json.NewEncoder(w).Encode(TarunaDashboardResponse{
		Status: "success",
		Data:   data,
	})
}
