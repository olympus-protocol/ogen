package models

type PartialExit struct {
	Hash                string
	BlockHash           string
	ValidatorPublicKey  string
	WithdrawalPublicKey string
	Amount              uint64
}
