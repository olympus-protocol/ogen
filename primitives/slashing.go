package primitives

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// VoteSlashing is a slashing where validators vote in the span of their
// other votes.
type VoteSlashing struct {
	Vote1 MultiValidatorVote
	Vote2 MultiValidatorVote

	ssz.Marshaler
	ssz.Unmarshaler
}

// Hash calculates the hash of the slashing.
func (vs *VoteSlashing) Hash() (chainhash.Hash, error) {
	ser, err := vs.MarshalSSZ()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO
// signature too early.
type RANDAOSlashing struct {
	RandaoReveal    bls.Signature
	Slot            uint64
	ValidatorPubkey bls.PublicKey

	ssz.Marshaler
	ssz.Unmarshaler
}

// Hash calculates the hash of the RANDAO slashing.
func (rs *RANDAOSlashing) Hash() (chainhash.Hash, error) {
	ser, err := rs.MarshalSSZ()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}

// ProposerSlashing is a slashing to a block proposer that proposed
// two blocks at the same slot.
type ProposerSlashing struct {
	BlockHeader1       BlockHeader
	BlockHeader2       BlockHeader
	Signature1         bls.Signature
	Signature2         bls.Signature
	ValidatorPublicKey bls.PublicKey

	ssz.Marshaler
	ssz.Unmarshaler
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() (chainhash.Hash, error) {
	ser, err := ps.MarshalSSZ()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}
