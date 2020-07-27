package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

var (
	// ErrorRandaoSlashingSize returns when the randao slashing is above MaxRandaoSlashingSize
	ErrorRandaoSlashingSize = errors.New("randao slashing too big")
	// ErrorProposerSlashingSize returns when the randao slashing is above MaxRandaoSlashingSize
	ErrorProposerSlashingSize = errors.New("proposer slashing too big")
	// ErrorVoteSlashingSize returns when the vote slashing is above MaxVoteSlashingSize
	ErrorVoteSlashingSize = errors.New("proposer slashing too big")
)

const (
	// MaxRandaoSlashingSize is the maximum amount of bytes a randao slashing can contain.
	MaxRandaoSlashingSize = 160
	// MaxProposerSlashingSize is the maximum amount of bytes a proposer slashing can contain.
	MaxProposerSlashingSize = MaxBlockHeaderBytes*2 + 96*2 + 48
	// MaxVoteSlashingSize is the maximum amount of bytes a vote slashing can contain.
	MaxVoteSlashingSize = MaxMultiValidatorVoteSize * 2
)

// VoteSlashing is a slashing where validators vote in the span of their other votes.
type VoteSlashing struct {
	Vote1 MultiValidatorVote
	Vote2 MultiValidatorVote
}

// Marshal encodes the data.
func (vs *VoteSlashing) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(vs)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxVoteSlashingSize {
		return nil, ErrorVoteSlashingSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (vs *VoteSlashing) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxVoteSlashingSize {
		return ErrorVoteSlashingSize
	}
	return ssz.Unmarshal(d, vs)
}

// Hash calculates the hash of the slashing.
func (vs *VoteSlashing) Hash() chainhash.Hash {
	b, _ := vs.Marshal()
	return chainhash.HashH(b)
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO signature too early.
type RANDAOSlashing struct {
	RandaoReveal    []byte
	Slot            uint64
	ValidatorPubkey []byte
}

// GetValidatorPubkey returns the validator bls public key.
func (rs *RANDAOSlashing) GetValidatorPubkey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(rs.ValidatorPubkey)
}

// GetRandaoReveal returns the bls signature of the randao reveal.
func (rs *RANDAOSlashing) GetRandaoReveal() (*bls.Signature, error) {
	return bls.SignatureFromBytes(rs.RandaoReveal)
}

// Marshal encodes the data.
func (rs *RANDAOSlashing) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(rs)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxRandaoSlashingSize {
		return nil, ErrorRandaoSlashingSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (rs *RANDAOSlashing) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxRandaoSlashingSize {
		return ErrorRandaoSlashingSize
	}
	return ssz.Unmarshal(d, rs)
}

// Hash calculates the hash of the RANDAO slashing.
func (rs *RANDAOSlashing) Hash() chainhash.Hash {
	b, _ := rs.Marshal()
	return chainhash.HashH(b)
}

// ProposerSlashing is a slashing to a block proposer that proposed two blocks at the same slot.
type ProposerSlashing struct {
	BlockHeader1       BlockHeader
	BlockHeader2       BlockHeader
	Signature1         []byte
	Signature2         []byte
	ValidatorPublicKey []byte
}

// GetValidatorPubkey returns the slashing bls validator public key.
func (ps *ProposerSlashing) GetValidatorPubkey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(ps.ValidatorPublicKey)
}

// GetSignature1 returns the slashing first bls validator signature.
func (ps *ProposerSlashing) GetSignature1() (*bls.Signature, error) {
	return bls.SignatureFromBytes(ps.Signature1)
}

// GetSignature2 returns the slashing second bls validator signature.
func (ps *ProposerSlashing) GetSignature2() (*bls.Signature, error) {
	return bls.SignatureFromBytes(ps.Signature2)
}

// Marshal encodes the data.
func (ps *ProposerSlashing) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(ps)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxProposerSlashingSize {
		return nil, ErrorProposerSlashingSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (ps *ProposerSlashing) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxProposerSlashingSize {
		return ErrorProposerSlashingSize
	}
	return ssz.Unmarshal(d, ps)
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() chainhash.Hash {
	b, _ := ps.Marshal()
	return chainhash.HashH(b)
}
