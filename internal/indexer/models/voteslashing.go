package models

type VoteSlashing struct {
	BlockHash string
	Vote1     Vote
	Vote2     Vote
}
