package models

type PartialExit struct {
	Hash                [32]byte `gorm:"primarykey"`
	BlockHash           [32]byte
	ValidatorPublicKey  [48]byte
	WithdrawalPublicKey [48]byte
	Signature           [96]byte
	Amount              uint64
}
