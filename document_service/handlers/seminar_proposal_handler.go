package handlers

import (
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"document_service/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
)

func UploadSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse form
	err := r.ParseMultipartForm(10 << 20) // 10MB max
	if err != nil {
		http.Error(w, "Gagal parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ambil data form
	userID := utils.ParseInt(r.FormValue("user_id"))
	ketua := utils.ParseInt(r.FormValue("ketua_penguji"))
	p1 := utils.ParseInt(r.FormValue("penguji1"))
	p2 := utils.ParseInt(r.FormValue("penguji2"))
	topik := r.FormValue("topik_penelitian")

	// Validasi dasar
	if userID == 0 || ketua == 0 || p1 == 0 || p2 == 0 || topik == "" {
		http.Error(w, "Semua field wajib diisi", http.StatusBadRequest)
		return
	}

	// Upload file
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Gagal ambil file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if filepath.Ext(handler.Filename) != ".pdf" {
		http.Error(w, "File harus PDF", http.StatusBadRequest)
		return
	}

	saveDir := "./uploads/seminar_proposal"
	os.MkdirAll(saveDir, os.ModePerm)

	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().Unix(), handler.Filename)
	filepath := filepath.Join(saveDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Gagal simpan file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	// Simpan ke DB
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Gagal konek database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	model := models.NewSeminarProposalModel(db)
	record := &entities.SeminarProposal{
		UserID:          userID,
		KetuaPengujiID:  ketua,
		Penguji1ID:      p1,
		Penguji2ID:      p2,
		TopikPenelitian: topik,
		FilePath:        filepath,
		Status:          "pending",
	}

	if err := model.Create(record); err != nil {
		http.Error(w, "Gagal simpan data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Seminar proposal berhasil diupload",
		"data":    record,
	})
}

func GetSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Verify token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header is required", http.StatusUnauthorized)
		return
	}

	// Get user_id from query params
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	// Connect to database
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database connection error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get seminar proposal data
	seminarProposalModel := models.NewSeminarProposalModel(db)
	seminarProposals, err := seminarProposalModel.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Failed to get seminar proposals: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   seminarProposals,
	})
}

func RegisterSeminarProposalRoutes(router *mux.Router) {
	router.HandleFunc("/seminarproposal/upload", UploadSeminarProposalHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/seminarproposal", GetSeminarProposalHandler).Methods("GET", "OPTIONS")
}
