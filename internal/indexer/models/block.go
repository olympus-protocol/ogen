package models

type Block struct {
	Hash     string `gorm:"primaryKey"`
	Height   uint64
	Slot     uint64
	Header   BlockHeader `gorm:"foreignKey:Hash"`
	Txs      []Tx
	Deposits []Deposit
	Votes    []Vote
}
