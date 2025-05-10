package handlers

import (
	"document_service/models"
	"encoding/json"
	"net/http"
)

func GetICPByDosenIDHandler(model *models.ICPModel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dosenID := r.URL.Query().Get("dosen_id")
		if dosenID == "" {
			http.Error(w, "dosen_id is required", http.StatusBadRequest)
			return
		}
		icps, err := model.GetByDosenID(dosenID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   icps,
		})
	}
}
