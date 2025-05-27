package handlers

import (
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func UploadSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get form values
	userID := r.FormValue("user_id")
	ketuaID := r.FormValue("ketua_penguji")
	penguji1ID := r.FormValue("penguji1")
	penguji2ID := r.FormValue("penguji2")
	topikPenelitian := r.FormValue("topik_penelitian")

	// Validate
	if userID == "" || ketuaID == "" || penguji1ID == "" || penguji2ID == "" || topikPenelitian == "" {
		http.Error(w, "Semua field wajib diisi", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/seminar_proposal"
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		http.Error(w, "Error creating upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Simpan file dengan nama asli
	filename := fmt.Sprintf("SeminarProposal_%s_%s_%s",
		userID,
		time.Now().Format("20060102150405"),
		handler.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save to database
	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Konversi string ke int
	userIDInt, _ := strconv.Atoi(userID)
	ketuaIDInt, _ := strconv.Atoi(ketuaID)
	penguji1IDInt, _ := strconv.Atoi(penguji1ID)
	penguji2IDInt, _ := strconv.Atoi(penguji2ID)

	// Simpan entitas
	seminarProposal := &entities.SeminarProposal{
		UserID:          userIDInt,
		KetuaPengujiID:  ketuaIDInt,
		Penguji1ID:      penguji1IDInt,
		Penguji2ID:      penguji2IDInt,
		TopikPenelitian: topikPenelitian,
		FilePath:        filePath,
		Status:          "pending",
	}

	model := models.NewSeminarProposalModel(db)
	if err := model.Create(seminarProposal); err != nil {
		os.Remove(filePath)
		http.Error(w, "Error saving to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Seminar proposal berhasil diunggah",
		"data": map[string]interface{}{
			"file_path": filePath,
		},
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
