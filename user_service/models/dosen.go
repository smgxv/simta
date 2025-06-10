package models

import (
	"user_service/database"
)

type Dosen struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Nama    string `json:"nama_lengkap"`
	Jurusan string `json:"jurusan"`
}

func GetDosenByUserID(userID string) (Dosen, error) {
	var dosen Dosen
	query := `SELECT id, user_id, nama_lengkap, jurusan FROM dosen WHERE user_id = ?`
	err := database.DB.QueryRow(query, userID).Scan(&dosen.ID, &dosen.UserID, &dosen.Nama, &dosen.Jurusan)
	return dosen, err
}
