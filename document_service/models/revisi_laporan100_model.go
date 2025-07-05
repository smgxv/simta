package models

import (
	"database/sql"
	"document_service/entities"
)

type RevisiLaporan100Model struct {
	db *sql.DB
}

func NewRevisiLaporan100Model(db *sql.DB) *RevisiLaporan100Model {
	return &RevisiLaporan100Model{
		db: db,
	}
}

func (m *RevisiLaporan100Model) Create(revisiLaporan100 *entities.RevisiLaporan100) error {
	query := `
		INSERT INTO revisi_laporan100 (
			user_id, nama_lengkap, jurusan, 
			kelas, topik_penelitian, file_path, keterangan, 
			status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.db.Exec(query,
		revisiLaporan100.UserID,
		revisiLaporan100.NamaLengkap,
		revisiLaporan100.Jurusan,
		revisiLaporan100.Kelas,
		revisiLaporan100.TopikPenelitian,
		revisiLaporan100.FilePath,
		revisiLaporan100.Keterangan,
		"pending", // default status
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	revisiLaporan100.ID = int(id)
	return nil
}

func (m *RevisiLaporan100Model) GetByUserID(userID string) ([]entities.RevisiLaporan100, error) {
	query := `
		SELECT 
			id, user_id, nama_lengkap, 
			jurusan, kelas, topik_penelitian, file_path, 
			keterangan, status, created_at, updated_at
		FROM revisi_laporan100 
		WHERE user_id = ?
		ORDER BY created_at DESC`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revisiLaporan100s []entities.RevisiLaporan100
	for rows.Next() {
		var revisiLaporan100 entities.RevisiLaporan100
		err := rows.Scan(
			&revisiLaporan100.ID,
			&revisiLaporan100.UserID,
			&revisiLaporan100.NamaLengkap,
			&revisiLaporan100.Jurusan,
			&revisiLaporan100.Kelas,
			&revisiLaporan100.TopikPenelitian,
			&revisiLaporan100.FilePath,
			&revisiLaporan100.Keterangan,
			&revisiLaporan100.Status,
			&revisiLaporan100.CreatedAt,
			&revisiLaporan100.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		revisiLaporan100s = append(revisiLaporan100s, revisiLaporan100)
	}

	return revisiLaporan100s, nil
}
