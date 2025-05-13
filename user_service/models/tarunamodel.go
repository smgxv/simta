package models

import (
	"database/sql"
	"user_service/config"
)

type TarunaModel struct {
	db *sql.DB
}

func NewTarunaModel() (*TarunaModel, error) {
	db, err := config.ConnectDB()
	if err != nil {
		return nil, err
	}
	return &TarunaModel{db: db}, nil
}

func (d *TarunaModel) CreateTaruna(userID int, namaLengkap, jurusan string) error {
	_, err := d.db.Exec("INSERT INTO dosen (id, nama_lengkap, jurusan) VALUES (?, ?, ?)",
		userID, namaLengkap, jurusan)
	return err
}

func (d *TarunaModel) GetAllTaruna() ([]map[string]interface{}, error) {
	rows, err := d.db.Query(`
        SELECT t.id, t.user_id, t.nama_lengkap, t.jurusan 
        FROM taruna t
        JOIN users u ON t.user_id = u.id
        WHERE u.role = 'Taruna'
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dosens []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var namaLengkap, jurusan string

		err := rows.Scan(&id, &userID, &namaLengkap, &jurusan)
		if err != nil {
			return nil, err
		}

		dosen := map[string]interface{}{
			"id":           id,
			"user_id":      userID,
			"nama_lengkap": namaLengkap,
			"jurusan":      jurusan,
		}
		dosens = append(dosens, dosen)
	}

	return dosens, nil
}
