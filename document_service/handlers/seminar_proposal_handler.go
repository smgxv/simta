package handlers

import (
	"database/sql"
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
)

func UploadSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ketuaPengujiID, err := strconv.Atoi(r.FormValue("ketua_penguji"))
	if err != nil {
		http.Error(w, "Invalid ketua penguji ID", http.StatusBadRequest)
		return
	}

	penguji1ID, err := strconv.Atoi(r.FormValue("penguji1"))
	if err != nil {
		http.Error(w, "Invalid penguji 1 ID", http.StatusBadRequest)
		return
	}

	penguji2ID, err := strconv.Atoi(r.FormValue("penguji2"))
	if err != nil {
		http.Error(w, "Invalid penguji 2 ID", http.StatusBadRequest)
		return
	}

	topikPenelitian := r.FormValue("topik_penelitian")
	if topikPenelitian == "" {
		http.Error(w, "Topik penelitian is required", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/seminar_proposal"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%d_%s_%s", userID, timestamp, handler.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Create file
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, file); err != nil {
		http.Error(w, "Error copying file", http.StatusInternalServerError)
		return
	}

	// Create seminar proposal record
	proposal := &entities.SeminarProposal{
		UserID:           userID,
		TopikPenelitian:  topikPenelitian,
		FileProposalPath: filepath,
		KetuaPengujiID:   ketuaPengujiID,
		Penguji1ID:       penguji1ID,
		Penguji2ID:       penguji2ID,
	}

	// Insert into database
	if err := models.InsertSeminarProposal(db, proposal); err != nil {
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Seminar proposal uploaded successfully",
	})
}

func GetSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user_id from query params
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get proposals from database
	proposals, err := models.GetSeminarProposalByUserID(db, userID)
	if err != nil {
		http.Error(w, "Error retrieving proposals", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   proposals,
	})
}

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}
