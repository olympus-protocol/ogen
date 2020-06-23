package primitives

import "github.com/olympus-protocol/ogen/utils/chainhash"

type VoteSlashing struct {
	Vote1 *MultiValidatorVote
	Vote2 *MultiValidatorVote
}

// Hash calculates the hash of the proposer slashing.
func (vs *VoteSlashing) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := vs.MarshalSSZ()
	return chainhash.HashH(b)
}

type RANDAOSlashing struct {
	RandaoReveal    []byte `ssz-size:"96"`
	Slot            uint64
	ValidatorPubkey []byte `ssz-size:"48"`
}

// Hash calculates the hash of the proposer slashing.
func (rs *RANDAOSlashing) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := rs.MarshalSSZ()
	return chainhash.HashH(b)
}

type ProposerSlashing struct {
	BlockHeader1       *BlockHeader
	BlockHeader2       *BlockHeader
	Signature1         []byte `ssz-size:"96"`
	Signature2         []byte `ssz-size:"96"`
	ValidatorPublicKey []byte `ssz-size:"48"`
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := ps.MarshalSSZ()
	return chainhash.HashH(b)
}
