package models

type Deposit struct {
	BlockHash string
	PublicKey string
	Signature string
	Data      DepositData
}

type DepositData struct {
	PublicKey         string
	ProofOfPossession string
	WithdrawalAddress string
}
