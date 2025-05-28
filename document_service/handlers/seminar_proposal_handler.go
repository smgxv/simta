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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		sendErrorResponse(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get form values
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		sendErrorResponse(w, "Invalid user ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	ketuaPengujiID, err := strconv.Atoi(r.FormValue("ketua_penguji"))
	if err != nil {
		sendErrorResponse(w, "Invalid ketua penguji ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	penguji1ID, err := strconv.Atoi(r.FormValue("penguji1"))
	if err != nil {
		sendErrorResponse(w, "Invalid penguji 1 ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	penguji2ID, err := strconv.Atoi(r.FormValue("penguji2"))
	if err != nil {
		sendErrorResponse(w, "Invalid penguji 2 ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	topikPenelitian := r.FormValue("topik_penelitian")
	if topikPenelitian == "" {
		sendErrorResponse(w, "Topik penelitian is required", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		sendErrorResponse(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/seminar_proposal"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		sendErrorResponse(w, "Error creating upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%d_%s_%s", userID, timestamp, handler.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Create file
	dst, err := os.Create(filepath)
	if err != nil {
		sendErrorResponse(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, file); err != nil {
		sendErrorResponse(w, "Error copying file: "+err.Error(), http.StatusInternalServerError)
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
		sendErrorResponse(w, "Error saving to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	sendSuccessResponse(w, "Seminar proposal uploaded successfully", nil)
}

func GetSeminarProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user_id from query params
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		sendErrorResponse(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		sendErrorResponse(w, "Invalid user ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get proposals from database
	proposals, err := models.GetSeminarProposalByUserID(db, userID)
	if err != nil {
		sendErrorResponse(w, "Error retrieving proposals: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	sendSuccessResponse(w, "Seminar proposals retrieved successfully", proposals)
}

// Helper functions for consistent response format
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}

func sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

var db *sql.DB

func InitDB(database *sql.DB) {
	db = database
}
