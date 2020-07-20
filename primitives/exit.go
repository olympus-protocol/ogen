package primitives

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// MaxExitLength is the maximum amount of bytes an exit can contain.
const MaxExitLength = 192

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey []byte
	WithdrawPubkey  []byte
	Signature       []byte
}

// GetWithdrawPubKey returns the withdraw bls public key
func (e *Exit) GetWithdrawPubKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(e.WithdrawPubkey)
}

// GetValidatorPubKey returns the validator bls public key
func (e *Exit) GetValidatorPubKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(e.ValidatorPubkey)
}

// GetSignature returns the exit bls signature.
func (e *Exit) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(e.Signature)
}

// Marshal encodes the data.
func (e *Exit) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(e)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (e *Exit) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, e)
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() chainhash.Hash {
	hash, _ := ssz.HashTreeRoot(e)
	return chainhash.Hash(hash)
}
