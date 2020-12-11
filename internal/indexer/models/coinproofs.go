package models

type CoinProofs struct {
	Hash          string `gorm:"primarykey"`
	Transaction   string
	RedeemAccount string
}
