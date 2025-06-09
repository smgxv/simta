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

	// Ambil semua ICP dari tabel final_icp
	db := tarunaModel.GetDB() // tambahkan method GetDB() di TarunaModel jika belum ada
	icpRows, err := db.Query("SELECT topik_penelitian, status FROM final_icp WHERE user_id = ? ORDER BY created_at DESC", userId)
	icpList := []map[string]string{}
	if err == nil {
		defer icpRows.Close()
		for icpRows.Next() {
			var topik, status string
			if err := icpRows.Scan(&topik, &status); err == nil {
				icpList = append(icpList, map[string]string{
					"topik_penelitian": topik,
					"status":           status,
				})
			}
		}
	}
	data["icp_list"] = icpList

	// Ambil dosen pembimbing dari tabel dosbing_proposal (status aktif)
	dosenPembimbing := "-"
	dbQuery := `SELECT d.nama_lengkap FROM dosbing_proposal dp JOIN dosen d ON dp.dosen_id = d.id WHERE dp.user_id = ? AND dp.status = 'aktif' ORDER BY dp.tanggal_ditetapkan DESC LIMIT 1`
	dbRow := db.QueryRow(dbQuery, userId)
	var namaDosen string
	if err := dbRow.Scan(&namaDosen); err == nil {
		dosenPembimbing = namaDosen
	}
	data["dosen_pembimbing"] = dosenPembimbing

	json.NewEncoder(w).Encode(TarunaDashboardResponse{
		Status: "success",
		Data:   data,
	})
}
