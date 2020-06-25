package primitives

import (
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey []byte
	WithdrawPubkey  []byte
	Signature       []byte
}

func (e *Exit) GetWithdrawPubKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(e.WithdrawPubkey)
}

func (e *Exit) GetValidatorPubKey() (*bls.PublicKey, error) {
	return bls.PublicKeyFromBytes(e.ValidatorPubkey)
}

func (e *Exit) GetSignature() (*bls.Signature, error) {
	return bls.SignatureFromBytes(e.Signature)
}

// Marshal encodes the data.
func (e *Exit) Marshal() ([]byte, error) {
	return ssz.Marshal(e)
}

// Unmarshal decodes the data.
func (e *Exit) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, e)
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() chainhash.Hash {
	b, _ := e.Marshal()
	return chainhash.DoubleHashH(b)
}
