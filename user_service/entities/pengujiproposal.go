// entities/pengujiproposal.go
package entities

type PengujiProposal struct {
	TarunaID int    `json:"taruna_id"` // Ganti jadi taruna_id
	DosenID  int    `json:"dosen_id"`
	Status   string `json:"status,omitempty"`
}
