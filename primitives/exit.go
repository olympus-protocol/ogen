package primitives

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey bls.PublicKey
	WithdrawPubkey  bls.PublicKey
	Signature       bls.Signature
	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (e *Exit) Marshal() ([]byte, error) {
	return e.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (e *Exit) Unmarshal(b []byte) error {
	return e.UnmarshalSSZ(b)
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() (chainhash.Hash, error) {
	exBytes, err := e.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(exBytes), nil
}
