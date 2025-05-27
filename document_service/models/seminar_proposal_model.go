package models

import (
	"database/sql"
	"document_service/entities"
)

type SeminarProposalModel struct {
	db *sql.DB
}

func NewSeminarProposalModel(db *sql.DB) *SeminarProposalModel {
	return &SeminarProposalModel{
		db: db,
	}
}

func (m *SeminarProposalModel) Create(seminarProposal *entities.SeminarProposal) error {
	query := `
		INSERT INTO seminar_proposal (
			user_id, ketua_penguji_id, penguji1_id, penguji2_id,
			topik_penelitian, file_path, status
		) VALUES (?, ?, ?, ?, ?, ?, 'pending')
	`

	result, err := m.db.Exec(
		query,
		seminarProposal.UserID,
		seminarProposal.KetuaPengujiID,
		seminarProposal.Penguji1ID,
		seminarProposal.Penguji2ID,
		seminarProposal.TopikPenelitian,
		seminarProposal.FilePath,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	seminarProposal.ID = int(id)
	return nil
}

func (m *SeminarProposalModel) GetByUserID(userID string) ([]entities.SeminarProposal, error) {
	query := `
		SELECT 
			sp.id, sp.user_id, sp.ketua_penguji_id, sp.penguji1_id, sp.penguji2_id,
			sp.topik_penelitian, sp.file_path, sp.status, sp.created_at, sp.updated_at
		FROM seminar_proposal sp
		WHERE sp.user_id = ?
	`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seminarProposals []entities.SeminarProposal
	for rows.Next() {
		var sp entities.SeminarProposal
		err := rows.Scan(
			&sp.ID, &sp.UserID, &sp.KetuaPengujiID, &sp.Penguji1ID, &sp.Penguji2ID,
			&sp.TopikPenelitian, &sp.FilePath, &sp.Status, &sp.CreatedAt, &sp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		seminarProposals = append(seminarProposals, sp)
	}

	return seminarProposals, nil
}
