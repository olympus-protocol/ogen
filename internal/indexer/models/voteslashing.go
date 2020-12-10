package models

type VoteSlashing struct {
	Hash      [32]byte `gorm:"primarykey"`
	BlockHash [32]byte
	Vote1     Vote `gorm:"foreignkey:Hash"`
	Vote2     Vote `gorm:"foreignkey:Hash"`
}
