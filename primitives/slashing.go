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
	MaxProposerSlashingSize = MaxBlockHeaderBytes*2 + 96*2 + 48
	// MaxVoteSlashingSize is the maximum amount of bytes a vote slashing can contain.
	MaxVoteSlashingSize = MaxMultiValidatorVoteSize * 2
)

// VoteSlashing is a slashing where validators vote in the span of their other votes.
type VoteSlashing struct {
	Vote1 *MultiValidatorVote
	Vote2 *MultiValidatorVote
}

// Marshal encodes the data.
func (v *VoteSlashing) Marshal() ([]byte, error) {
	b, err := v.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxVoteSlashingSize {
		return nil, ErrorVoteSlashingSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (v *VoteSlashing) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxVoteSlashingSize {
		return ErrorVoteSlashingSize
	}
	return v.UnmarshalSSZ(d)
}

// Hash calculates the hash of the slashing.
func (v *VoteSlashing) Hash() chainhash.Hash {
	b, _ := v.Marshal()
	return chainhash.HashH(b)
}

// RANDAOSlashing is a slashing where a validator reveals their RANDAO signature too early.
type RANDAOSlashing struct {
	RandaoReveal    [96]byte
	Slot            uint64
	ValidatorPubkey [48]byte
}

// GetValidatorPubkey returns the validator bls public key.
func (r *RANDAOSlashing) GetValidatorPubkey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(r.ValidatorPubkey)
}

// GetRandaoReveal returns the bls signature of the randao reveal.
func (r *RANDAOSlashing) GetRandaoReveal() (*bls.Signature, error) {
	return bls.SignatureFromBytes(r.RandaoReveal)
}

// Marshal encodes the data.
func (r *RANDAOSlashing) Marshal() ([]byte, error) {
	b, err := r.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxRandaoSlashingSize {
		return nil, ErrorRandaoSlashingSize
	}
	return b, nil
}

// Unmarshal decodes the data.
func (r *RANDAOSlashing) Unmarshal(b []byte) error {
	if len(b) > MaxRandaoSlashingSize {
		return ErrorRandaoSlashingSize
	}
	return r.UnmarshalSSZ(b)
}

// Hash calculates the hash of the RANDAO slashing.
func (r *RANDAOSlashing) Hash() chainhash.Hash {
	b, _ := r.Marshal()
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
