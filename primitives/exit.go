package primitives

import (
	"errors"

	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// ErrorExitSize returned when the exit size is above MaxExitSize
var ErrorExitSize = errors.New("error size is to big")

// MaxExitSize is the maximum amount of bytes an exit can contain.
const MaxExitSize = 192

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
	if len(b) > MaxExitSize {
		return nil, ErrorExitSize
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal decodes the data.
func (e *Exit) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if len(d) > MaxExitSize {
		return ErrorExitSize
	}
	return ssz.Unmarshal(d, e)
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() chainhash.Hash {
	b, _ := e.Marshal()
	return chainhash.HashH(b)
}
