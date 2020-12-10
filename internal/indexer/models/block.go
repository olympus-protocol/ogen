package models

type Block struct {
	Hash            string `gorm:"primarykey"`
	Header          BlockHeader
	Signature       string
	RandaoSignature string
	Height          int
}
