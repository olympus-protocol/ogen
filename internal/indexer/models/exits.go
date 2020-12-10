package models

type Exit struct {
	BlockHash           string
	ValidatorPublicKey  string
	WithdrawalPublicKey string
	Signature           string
}
