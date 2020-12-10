package models

type RandaoSlashing struct {
	Hash               string `gorm:"primarykey"`
	BlockHash          string
	RandaoReveal       string
	Slot               int
	ValidatorPublicKey string
}
