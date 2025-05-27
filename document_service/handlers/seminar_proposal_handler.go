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
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func UploadSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form with larger size limit
	err := r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get form values
	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	ketuaPengujiID := r.FormValue("ketua_penguji")
	if ketuaPengujiID == "" {
		http.Error(w, "ketua_penguji is required", http.StatusBadRequest)
		return
	}

	penguji1ID := r.FormValue("penguji1")
	if penguji1ID == "" {
		http.Error(w, "penguji1 is required", http.StatusBadRequest)
		return
	}

	penguji2ID := r.FormValue("penguji2")
	if penguji2ID == "" {
		http.Error(w, "penguji2 is required", http.StatusBadRequest)
		return
	}

	topikPenelitian := r.FormValue("topik_penelitian")
	if topikPenelitian == "" {
		http.Error(w, "topik_penelitian is required", http.StatusBadRequest)
		return
	}

	// Get the file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(handler.Filename), ".pdf") {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "./uploads/seminar_proposal"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		http.Error(w, "Failed to create uploads directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s_%s", userID, timestamp, handler.Filename)
	filePath := filepath.Join(uploadsDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath) // Clean up on error
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Connect to database
	db, err := config.GetDB()
	if err != nil {
		os.Remove(filePath) // Clean up on error
		http.Error(w, "Database connection error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create seminar proposal record
	seminarProposalModel := models.NewSeminarProposalModel(db)
	userIDInt := utils.ParseInt(userID)
	ketuaPengujiIDInt := utils.ParseInt(ketuaPengujiID)
	penguji1IDInt := utils.ParseInt(penguji1ID)
	penguji2IDInt := utils.ParseInt(penguji2ID)

	seminarProposal := &entities.SeminarProposal{
		UserID:          userIDInt,
		KetuaPengujiID:  ketuaPengujiIDInt,
		Penguji1ID:      penguji1IDInt,
		Penguji2ID:      penguji2IDInt,
		TopikPenelitian: topikPenelitian,
		FilePath:        filePath,
	}

	if err := seminarProposalModel.Create(seminarProposal); err != nil {
		os.Remove(filePath) // Clean up on error
		http.Error(w, "Failed to save seminar proposal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Seminar proposal berhasil diupload",
		"data":    seminarProposal,
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
