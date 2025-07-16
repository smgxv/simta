package handlers

import (
	"database/sql"
	"document_service/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Struktur untuk response hasil telaah
type HasilTelaahResponse struct {
	ID              int       `json:"id"`
	NamaDosen       string    `json:"nama_dosen"`
	TopikPenelitian string    `json:"topik_penelitian"`
	FilePath        string    `json:"file_path"`
	TanggalTelaah   time.Time `json:"tanggal_telaah"`
	JumlahTelaah    int       `json:"jumlah_telaah"`
	StatusTelaah    string    `json:"status_telaah"`
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
	userID := r.FormValue("user_id")
	dosenID := r.FormValue("dosen_id")
	topikPenelitian := r.FormValue("topik_penelitian")

	// Debug log
	fmt.Printf("Received values - userID: %s, dosenID: %s, topik: %s\n", userID, dosenID, topikPenelitian)

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
	err = db.QueryRow("SELECT id FROM final_icp WHERE user_id = ? AND topik_penelitian = ?", userID, topikPenelitian).Scan(&icpID)
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
		userID,
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
	query := `INSERT INTO hasil_telaah_icp (icp_id, dosen_id, user_id, topik_penelitian, file_path, tanggal_telaah) 
			 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	result, err := db.Exec(query, icpID, dosenID, userID, topikPenelitian, filePath)
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
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080") // disamakan
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT 
			ht.id, 
			u.nama_lengkap AS nama_taruna,
			d.nama_lengkap AS nama_dosen, 
			ht.topik_penelitian, 
			ht.file_path, 
			ht.tanggal_telaah
		FROM hasil_telaah_icp ht
		JOIN users u ON u.id = ht.user_id
		JOIN dosen d ON d.id = ht.dosen_id
		ORDER BY ht.tanggal_telaah DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying database: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error querying database",
		})
		return
	}
	defer rows.Close()

	type HasilTelaahData struct {
		ID              int    `json:"id"`
		NamaTaruna      string `json:"nama_taruna"`
		NamaDosen       string `json:"nama_dosen"`
		TopikPenelitian string `json:"topik_penelitian"`
		FilePath        string `json:"file_path"`
		TanggalTelaah   string `json:"tanggal_telaah"`
	}

	var result []HasilTelaahData
	for rows.Next() {
		var h HasilTelaahData
		err := rows.Scan(
			&h.ID,
			&h.NamaTaruna,
			&h.NamaDosen,
			&h.TopikPenelitian,
			&h.FilePath,
			&h.TanggalTelaah,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		result = append(result, h)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Error reading data",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   result,
	})
}

func GetMonitoringTelaahHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Query utama untuk menghitung jumlah hasil telaah per ICP dengan informasi lengkap
	query := `
		SELECT 
			fi.id AS final_icp_id,
			u.nama_lengkap AS nama_taruna,
			fi.jurusan,
			fi.kelas,
			fi.topik_penelitian,

			d1.nama_lengkap AS penelaah_1,
			d2.nama_lengkap AS penelaah_2,

			CASE
				WHEN COUNT(CASE WHEN ht.id IS NOT NULL THEN 1 END) = 2 THEN 'Lengkap'
				ELSE 'Belum Lengkap'
			END AS status_kelengkapan

		FROM final_icp fi
		JOIN users u ON u.id = fi.user_id
		JOIN penelaah_icp pi ON pi.final_icp_id = fi.id
		LEFT JOIN dosen d1 ON d1.id = pi.penelaah_1_id
		LEFT JOIN dosen d2 ON d2.id = pi.penelaah_2_id
		LEFT JOIN hasil_telaah_icp ht ON ht.icp_id = fi.id

		GROUP BY fi.id, u.nama_lengkap, fi.jurusan, fi.kelas, fi.topik_penelitian, d1.nama_lengkap, d2.nama_lengkap
		ORDER BY u.nama_lengkap ASC;
	`

	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("[Error] Query error: %v\n", err)
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MonitoringResponse struct {
		FinalICPId      int    `json:"final_icp_id"`
		NamaTaruna      string `json:"nama_taruna"`
		Jurusan         string `json:"jurusan"`
		Kelas           string `json:"kelas"`
		TopikPenelitian string `json:"topik_penelitian"`
		JumlahTelaah    int    `json:"jumlah_telaah"`
		StatusTelaah    string `json:"status_telaah"`
		StatusICP       string `json:"status_icp"`
	}

	var results []MonitoringResponse

	for rows.Next() {
		var res MonitoringResponse
		var statusICP string
		err := rows.Scan(
			&res.FinalICPId,
			&res.NamaTaruna,
			&res.Jurusan,
			&res.Kelas,
			&res.TopikPenelitian,
			&res.JumlahTelaah,
			&statusICP,
		)
		if err != nil {
			fmt.Printf("[Error] Scan error: %v\n", err)
			continue
		}

		// Set status telaah berdasarkan jumlah telaah
		switch res.JumlahTelaah {
		case 2:
			res.StatusTelaah = "✅ Lengkap"
		case 1:
			res.StatusTelaah = "⚠️ Kurang 1 Telaah"
		default:
			res.StatusTelaah = "❌ Belum Ditelaah"
		}

		res.StatusICP = statusICP
		results = append(results, res)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   results,
	})
}

// Handler untuk mendapatkan detail telaah berdasarkan final_icp_id
func GetDetailTelaahICPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Ambil parameter ?id=
	icpIDStr := r.URL.Query().Get("id")
	icpID, err := strconv.Atoi(icpIDStr)
	if err != nil || icpID <= 0 {
		http.Error(w, "ID ICP tidak valid", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Ambil info taruna dan ICP
	var info struct {
		NamaTaruna      string `json:"nama_taruna"`
		Jurusan         string `json:"jurusan"`
		Kelas           string `json:"kelas"`
		TopikPenelitian string `json:"topik_penelitian"`
		StatusICP       string `json:"status_icp"`
	}
	infoQuery := `
		SELECT u.nama_lengkap, fi.jurusan, fi.kelas, fi.topik_penelitian, fi.status
		FROM final_icp fi
		JOIN users u ON u.id = fi.user_id
		WHERE fi.id = ?
	`
	err = db.QueryRow(infoQuery, icpID).Scan(&info.NamaTaruna, &info.Jurusan, &info.Kelas, &info.TopikPenelitian, &info.StatusICP)
	if err != nil {
		http.Error(w, "ICP tidak ditemukan", http.StatusNotFound)
		return
	}

	// Ambil info penelaah dan status telaahnya
	telaahQuery := `
		SELECT d.nama_lengkap, ht.tanggal_telaah, ht.file_path
		FROM penelaah_icp p
		JOIN dosen d ON d.id = p.penelaah_1_id
		LEFT JOIN hasil_telaah_icp ht ON ht.icp_id = p.final_icp_id AND ht.dosen_id = p.penelaah_1_id
		WHERE p.final_icp_id = ?

		UNION ALL

		SELECT d.nama_lengkap, ht.tanggal_telaah, ht.file_path
		FROM penelaah_icp p
		JOIN dosen d ON d.id = p.penelaah_2_id
		LEFT JOIN hasil_telaah_icp ht ON ht.icp_id = p.final_icp_id AND ht.dosen_id = p.penelaah_2_id
		WHERE p.final_icp_id = ?
	`

	rows, err := db.Query(telaahQuery, icpID, icpID)
	if err != nil {
		http.Error(w, "Gagal mengambil data telaah", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type TelaahItem struct {
		NamaDosen     string `json:"nama_dosen"`
		TanggalTelaah string `json:"tanggal_telaah"`
		FilePath      string `json:"file_path"`
	}
	var telaahList []TelaahItem
	telaahCount := 0

	for rows.Next() {
		var item TelaahItem
		var tanggal sql.NullString
		var file sql.NullString

		if err := rows.Scan(&item.NamaDosen, &tanggal, &file); err == nil {
			if tanggal.Valid {
				item.TanggalTelaah = tanggal.String
				telaahCount++
			}
			if file.Valid {
				item.FilePath = file.String
			}
			telaahList = append(telaahList, item)
		}
	}

	// Tentukan status telaah berdasarkan jumlah
	statusTelaah := "❌ Belum Ditelaah"
	if telaahCount == 2 {
		statusTelaah = "✅ Lengkap"
	} else if telaahCount == 1 {
		statusTelaah = "⚠️ Kurang 1 Telaah"
	}

	// Keluarkan JSON
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"info": map[string]interface{}{
			"nama_taruna":      info.NamaTaruna,
			"jurusan":          info.Jurusan,
			"kelas":            info.Kelas,
			"topik_penelitian": info.TopikPenelitian,
			"status_icp":       info.StatusICP,
			"status_telaah":    statusTelaah,
		},
		"telaah": telaahList,
	})
}

// Handler untuk mendapatkan daftar taruna yang ditelaah oleh dosen
func GetTarunaTopicsHandler(w http.ResponseWriter, r *http.Request) {
	// Tangani preflight CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	dosenID := r.URL.Query().Get("dosen_id")
	if dosenID == "" {
		http.Error(w, "Missing dosen_id parameter", http.StatusBadRequest)
		return
	}

	query := `
	SELECT 
		u.id AS user_id,
		u.nama_lengkap,
		fi.topik_penelitian
	FROM penelaah_icp p
	JOIN final_icp fi ON fi.id = p.final_icp_id
	JOIN users u ON u.id = fi.user_id
	WHERE p.penelaah_1_id = ? OR p.penelaah_2_id = ?
	ORDER BY u.nama_lengkap ASC;
	`

	rows, err := db.Query(query, dosenID, dosenID)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []struct {
		UserID          int    `json:"user_id"`
		NamaLengkap     string `json:"nama_lengkap"`
		TopikPenelitian string `json:"topik_penelitian"`
	}

	for rows.Next() {
		var row struct {
			UserID          int    `json:"user_id"`
			NamaLengkap     string `json:"nama_lengkap"`
			TopikPenelitian string `json:"topik_penelitian"`
		}
		if err := rows.Scan(&row.UserID, &row.NamaLengkap, &row.TopikPenelitian); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, row)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   results,
	})
}

// GetFinalICPByDosenHandler menangani request untuk mendapatkan data telaah ICP berdasarkan ID dosen
func GetFinalICPByDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	dosenID := r.URL.Query().Get("dosen_id")
	if dosenID == "" {
		http.Error(w, "Dosen ID is required", http.StatusBadRequest)
		return
	}

	dosenIDInt, err := strconv.Atoi(dosenID)
	if err != nil {
		http.Error(w, "Invalid dosen ID", http.StatusBadRequest)
		return
	}

	db, err := config.GetDB()
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT fl.id, fl.user_id, fl.topik_penelitian, fl.file_path, u.nama_lengkap
		FROM final_icp fl
		JOIN users u ON fl.user_id = u.id
		JOIN penelaah_icp pl ON fl.id = pl.final_icp_id
		WHERE pl.penelaah_1_id = ? OR pl.penelaah_2_id = ?
	`

	rows, err := db.Query(query, dosenIDInt, dosenIDInt)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type FinalICPData struct {
		ID              int    `json:"id"`
		UserID          int    `json:"user_id"`
		TopikPenelitian string `json:"topik_penelitian"`
		FilePath        string `json:"file_path"`
		TarunaNama      string `json:"taruna_nama"`
	}

	var finalicps []FinalICPData
	for rows.Next() {
		var p FinalICPData
		err := rows.Scan(&p.ID, &p.UserID, &p.TopikPenelitian, &p.FilePath, &p.TarunaNama)
		if err != nil {
			http.Error(w, "Error scanning rows: "+err.Error(), http.StatusInternalServerError)
			return
		}
		finalicps = append(finalicps, p)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   finalicps,
	})
}
