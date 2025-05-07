package models

import (
	"database/sql"
	"user_service/config"
)

type DosenModel struct {
	db *sql.DB
}

func NewDosenModel() (*DosenModel, error) {
	db, err := config.ConnectDB()
	if err != nil {
		return nil, err
	}
	return &DosenModel{db: db}, nil
}

func (d *DosenModel) CreateDosen(userID int, namaLengkap, jurusan string) error {
	_, err := d.db.Exec("INSERT INTO dosen (id, nama_lengkap, jurusan) VALUES (?, ?, ?)",
		userID, namaLengkap, jurusan)
	return err
}

func (d *DosenModel) GetAllDosen() ([]map[string]interface{}, error) {
	rows, err := d.db.Query(`
        SELECT d.id, d.nama_lengkap, d.jurusan 
        FROM dosen d
        JOIN users u ON d.user_id = u.id
        WHERE u.role = 'Dosen'
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dosens []map[string]interface{}
	for rows.Next() {
		var id int
		var namaLengkap, jurusan string

		err := rows.Scan(&id, &namaLengkap, &jurusan)
		if err != nil {
			return nil, err
		}

		dosen := map[string]interface{}{
			"id":           id,
			"nama_lengkap": namaLengkap,
			"jurusan":      jurusan,
		}
		dosens = append(dosens, dosen)
	}

	return dosens, nil
}
