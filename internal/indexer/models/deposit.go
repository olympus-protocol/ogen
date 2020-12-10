package models

type Deposit struct {
	Hash      string `gorm:"primarykey"`
	BlockHash string
	PublicKey string
	Signature string
	Data      DepositData
}

type DepositData struct {
	Hash              string `gorm:"primarykey"`
	PublicKey         string
	ProofOfPossession string
	WithdrawalAddress string
}
