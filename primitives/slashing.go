package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
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
	MaxRandaoSlashingSize = 152
	// MaxProposerSlashingSize is the maximum amount of bytes a proposer slashing can contain.
	MaxProposerSlashingSize = MaxBlockHeaderBytes*2 + (240)
	// MaxVoteSlashingSize is the maximum amount of bytes a vote slashing can contain.
	MaxVoteSlashingSize = MaxMultiValidatorVoteSize * 2
)

// VoteSlashing is a slashing where validators vote in the span of their other votes.
type VoteSlashing struct {
	Vote1 *MultiValidatorVote
	Vote2 *MultiValidatorVote
}

// Marshal encodes the data.
func (vs *VoteSlashing) Marshal() ([]byte, error) {
	b, err := vs.MarshalSSZ()
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
	return vs.UnmarshalSSZ(d)
}

// Hash calculates the hash of the slashing.
func (vs *VoteSlashing) Hash() chainhash.Hash {
	b, _ := vs.Marshal()
	return chainhash.HashH(b)
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO signature too early.
type RANDAOSlashing struct {
	RandaoReveal    [96]byte `ssz-size:"96"`
	Slot            uint64
	ValidatorPubkey [48]byte `ssz-size:"48"`
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
	b, err := rs.MarshalSSZ()
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
	return rs.UnmarshalSSZ(d)
}

// Hash calculates the hash of the RANDAO slashing.
func (rs *RANDAOSlashing) Hash() chainhash.Hash {
	b, _ := rs.Marshal()
	return chainhash.HashH(b)
}

// ProposerSlashing is a slashing to a block proposer that proposed two blocks at the same slot.
type ProposerSlashing struct {
	BlockHeader1       *BlockHeader
	BlockHeader2       *BlockHeader
	Signature1         [96]byte `ssz-size:"96"`
	Signature2         [96]byte `ssz-size:"96"`
	ValidatorPublicKey [48]byte `ssz-size:"48"`
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
	b, err := ps.MarshalSSZ()
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
	return ps.UnmarshalSSZ(d)
}

// Hash calculates the hash of the proposer slashing.
func (ps *ProposerSlashing) Hash() chainhash.Hash {
	b, _ := ps.Marshal()
	return chainhash.HashH(b)
}
