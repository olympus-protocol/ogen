package primitives

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// VoteSlashing is a slashing where validators vote in the span of their
// other votes.
type VoteSlashing struct {
	Vote1 MultiValidatorVote
	Vote2 MultiValidatorVote
}

// Marshal encodes the data.
func (vs *VoteSlashing) Marshal() ([]byte, error) {
	return ssz.Marshal(vs)
}

// Unmarshal decodes the data.
func (vs *VoteSlashing) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, vs)
}

// Hash calculates the hash of the slashing.
func (vs *VoteSlashing) Hash() chainhash.Hash {
	b, _ := vs.Marshal()
	return chainhash.HashH(b)
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO
// signature too early.
type RANDAOSlashing struct {
	RandaoReveal    bls.Signature
	Slot            uint64
	ValidatorPubkey bls.PublicKey
}

// Marshal encodes the data.
func (rs *RANDAOSlashing) Marshal() ([]byte, error) {
	return ssz.Marshal(rs)
}

// Unmarshal decodes the data.
func (rs *RANDAOSlashing) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, rs)
}

// Hash calculates the hash of the RANDAO slashing.
func (rs *RANDAOSlashing) Hash() chainhash.Hash {
	b, _ := rs.Marshal()
	return chainhash.HashH(b)
}

// ProposerSlashing is a slashing to a block proposer that proposed
// two blocks at the same slot.
type ProposerSlashing struct {
	BlockHeader1       BlockHeader
	BlockHeader2       BlockHeader
	Signature1         bls.Signature
	Signature2         bls.Signature
	ValidatorPublicKey bls.PublicKey
}

// Marshal encodes the data.
func (ps *ProposerSlashing) Marshal() ([]byte, error) {
	return ssz.Marshal(ps)
}

// Unmarshal decodes the data.
func (ps *ProposerSlashing) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, ps)
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() chainhash.Hash {
	b, _ := ps.Marshal()
	return chainhash.HashH(b)
}
