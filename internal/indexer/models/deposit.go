package models

type Deposit struct {
	Hash      [32]byte `gorm:"primarykey"`
	BlockHash [32]byte
	PublicKey [48]byte
	Signature [96]byte
	Data      DepositData `gorm:"foreignkey:Hash"`
}

type DepositData struct {
	Hash              [32]byte `gorm:"primarykey"`
	PublicKey         [48]byte
	ProofOfPossession [96]byte
	WithdrawalAddress [20]byte
}
