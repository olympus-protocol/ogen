package models

type Slot struct {
	Slot          int `gorm:"primarykey"`
	BlockHash     string
	ProposerIndex int
	Proposed      bool
}
