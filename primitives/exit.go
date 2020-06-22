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

// Hash calculates the hash of the exit.
func (e *Exit) Hash() (chainhash.Hash, error) {
	exBytes, err := e.MarshalSSZ()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(exBytes), nil
}
