package models

type RandaoSlashing struct {
	BlockHash          string
	RandaoReveal       string
	Slot               int
	ValidatorPublicKey string
}
