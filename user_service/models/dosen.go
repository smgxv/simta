package models

import (
	"user_service/database"
)

type Dosen struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Nama    string `json:"nama"`
	Jurusan string `json:"jurusan"`
}

func GetDosenByUserID(userID string) (Dosen, error) {
	var dosen Dosen
	query := `SELECT id, user_id, nama, jurusan FROM dosen WHERE user_id = $1`
	err := database.DB.QueryRow(query, userID).Scan(&dosen.ID, &dosen.UserID, &dosen.Nama, &dosen.Jurusan)
	return dosen, err
}
