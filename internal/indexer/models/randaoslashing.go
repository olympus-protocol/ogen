package models

type RandaoSlashing struct {
	Hash               [32]byte `gorm:"primarykey"`
	BlockHash          [32]byte
	RandaoReveal       [32]byte
	Slot               uint64
	ValidatorPublicKey [96]byte
}
