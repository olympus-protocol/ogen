package primitives

import (
	"errors"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// Execution is a block in the blockchain.
type Execution struct {
	FromPubKey [48]byte
	To         [20]byte
	Input      []byte `ssz-max:"32768"`
	Signature  [96]byte
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
