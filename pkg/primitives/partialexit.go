package primitives

import (
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/bls/common"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// PartialExit claims a partial amount of a validator balance without removing it from the validator registry.
type PartialExit struct {
	ValidatorPubkey [48]byte
	WithdrawPubkey  [48]byte
	Signature       [96]byte
	Amount          uint64
}

// GetWithdrawPubKey returns the withdraw bls public key
func (p *PartialExit) GetWithdrawPubKey() (common.PublicKey, error) {
	return bls.PublicKeyFromBytes(p.WithdrawPubkey[:])
}

// GetValidatorPubKey returns the validator bls public key
func (p *PartialExit) GetValidatorPubKey() (common.PublicKey, error) {
	return bls.PublicKeyFromBytes(p.ValidatorPubkey[:])
}

// GetSignature returns the exit bls signature.
func (p *PartialExit) GetSignature() (common.Signature, error) {
	return bls.SignatureFromBytes(p.Signature[:])
}

// Marshal encodes the data.
func (p *PartialExit) Marshal() ([]byte, error) {
	return p.MarshalSSZ()
}

// Unmarshal decodes the data.
func (p *PartialExit) Unmarshal(b []byte) error {
	return p.UnmarshalSSZ(b)
}

// Hash calculates the hash of the exit.
func (p *PartialExit) Hash() chainhash.Hash {
	b, _ := p.Marshal()
	return chainhash.HashH(b)
}
