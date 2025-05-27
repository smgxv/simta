package handlers

import (
	"database/sql"
	"document_service/config"
	"document_service/entities"
	"document_service/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func UploadProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
	dosenID := r.FormValue("dosen_id")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	// Get the file from form
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/proposal"
	if err := os.MkdirAll(uploadDir, 0777); err != nil { // Ubah permission ke 0777
		http.Error(w, "Error creating upload directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Simpan file dengan nama asli
	filename := fmt.Sprintf("Proposal_%s_%s_%s",
		userID,
		time.Now().Format("20060102150405"),
		handler.Filename) // Gunakan nama asli file

	filePath := filepath.Join(uploadDir, filename)

	// Log untuk debugging
	log.Printf("Saving file to: %s", filePath)

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
	dosenIDInt, _ := strconv.Atoi(dosenID)

	// Create ICP record
	proposal := &entities.Proposal{
		UserID:          userIDInt,
		DosenID:         dosenIDInt,
		TopikPenelitian: topikPenelitian,
		Keterangan:      keterangan,
		FilePath:        filePath,
	}

	proposalModel := models.NewProposalModel(db)
	if err := proposalModel.Create(proposal); err != nil {
		// If database insert fails, delete the uploaded file
		os.Remove(filePath)
		http.Error(w, "Error saving to database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Proposal berhasil diunggah",
		"data": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

func GetProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	proposalModel := models.NewProposalModel(db)
	proposals, err := proposalModel.GetByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   proposals,
	})
}

func DownloadFileProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")

	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "File path is required", http.StatusBadRequest)
		return
	}

	// Buka file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set header untuk download
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	w.Header().Set("Content-Type", "application/pdf")

	// Copy file ke response
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error downloading file", http.StatusInternalServerError)
		return
	}
}

func GetProposalByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	// Get either dosen_id or taruna_id from query parameter
	dosenID := r.URL.Query().Get("dosen_id")
	tarunaID := r.URL.Query().Get("taruna_id")

	if dosenID == "" && tarunaID == "" {
		http.Error(w, "Either dosen_id or taruna_id is required", http.StatusBadRequest)
		return
	}

	// Ambil ID dari path parameter
	vars := mux.Vars(r)
	proposalID := vars["id"]
	if proposalID == "" {
		http.Error(w, "Proposal ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Check access rights
	var authorized bool
	var query string
	var args []interface{}

	if dosenID != "" {
		query = "SELECT EXISTS(SELECT 1 FROM proposal WHERE id = ? AND dosen_id = ?)"
		args = []interface{}{proposalID, dosenID}
	} else {
		// For taruna, we need to check against user_id since that's what we store in ICP table
		query = "SELECT EXISTS(SELECT 1 FROM proposal WHERE id = ? AND user_id = ?)"
		args = []interface{}{proposalID, tarunaID}
	}

	err = db.QueryRow(query, args...).Scan(&authorized)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !authorized {
		http.Error(w, "Unauthorized: You don't have access to this Proposal", http.StatusForbidden)
		return
	}

	// Get ICP details with joins to get names
	detailQuery := `
        SELECT 
            i.*,
            d.nama_lengkap as dosen_nama,
            t.nama_lengkap as nama_taruna
        FROM proposal i 
        LEFT JOIN dosen d ON i.dosen_id = d.id
        LEFT JOIN taruna t ON i.user_id = t.user_id
        WHERE i.id = ?
    `

	var proposal struct {
		ID              int            `json:"id"`
		UserID          int            `json:"user_id"`
		DosenID         int            `json:"dosen_id"`
		TopikPenelitian string         `json:"topik_penelitian"`
		Keterangan      string         `json:"keterangan"`
		FilePath        string         `json:"file_path"`
		Status          string         `json:"status"`
		CreatedAt       string         `json:"created_at"`
		UpdatedAt       string         `json:"updated_at"`
		DosenNama       sql.NullString `json:"dosen_nama"`
		NamaTaruna      sql.NullString `json:"nama_taruna"`
	}

	err = db.QueryRow(detailQuery, proposalID).Scan(
		&proposal.ID,
		&proposal.UserID,
		&proposal.DosenID,
		&proposal.TopikPenelitian,
		&proposal.Keterangan,
		&proposal.FilePath,
		&proposal.Status,
		&proposal.CreatedAt,
		&proposal.UpdatedAt,
		&proposal.DosenNama,
		&proposal.NamaTaruna,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Proposal not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert sql.NullString to string for JSON response
	response := map[string]interface{}{
		"id":               proposal.ID,
		"user_id":          proposal.UserID,
		"dosen_id":         proposal.DosenID,
		"topik_penelitian": proposal.TopikPenelitian,
		"keterangan":       proposal.Keterangan,
		"file_path":        proposal.FilePath,
		"status":           proposal.Status,
		"created_at":       proposal.CreatedAt,
		"updated_at":       proposal.UpdatedAt,
		"dosen_nama":       proposal.DosenNama.String,
		"nama_taruna":      proposal.NamaTaruna.String,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   response,
	})
}

func EditProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	dosenID := r.FormValue("dosen_id")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	proposalModel := models.NewProposalModel(db)
	proposal, err := proposalModel.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update fields
	proposal.DosenID, _ = strconv.Atoi(dosenID)
	proposal.TopikPenelitian = topikPenelitian
	proposal.Keterangan = keterangan

	// Handle file upload if new file is provided
	file, handler, err := r.FormFile("file")
	if err == nil {
		defer file.Close()

		uploadDir := "uploads/proposal"
		filename := fmt.Sprintf("Proposal_%d_%s_%s",
			proposal.UserID,
			time.Now().Format("20060102150405"),
			handler.Filename)

		filePath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete old file if exists
		if proposal.FilePath != "" {
			os.Remove(proposal.FilePath)
		}

		proposal.FilePath = filePath
	}

	if err := proposalModel.Update(proposal); err != nil {
		http.Error(w, "Error updating Proposal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Proposal berhasil diupdate",
	})
}
