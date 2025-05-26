package handlers

import (
	"document_service/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Struktur untuk response hasil telaah
type HasilTelaahResponse struct {
	ID              int       `json:"id"`
	NamaDosen       string    `json:"nama_dosen"`
	TopikPenelitian string    `json:"topik_penelitian"`
	FilePath        string    `json:"file_path"`
	TanggalTelaah   time.Time `json:"tanggal_telaah"`
}

func UploadHasilTelaahHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error parsing form: " + err.Error(),
		})
		return
	}

	// Get form values
	tarunaID := r.FormValue("taruna_id")
	dosenID := r.FormValue("dosen_id")
	topikPenelitian := r.FormValue("topik_penelitian")

	// Debug log
	fmt.Printf("Received values - tarunaID: %s, dosenID: %s, topik: %s\n", tarunaID, dosenID, topikPenelitian)

	// Connect to database for getting icp_id
	db, err := config.GetDB()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer db.Close()

	// Get icp_id from final_icp table
	var icpID int
	err = db.QueryRow("SELECT id FROM final_icp WHERE user_id = ? AND topik_penelitian = ?", tarunaID, topikPenelitian).Scan(&icpID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error getting ICP ID: " + err.Error(),
		})
		return
	}

	// Handle file upload
	file, handler, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error retrieving file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Create upload directory if not exists
	uploadDir := "uploads/hasil_telaah_icp"
	if err := os.MkdirAll(uploadDir, 0777); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating upload directory: " + err.Error(),
		})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("HASIL_TELAAH_ICP_%s_%s_%s_%s",
		tarunaID,
		dosenID,
		time.Now().Format("20060102150405"),
		handler.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error creating file: " + err.Error(),
		})
		return
	}
	defer dst.Close()

	// Copy the uploaded file
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving file: " + err.Error(),
		})
		return
	}

	// Insert into database
	query := `INSERT INTO hasil_telaah_icp (icp_id, dosen_id, taruna_id, topik_penelitian, file_path, tanggal_telaah) 
			 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	result, err := db.Exec(query, icpID, dosenID, tarunaID, topikPenelitian, filePath)
	if err != nil {
		os.Remove(filePath)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error saving to database: " + err.Error(),
		})
		return
	}

	// Get the inserted ID
	id, _ := result.LastInsertId()

	// Return success response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Hasil telaah berhasil diunggah",
		"data": map[string]interface{}{
			"id":        id,
			"file_path": filePath,
		},
	})
}

func GetHasilTelaahTarunaHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Get user_id from query parameter
	userID := r.URL.Query().Get("user_id")
	fmt.Printf("[Debug] Received request for user_id: %s\n", userID)

	if userID == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "User ID is required",
		})
		return
	}

	// Get database connection
	db, err := config.GetDB()
	if err != nil {
		fmt.Printf("[Error] Database connection error: %v\n", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer db.Close()

	// Pertama, dapatkan taruna_id dari user_id
	var tarunaID int
	err = db.QueryRow("SELECT id FROM taruna WHERE user_id = ?", userID).Scan(&tarunaID)
	if err != nil {
		fmt.Printf("[Error] Failed to get taruna_id: %v\n", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error getting taruna data: " + err.Error(),
		})
		return
	}

	fmt.Printf("[Debug] Found taruna_id: %d for user_id: %s\n", tarunaID, userID)

	// Query untuk mengambil data hasil telaah menggunakan taruna_id
	query := `
		SELECT ht.id, d.nama_lengkap, ht.topik_penelitian, ht.file_path, ht.tanggal_telaah
		FROM hasil_telaah_icp ht
		JOIN dosen d ON ht.dosen_id = d.id
		WHERE ht.taruna_id = ?
		ORDER BY ht.tanggal_telaah DESC`

	fmt.Printf("[Debug] Executing query with taruna_id: %d\n", tarunaID)
	rows, err := db.Query(query, tarunaID)
	if err != nil {
		fmt.Printf("[Error] Query execution error: %v\n", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Query error: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var results []HasilTelaahResponse
	for rows.Next() {
		var result HasilTelaahResponse
		err := rows.Scan(
			&result.ID,
			&result.NamaDosen,
			&result.TopikPenelitian,
			&result.FilePath,
			&result.TanggalTelaah,
		)
		if err != nil {
			fmt.Printf("[Error] Row scan error: %v\n", err)
			continue
		}
		results = append(results, result)
	}

	fmt.Printf("[Debug] Found %d hasil telaah records\n", len(results))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   results,
	})
}
