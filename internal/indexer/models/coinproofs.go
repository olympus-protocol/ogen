package models

type CoinProofs struct {
	Hash          [32]byte `gorm:"primarykey"`
	PkScript      [25]byte
	Transaction   [192]byte
	RedeemAccount [44]byte
}
