package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// VoteSlashing is a slashing where validators vote in the span of their
// other votes.
type VoteSlashing struct {
	Vote1 MultiValidatorVote
	Vote2 MultiValidatorVote
}

// Marshal serializes the struct to bytes
func (vs *VoteSlashing) Marshal() ([]byte, error) {
	return ssz.Marshal(vs)
}

// Unmarshal deserializes the struct from bytes
func (vs *VoteSlashing) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, vs)
}

// Hash calculates the hash of the slashing.
func (vs *VoteSlashing) Hash() (chainhash.Hash, error) {
	ser, err := vs.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO
// signature too early.
type RANDAOSlashing struct {
	RandaoReveal    []byte `ssz:"size=96"`
	Slot            uint64
	ValidatorPubkey []byte `ssz:"size=48"`
}

// Marshal serializes the struct to bytes
func (rs *RANDAOSlashing) Marshal() ([]byte, error) {
	return ssz.Marshal(rs)
}

// Unmarshal deserializes the struct from bytes
func (rs *RANDAOSlashing) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, rs)
}

// Hash calculates the hash of the RANDAO slashing.
func (rs *RANDAOSlashing) Hash() (chainhash.Hash, error) {
	ser, err := rs.Marshal()
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
	Signature1         []byte `ssz:"size=96"`
	Signature2         []byte `ssz:"size=96"`
	ValidatorPublicKey []byte `ssz:"size=48"`
}

// Marshal serializes the struct to bytes
func (ps *ProposerSlashing) Marshal() ([]byte, error) {
	return ssz.Marshal(ps)
}

// Unmarshal deserializes the struct from bytes
func (ps *ProposerSlashing) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, ps)
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() (chainhash.Hash, error) {
	ser, err := ps.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(ser), nil
}
