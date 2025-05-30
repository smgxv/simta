// entities/dosbing_proposal.go
package entities

type DosbingProposal struct {
	UserID  int    `json:"user_id"`
	DosenID int    `json:"dosen_id"`
	Status  string `json:"status,omitempty"` // optional input
}
