package primitives

import (
	bls_interface "github.com/olympus-protocol/ogen/pkg/bls/interface"

	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// MaxExitSize is the maximum amount of bytes an exit can contain.
const MaxExitSize = 192

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey [48]byte
	WithdrawPubkey  [48]byte
	Signature       [96]byte
}

// GetWithdrawPubKey returns the withdraw bls public key
func (e *Exit) GetWithdrawPubKey() (bls_interface.PublicKey, error) {
	return bls.CurrImplementation.PublicKeyFromBytes(e.WithdrawPubkey[:])
}

// GetValidatorPubKey returns the validator bls public key
func (e *Exit) GetValidatorPubKey() (bls_interface.PublicKey, error) {
	return bls.CurrImplementation.PublicKeyFromBytes(e.ValidatorPubkey[:])
}

// GetSignature returns the exit bls signature.
func (e *Exit) GetSignature() (bls_interface.Signature, error) {
	return bls.CurrImplementation.SignatureFromBytes(e.Signature[:])
}

// Marshal encodes the data.
func (e *Exit) Marshal() ([]byte, error) {
	return e.MarshalSSZ()
}

// Unmarshal decodes the data.
func (e *Exit) Unmarshal(b []byte) error {
	return e.UnmarshalSSZ(b)
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() chainhash.Hash {
	b, _ := e.Marshal()
	return chainhash.HashH(b)
}
