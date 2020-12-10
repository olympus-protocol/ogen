package models

type ProposerSlashing struct {
	Hash               [32]byte `gorm:"primarykey"`
	BlockHash          [32]byte
	BlockHeader1       BlockHeader `gorm:"foreignkey:Hash"`
	BlockHeader2       BlockHeader `gorm:"foreignkey:Hash"`
	Signature1         [96]byte
	Signature2         [96]byte
	ValidatorPublicKey [48]byte
}
