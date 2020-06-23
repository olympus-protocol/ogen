package primitives

import "github.com/olympus-protocol/ogen/utils/chainhash"

type Exit struct {
	ValidatorPubkey []byte `ssz-size:"48"`
	WithdrawPubkey  []byte `ssz-size:"48"`
	Signature       []byte `ssz-size:"96"`
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() chainhash.Hash {
	// TODO handle error
	b, _ := e.MarshalSSZ()
	return chainhash.HashH(b)
}
