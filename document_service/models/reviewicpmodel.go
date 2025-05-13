package models

import (
	"database/sql"
	"document_service/entities"
	"time"
)

type ReviewICPModel struct {
	db *sql.DB
}

func NewReviewICPModel(db *sql.DB) *ReviewICPModel {
	return &ReviewICPModel{db: db}
}

func (m *ReviewICPModel) Create(review *entities.ReviewICP) error {
	query := `
		INSERT INTO review_icp (
			dosen_id, taruna_id, topik_penelitian, keterangan, 
			file_path, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := m.db.Exec(
		query,
		review.DosenID,
		review.TarunaID,
		review.TopikPenelitian,
		review.Keterangan,
		review.FilePath,
		"pending",
		now,
		now,
	)
	return err
}
