package models

type Slot struct {
	Slot          uint64 `gorm:"primaryKey"`
	BlockHash     string
	ProposerIndex uint64
	Proposed      bool
}
