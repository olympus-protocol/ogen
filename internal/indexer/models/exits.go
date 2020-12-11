package models

type Exit struct {
	Hash                string `gorm:"primarykey"`
	BlockHash           string
	ValidatorPublicKey  string
	WithdrawalPublicKey string
}
