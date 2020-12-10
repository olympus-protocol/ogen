package models

type Tx struct {
	BlockHash         [32]byte
	Hash              [32]byte `gorm:"primarykey"`
	TxType            uint64
	ToAddress         [20]byte
	FromPublicKey     [48]byte
	FromPublicKeyHash [20]byte
	Amount            uint64
	Nonce             uint64
	Fee               uint64
	Signature         [96]byte
}
