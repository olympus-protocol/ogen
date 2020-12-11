package models

type Deposit struct {
	Hash      string `gorm:"primaryKey"`
	BlockHash string
	PublicKey string
	Data      DepositData `gorm:"foreignKey:Hash"`
}

type DepositData struct {
	Hash              string
	PublicKey         string `gorm:"primaryKey"`
	ProofOfPossession string
	WithdrawalAddress string
}
