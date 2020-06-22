package primitives

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey []byte
	WithdrawPubkey  []byte
	Signature       []byte
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() (chainhash.Hash, error) {
	exBytes, err := e.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.HashH(exBytes), nil
}
