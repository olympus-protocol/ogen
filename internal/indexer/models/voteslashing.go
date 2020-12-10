package models

type VoteSlashing struct {
	Hash      string `gorm:"primarykey"`
	BlockHash string
	Vote1     Vote
	Vote2     Vote
}
