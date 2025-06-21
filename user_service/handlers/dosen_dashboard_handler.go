package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"user_service/config"
	"user_service/models"
)

type ICPDitelaah struct {
	NamaTaruna      string `json:"nama_taruna"`
	TopikPenelitian string `json:"topik_penelitian"`
	Status          string `json:"status"`
}
type DosenDashboardResponse struct {
	NamaLengkap string        `json:"nama_lengkap"`
	Jurusan     string        `json:"jurusan"`
	ICPs        []ICPDitelaah `json:"icp_ditelaah"`
}

type BimbinganResponse struct {
	NamaTaruna string `json:"nama_taruna"`
	Jurusan    string `json:"jurusan"`
}

func DosenDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// CORS setup
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ambil userId dari query param
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	dosenModel, err := models.NewDosenModel()
	if err != nil {
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ambil data dosen berdasarkan user_id
	dosen, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		http.Error(w, "Dosen not found", http.StatusNotFound)
		return
	}

	// 🔁 Ambil ICP berdasarkan dosen.ID (bukan userID)
	icpList, err := getICPListByDosen(dosen.ID)
	if err != nil {
		http.Error(w, "Gagal mengambil ICP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := DosenDashboardResponse{
		NamaLengkap: dosen.NamaLengkap,
		Jurusan:     dosen.Jurusan,
		ICPs:        icpList,
	}

	json.NewEncoder(w).Encode(resp)
}

func ICPDitelaahHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	// Ambil ID dosen berdasarkan user_id
	dosenModel, err := models.NewDosenModel()
	if err != nil {
		http.Error(w, "Gagal koneksi model", http.StatusInternalServerError)
		return
	}

	dosen, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		http.Error(w, "Dosen tidak ditemukan", http.StatusNotFound)
		return
	}

	// Gunakan dosen.ID, bukan userID
	icpList, err := getICPListByDosen(dosen.ID)
	if err != nil {
		http.Error(w, "Failed to fetch ICP list: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(icpList)
}

func getICPListByDosen(userID int) ([]ICPDitelaah, error) {
	db, err := config.ConnectDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT f.id, f.nama_lengkap, f.topik_penelitian
		FROM penelaah_icp p
		JOIN final_icp f ON f.id = p.final_icp_id
		WHERE p.penelaah_1_id = ? OR p.penelaah_2_id = ?
	`

	rows, err := db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ICPDitelaah
	for rows.Next() {
		var icpID int
		var nama, topik string
		if err := rows.Scan(&icpID, &nama, &topik); err != nil {
			return nil, err
		}

		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM hasil_telaah_icp
			WHERE dosen_id = ? AND icp_id = ?
		`, userID, icpID).Scan(&count)
		if err != nil {
			return nil, err
		}

		status := "Belum Ditelaah"
		if count > 0 {
			status = "Sudah Ditelaah"
		}

		results = append(results, ICPDitelaah{
			NamaTaruna:      nama,
			TopikPenelitian: topik,
			Status:          status,
		})
	}

	return results, nil
}

func GetBimbinganByDosenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	dosenModel, err := models.NewDosenModel()
	if err != nil {
		http.Error(w, "Model error", http.StatusInternalServerError)
		return
	}

	dosen, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		http.Error(w, "Dosen tidak ditemukan", http.StatusNotFound)
		return
	}

	db, err := config.ConnectDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT t.nama_lengkap, t.jurusan
		FROM dosbing_proposal d
		JOIN users u ON d.user_id = u.id
		JOIN taruna t ON u.id = t.user_id
		WHERE d.dosen_id = ?
	`

	rows, err := db.Query(query, dosen.ID)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []BimbinganResponse
	for rows.Next() {
		var nama, jurusan string
		err := rows.Scan(&nama, &jurusan)
		if err != nil {
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}
		results = append(results, BimbinganResponse{
			NamaTaruna: nama,
			Jurusan:    jurusan,
		})
	}

	json.NewEncoder(w).Encode(results)
}

func GetPengujianProposalHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://104.43.89.154:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	dosenModel, err := models.NewDosenModel()
	if err != nil {
		http.Error(w, "Model error", http.StatusInternalServerError)
		return
	}

	dosen, err := dosenModel.GetDosenByUserID(userID)
	if err != nil {
		http.Error(w, "Dosen tidak ditemukan", http.StatusNotFound)
		return
	}

	db, err := config.ConnectDB()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `
		SELECT 
			fp.nama_lengkap,
			fp.topik_penelitian,
			CASE 
				WHEN pp.ketua_penguji_id = ? THEN 'Ketua Penguji'
				WHEN pp.penguji_1_id = ? THEN 'Penguji 1'
				WHEN pp.penguji_2_id = ? THEN 'Penguji 2'
				ELSE ''
			END AS penguji_ke,
			COALESCE(spp.status_pengumpulan, 'belum') AS status_pengumpulan
		FROM penguji_proposal pp
		JOIN final_proposal fp ON pp.final_proposal_id = fp.id
		LEFT JOIN seminar_proposal_penilaian spp 
			ON spp.final_proposal_id = pp.final_proposal_id 
			AND spp.dosen_id = ?
		WHERE 
			pp.ketua_penguji_id = ? OR 
			pp.penguji_1_id = ? OR 
			pp.penguji_2_id = ?
	`

	rows, err := db.Query(query, dosen.ID, dosen.ID, dosen.ID, dosen.ID, dosen.ID, dosen.ID)
	if err != nil {
		http.Error(w, "Query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type PengujianResponse struct {
		NamaTaruna        string `json:"nama_taruna"`
		Topik             string `json:"topik"`
		PengujiKe         string `json:"penguji_ke"`
		StatusPengumpulan string `json:"status_pengumpulan"`
	}

	var results []PengujianResponse
	for rows.Next() {
		var nama, topik, pengujiKe, status string
		err := rows.Scan(&nama, &topik, &pengujiKe, &status)
		if err != nil {
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}

		results = append(results, PengujianResponse{
			NamaTaruna:        nama,
			Topik:             topik,
			PengujiKe:         pengujiKe,
			StatusPengumpulan: status,
		})
	}

	json.NewEncoder(w).Encode(results)
}
