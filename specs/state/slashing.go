package state

type VoteSlashing struct {
	Vote1 *MultiValidatorVote
	Vote2 *MultiValidatorVote
}

type RANDAOSlashing struct {
	RandaoReveal    []byte `ssz-size:"96"`
	Slot            uint64
	ValidatorPubkey []byte `ssz-size:"48"`
}

type ProposerSlashing struct {
	BlockHeader1       *BlockHeader
	BlockHeader2       *BlockHeader
	Signature1         []byte `ssz-size:"96"`
	Signature2         []byte `ssz-size:"96"`
	ValidatorPublicKey []byte `ssz-size:"48"`
}
