package entities

type SeminarProposal struct {
	ID              int    `json:"id"`
	UserID          int    `json:"user_id"`
	KetuaPengujiID  int    `json:"ketua_penguji_id"`
	Penguji1ID      int    `json:"penguji1_id"`
	Penguji2ID      int    `json:"penguji2_id"`
	TopikPenelitian string `json:"topik_penelitian"`
	FilePath        string `json:"file_path"`
	Status          string `json:"status"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}
