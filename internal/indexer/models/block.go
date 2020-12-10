package models

type Block struct {
	Hash            [32]byte    `gorm:"primarykey"`
	Header          BlockHeader `gorm:"foreignkey:Hash"`
	Signature       [96]byte
	RandaoSignature [96]byte
	Height          uint64
}
