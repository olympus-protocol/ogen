package models

type Slot struct {
	Slot          int
	BlockHash     string
	ProposerIndex int
	Proposed      bool
}
