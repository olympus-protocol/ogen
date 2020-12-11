package models

type Tx struct {
	BlockHash         string
	Hash              string `gorm:"primarykey"`
	ToAddress         string
	FromPublicKeyHash string
	FromPublicKey     string
	Amount            uint64
	Nonce             uint64
	Fee               uint64
}
