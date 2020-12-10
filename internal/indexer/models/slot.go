package models

type Slot struct {
	Slot          uint64 `gorm:"primarykey"`
	BlockHash     [96]byte
	ProposerIndex uint64
	Proposed      bool
}
