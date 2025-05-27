package handlers

import (
	"database/sql"
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
)

// Handler untuk mengambil ICP berdasarkan dosen_id
func GetProposalByDosenIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	dosenID := r.URL.Query().Get("dosen_id")
	if dosenID == "" {
		http.Error(w, "Dosen ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	proposalModel := models.NewProposalModel(db)
	proposals, err := proposalModel.GetByDosenID(dosenID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   proposals,
	})
}

// Handler untuk mengubah status ICP
func UpdateProposalStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	status := r.URL.Query().Get("status")
	if id == "" || status == "" {
		http.Error(w, "ID dan status diperlukan", http.StatusBadRequest)
		return
	}

	// Validasi status hanya boleh "approved", "rejected", atau "on review"
	if status != "approved" && status != "rejected" && status != "on review" {
		http.Error(w, "Status tidak valid", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	_, err = db.Exec("UPDATE icp SET status = ? WHERE id = ?", status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := "ICP berhasil diupdate"
	if status == "approved" {
		msg = "ICP berhasil di-approve"
	} else if status == "rejected" {
		msg = "ICP berhasil di-reject"
	} else if status == "on review" {
		msg = "ICP berhasil diubah ke status review"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": msg,
	})
}

// Handler untuk mengupload review ICP oleh dosen ke table review_icp
func UploadReviewProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error parsing form: " + err.Error(),
		})
		return
	}

	dosenID := r.FormValue("dosen_id")
	tarunaID := r.FormValue("taruna_id")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	if dosenID == "" || tarunaID == "" || topikPenelitian == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Missing required fields",
		})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error retrieving file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	uploadDir := "uploads/reviewicp"
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating upload directory: " + err.Error(),
		})
		return
	}

	// Determine if this is a dosen review or taruna revision based on the request origin
	isDosenReview := r.Header.Get("X-User-Role") == "dosen"
	filePrefix := "REVIEW_ICP"
	if !isDosenReview {
		filePrefix = "REVISI_ICP"
	}

	filename := fmt.Sprintf("%s_%s_%s_%s",
		filePrefix,
		dosenID,
		time.Now().Format("20060102150405"),
		handler.Filename)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating file: " + err.Error(),
		})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving file: " + err.Error(),
		})
		return
	}

	db, err := config.GetDB()
	if err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to start transaction: " + err.Error(),
		})
		return
	}

	// Update status ICP dalam transaksi
	// Status selalu "on review" baik untuk review dosen maupun revisi taruna
	// Status hanya berubah ketika dosen melakukan approve/reject
	_, err = tx.Exec("UPDATE icp SET status = ? WHERE user_id = ? AND topik_penelitian = ?",
		"on review", tarunaID, topikPenelitian)
	if err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to update ICP status: " + err.Error(),
		})
		return
	}

	// Insert review ICP dalam transaksi
	dosenIDInt, _ := strconv.Atoi(dosenID)
	tarunaIDInt, _ := strconv.Atoi(tarunaID)

	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = tx.Exec(`
		INSERT INTO review_icp (
			dosen_id, taruna_id, topik_penelitian, 
			keterangan, file_path, status, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		dosenIDInt, tarunaIDInt, topikPenelitian,
		keterangan, filePath, "on review",
		now, now,
	)

	if err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving to database: " + err.Error(),
		})
		return
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Review ICP berhasil diunggah dan status diperbarui",
		"data": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// Handler untuk mengambil daftar review ICP dari table review_icp
func GetReviewProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	dosenID := r.URL.Query().Get("dosen_id")
	tarunaID := r.URL.Query().Get("taruna_id")

	if dosenID == "" && tarunaID == "" {
		http.Error(w, "Either dosen_id or taruna_id is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	reviewModel := models.NewReviewProposalModel(db)

	var reviews []entities.ReviewProposal
	if dosenID != "" {
		reviews, err = reviewModel.GetByDosenID(dosenID)
	} else {
		reviews, err = reviewModel.GetByTarunaID(tarunaID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   reviews,
	})
}

// Handler untuk upload review ICP oleh dosen ke table review_icp_dosen
func UploadDosenReviewProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error parsing form: " + err.Error(),
		})
		return
	}

	dosenID := r.FormValue("dosen_id")
	tarunaID := r.FormValue("taruna_id")
	userID := r.FormValue("user_id")
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	if dosenID == "" || tarunaID == "" || topikPenelitian == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Missing required fields",
		})
		return
	}

	db, err := config.GetDB()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer db.Close()

	// Verify taruna_id and user_id match
	var verifiedTarunaID int
	err = db.QueryRow("SELECT id FROM taruna WHERE id = ? AND user_id = ?", tarunaID, userID).Scan(&verifiedTarunaID)
	if err != nil {
		if err == sql.ErrNoRows {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "Invalid taruna_id and user_id combination",
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Get the ICP ID using user_id and topik_penelitian
	var icpID int
	err = db.QueryRow(`
		SELECT id 
		FROM icp 
		WHERE user_id = ? AND topik_penelitian = ?`,
		userID, topikPenelitian).Scan(&icpID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "ICP not found for the given user and topic: " + err.Error(),
		})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error retrieving file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	uploadDir := "uploads/reviewicp/dosen"
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating upload directory: " + err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("REVIEW_ICP_DOSEN_%s_%s_%s",
		dosenID,
		time.Now().Format("20060102150405"),
		handler.Filename)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating file: " + err.Error(),
		})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving file: " + err.Error(),
		})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to start transaction: " + err.Error(),
		})
		return
	}

	// Update status ICP dalam transaksi
	_, err = tx.Exec("UPDATE icp SET status = ? WHERE id = ?",
		"on review", icpID)
	if err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to update ICP status: " + err.Error(),
		})
		return
	}

	// Get current cycle number
	var cycleNumber int = 1
	err = tx.QueryRow(`
		SELECT COALESCE(MAX(cycle_number), 0) + 1 
		FROM review_icp_dosen 
		WHERE icp_id = ?`, icpID).Scan(&cycleNumber)
	if err != nil {
		// If error, default to 1
		cycleNumber = 1
	}

	// Insert review ICP dosen dalam transaksi
	dosenIDInt, _ := strconv.Atoi(dosenID)
	tarunaIDInt, _ := strconv.Atoi(tarunaID)

	_, err = tx.Exec(`
		INSERT INTO review_icp_dosen (
			icp_id, taruna_id, dosen_id, cycle_number,
			topik_penelitian, file_path, keterangan,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		icpID, tarunaIDInt, dosenIDInt, cycleNumber,
		topikPenelitian, filePath, keterangan,
	)

	if err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving to database: " + err.Error(),
		})
		return
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Review ICP dosen berhasil diunggah dan status diperbarui",
		"data": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// Handler untuk mengambil daftar review ICP dosen dari table review_icp_dosen
func GetReviewProposalDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	dosenID := r.URL.Query().Get("dosen_id")
	tarunaID := r.URL.Query().Get("taruna_id")

	if dosenID == "" && tarunaID == "" {
		http.Error(w, "Either dosen_id or taruna_id is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	reviewModel := models.NewReviewProposalDosenModel(db)

	var reviews []entities.ReviewProposal
	if dosenID != "" {
		reviews, err = reviewModel.GetByDosenID(dosenID)
	} else {
		reviews, err = reviewModel.GetByTarunaID(tarunaID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   reviews,
	})
}

// Handler untuk upload revisi ICP oleh taruna ke table review_icp_taruna
func UploadTarunaRevisiProposalHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error parsing form: " + err.Error(),
		})
		return
	}

	dosenID := r.FormValue("dosen_id")
	userID := r.FormValue("taruna_id") // This is actually the user_id from the frontend
	topikPenelitian := r.FormValue("topik_penelitian")
	keterangan := r.FormValue("keterangan")

	if dosenID == "" || userID == "" || topikPenelitian == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Missing required fields",
		})
		return
	}

	// Get ICP ID based on user_id and topik_penelitian
	db, err := config.GetDB()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer db.Close()

	// First get the taruna_id from taruna table
	var tarunaID int
	err = db.QueryRow("SELECT id FROM taruna WHERE user_id = ?", userID).Scan(&tarunaID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Taruna not found: " + err.Error(),
		})
		return
	}

	// Then get the ICP ID using the user_id
	var icpID int
	err = db.QueryRow("SELECT id FROM icp WHERE user_id = ? AND topik_penelitian = ?", userID, topikPenelitian).Scan(&icpID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "ICP not found for the given taruna and topic",
		})
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error retrieving file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	uploadDir := "uploads/reviewicp/taruna"
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating upload directory: " + err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("REVISI_ICP_TARUNA_%s_%s_%s",
		dosenID,
		time.Now().Format("20060102150405"),
		handler.Filename)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating file: " + err.Error(),
		})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving file: " + err.Error(),
		})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to start transaction: " + err.Error(),
		})
		return
	}

	// Update status ICP dalam transaksi
	_, err = tx.Exec("UPDATE icp SET status = ? WHERE id = ?",
		"on review", icpID)
	if err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to update ICP status: " + err.Error(),
		})
		return
	}

	// Get current cycle number
	var cycleNumber int = 1
	err = tx.QueryRow(`
		SELECT COALESCE(MAX(cycle_number), 0) + 1 
		FROM review_icp_taruna 
		WHERE icp_id = ?`, icpID).Scan(&cycleNumber)
	if err != nil {
		// If error, default to 1
		cycleNumber = 1
	}

	// Insert review ICP taruna dalam transaksi
	dosenIDInt, _ := strconv.Atoi(dosenID)

	_, err = tx.Exec(`
		INSERT INTO review_icp_taruna (
			icp_id, taruna_id, dosen_id, cycle_number,
			topik_penelitian, file_path, keterangan,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		icpID, tarunaID, dosenIDInt, cycleNumber,
		topikPenelitian, filePath, keterangan,
	)

	if err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving to database: " + err.Error(),
		})
		return
	}

	// Commit transaksi
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Failed to commit transaction: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Revisi ICP taruna berhasil diunggah dan status diperbarui",
		"data": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// Handler untuk mengambil daftar revisi ICP taruna dari table review_icp_taruna
func GetRevisiProposalTarunaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	userID := r.URL.Query().Get("taruna_id") // We receive user_id as taruna_id from frontend
	dosenID := r.URL.Query().Get("dosen_id")
	proposalID := r.URL.Query().Get("proposal_id")

	if userID == "" && dosenID == "" {
		http.Error(w, "Either taruna_id or dosen_id is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Base query with joins to get taruna and dosen names
	query := `
		SELECT 
			rit.id,
			rit.proposal_id,
			rit.taruna_id,
			rit.dosen_id,
			rit.topik_penelitian,
			rit.file_path,
			rit.keterangan,
			rit.cycle_number,
			rit.created_at,
			rit.updated_at,
			t.nama_lengkap as taruna_nama,
			t.user_id as user_id,
			d.nama_lengkap as dosen_nama
		FROM review_proposal rit
		LEFT JOIN taruna t ON rit.taruna_id = t.id
		LEFT JOIN dosen d ON rit.dosen_id = d.id
		WHERE 1=1
	`
	var args []interface{}

	if userID != "" {
		query += " AND t.user_id = ?"
		args = append(args, userID)
	}

	if dosenID != "" {
		query += " AND rit.dosen_id = ?"
		args = append(args, dosenID)
	}

	if proposalID != "" {
		query += " AND rit.proposal_id = ?"
		args = append(args, proposalID)
	}

	query += " ORDER BY rit.cycle_number DESC, rit.created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var revisions []map[string]interface{}
	for rows.Next() {
		var revision struct {
			ID              int
			ProposalID      int
			TarunaID        int
			DosenID         int
			TopikPenelitian string
			FilePath        string
			Keterangan      string
			CycleNumber     int
			CreatedAt       string
			UpdatedAt       string
			TarunaNama      sql.NullString
			UserID          sql.NullInt64
			DosenNama       sql.NullString
		}

		err := rows.Scan(
			&revision.ID,
			&revision.ProposalID,
			&revision.TarunaID,
			&revision.DosenID,
			&revision.TopikPenelitian,
			&revision.FilePath,
			&revision.Keterangan,
			&revision.CycleNumber,
			&revision.CreatedAt,
			&revision.UpdatedAt,
			&revision.TarunaNama,
			&revision.UserID,
			&revision.DosenNama,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		revisions = append(revisions, map[string]interface{}{
			"id":               revision.ID,
			"proposal_id":      revision.ProposalID,
			"taruna_id":        revision.TarunaID,
			"dosen_id":         revision.DosenID,
			"topik_penelitian": revision.TopikPenelitian,
			"file_path":        revision.FilePath,
			"keterangan":       revision.Keterangan,
			"cycle_number":     revision.CycleNumber,
			"created_at":       revision.CreatedAt,
			"updated_at":       revision.UpdatedAt,
			"taruna_nama":      revision.TarunaNama.String,
			"user_id":          revision.UserID.Int64,
			"dosen_nama":       revision.DosenNama.String,
		})
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   revisions,
	})
}

// Handler untuk mengambil detail review ICP dosen berdasarkan ID
func GetReviewProposalDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	// Get review ID from query parameter
	reviewID := r.URL.Query().Get("id")
	if reviewID == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT 
			r.id, r.proposal_id, r.taruna_id, r.dosen_id, r.cycle_number,
			r.topik_penelitian, r.file_path, r.keterangan, r.created_at,
			r.updated_at, t.nama_lengkap as nama_taruna, d.nama_lengkap as dosen_nama
		FROM review_proposal r
		LEFT JOIN taruna t ON r.taruna_id = t.id
		LEFT JOIN dosen d ON r.dosen_id = d.id
		WHERE r.id = ?
	`

	var review entities.ReviewProposal
	var namaTaruna, dosenNama sql.NullString
	err = db.QueryRow(query, reviewID).Scan(
		&review.ID,
		&review.ProposalID,
		&review.TarunaID,
		&review.DosenID,
		&review.CycleNumber,
		&review.TopikPenelitian,
		&review.FilePath,
		&review.Keterangan,
		&review.CreatedAt,
		&review.UpdatedAt,
		&namaTaruna,
		&dosenNama,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if namaTaruna.Valid {
		review.NamaTaruna = namaTaruna.String
	}
	if dosenNama.Valid {
		review.DosenNama = dosenNama.String
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   review,
	})
}

// Handler untuk mengambil detail review ICP dosen berdasarkan ID
func GetReviewProposalDosenDetailHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Content-Type", "application/json")

	// Get review ID from query parameter
	reviewID := r.URL.Query().Get("id")
	if reviewID == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT 
			r.id, r.proposal_id, r.taruna_id, r.dosen_id, r.cycle_number,
			r.topik_penelitian, r.file_path, r.keterangan, r.created_at,
			r.updated_at, t.nama_lengkap as nama_taruna, d.nama_lengkap as nama_dosen,
			p.status as proposal_status
		FROM review_proposal r
		LEFT JOIN taruna t ON r.taruna_id = t.id
		LEFT JOIN dosen d ON r.dosen_id = d.id
		LEFT JOIN proposal p ON r.proposal_id = p.id
		WHERE r.id = ?
	`

	var review struct {
		ID              int            `json:"id"`
		ProposalID      int            `json:"proposal_id"`
		TarunaID        int            `json:"taruna_id"`
		DosenID         int            `json:"dosen_id"`
		CycleNumber     int            `json:"cycle_number"`
		TopikPenelitian string         `json:"topik_penelitian"`
		FilePath        string         `json:"file_path"`
		Keterangan      string         `json:"keterangan"`
		CreatedAt       string         `json:"created_at"`
		UpdatedAt       string         `json:"updated_at"`
		NamaTaruna      sql.NullString `json:"nama_taruna"`
		NamaDosen       sql.NullString `json:"nama_dosen"`
		ProposalStatus  sql.NullString `json:"proposal_status"`
	}

	err = db.QueryRow(query, reviewID).Scan(
		&review.ID,
		&review.ProposalID,
		&review.TarunaID,
		&review.DosenID,
		&review.CycleNumber,
		&review.TopikPenelitian,
		&review.FilePath,
		&review.Keterangan,
		&review.CreatedAt,
		&review.UpdatedAt,
		&review.NamaTaruna,
		&review.NamaDosen,
		&review.ProposalStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert sql.NullString to string for JSON response
	response := map[string]interface{}{
		"id":               review.ID,
		"proposal_id":      review.ProposalID,
		"taruna_id":        review.TarunaID,
		"dosen_id":         review.DosenID,
		"cycle_number":     review.CycleNumber,
		"topik_penelitian": review.TopikPenelitian,
		"file_path":        review.FilePath,
		"keterangan":       review.Keterangan,
		"created_at":       review.CreatedAt,
		"updated_at":       review.UpdatedAt,
		"nama_taruna":      review.NamaTaruna.String,
		"nama_dosen":       review.NamaDosen.String,
		"proposal_status":  review.ProposalStatus.String,
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   response,
	})
}
