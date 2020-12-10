package models

type Tx struct {
	BlockHash         string
	Hash              string `gorm:"primarykey"`
	TxType            int
	ToAddress         string
	FromPublicKey     string
	FromPublicKeyHash string
	Amount            int
	Nonce             int
	Fee               int
	Signature         string
}
