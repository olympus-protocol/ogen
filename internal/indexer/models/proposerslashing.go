package models

type ProposerSlashing struct {
	BlockHash          string
	BlockHeader1       BlockHeader
	BlockHeader2       BlockHeader
	Signature1         string
	Signature2         string
	ValidatorPublicKey string
}
