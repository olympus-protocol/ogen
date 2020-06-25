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
	hash, _ := ssz.HashTreeRoot(vs)
	return chainhash.Hash(hash)
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO
// signature too early.
type RANDAOSlashing struct {
	RandaoReveal    []byte
	Slot            uint64
	ValidatorPubkey []byte
}

func (rs *RANDAOSlashing) GetValidatorPubkey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(rs.ValidatorPubkey)
}

func (rs *RANDAOSlashing) GetRandaoReveal() (*bls.Signature, error) {
	return bls.SignatureFromBytes(rs.RandaoReveal)
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
	hash, _ := ssz.HashTreeRoot(rs)
	return chainhash.Hash(hash)
}

// ProposerSlashing is a slashing to a block proposer that proposed
// two blocks at the same slot.
type ProposerSlashing struct {
	BlockHeader1       BlockHeader
	BlockHeader2       BlockHeader
	Signature1         []byte
	Signature2         []byte
	ValidatorPublicKey []byte
}

func (ps *ProposerSlashing) GetValidatorPubkey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(ps.ValidatorPublicKey)
}

func (ps *ProposerSlashing) GetSignature1() (*bls.Signature, error) {
	return bls.SignatureFromBytes(ps.Signature1)
}

func (ps *ProposerSlashing) GetSignature2() (*bls.Signature, error) {
	return bls.SignatureFromBytes(ps.Signature2)
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
	hash, _ := ssz.HashTreeRoot(ps)
	return chainhash.Hash(hash)
}
