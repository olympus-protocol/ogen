package primitives

import (
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// Execution is a block in the blockchain.
type Execution struct {

	// FromPubKey is the public account of the user that wants to execute the input
	FromPubKey [48]byte

	// To is the account of the contract on which the input is executed.
	To [20]byte

	// Input is the compiled bytecode of the code to be executed.

	Input []byte `ssz-max:"32768"` // Maximum bytecode size is 32Kb

	// Signature is the signature of the user that executes the bytecode
	Signature [96]byte

	Gas      uint64
	GasLimit uint64
}

// Marshal encodes the block.
func (e *Execution) Marshal() ([]byte, error) {
	return e.MarshalSSZ()
}

func (e *Execution) Unmarshal(b []byte) error {
	return e.UnmarshalSSZ(b)
}

// SignatureMessage gets the message the needs to be signed.
func (e Execution) SignatureMessage() chainhash.Hash {
	cp := e
	cp.Signature = [96]byte{}
	b, _ := cp.Marshal()
	return chainhash.HashH(b)
}

// Hash gets the message the needs to be signed.
func (e Execution) Hash() chainhash.Hash {
	b, _ := e.Marshal()
	return chainhash.HashH(b)
}

// SignatureMessage gets the message the needs to be signed.
func (e Execution) VerifySig() error {
	sig, err := bls.SignatureFromBytes(e.Signature[:])
	if err != nil {
		return err
	}

	pub, err := bls.PublicKeyFromBytes(e.FromPubKey[:])
	if err != nil {
		return err
	}

	msg := e.SignatureMessage()

	valid := sig.Verify(pub, msg[:])
	if !valid {
		return errors.New("invalid signature from execution call")
	}

	return nil
}
