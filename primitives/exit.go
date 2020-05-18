package primitives

import (
	"bytes"
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey bls.PublicKey
	WithdrawPubkey  bls.PublicKey
	Signature       bls.Signature
}

// Encode encodes the exit to the given writer.
func (e *Exit) Encode(w io.Writer) error {
	sigBytes := e.Signature.Serialize()
	pubBytes := e.ValidatorPubkey.Serialize()
	withdrawPub := e.WithdrawPubkey.Serialize()

	return serializer.WriteElements(w, sigBytes, pubBytes, withdrawPub)
}

// Decode decodes the exit from the given reader.
func (e *Exit) Decode(r io.Reader) error {
	var sigBytes [96]byte
	var pubBytes [48]byte
	var withdrawPubBytes [48]byte

	if err := serializer.ReadElements(r, &sigBytes, &pubBytes, &withdrawPubBytes); err != nil {
		return err
	}

	sig, err := bls.DeserializeSignature(sigBytes)
	if err != nil {
		return err
	}
	pub, err := bls.DeserializePublicKey(pubBytes)
	if err != nil {
		return err
	}
	withdrawPub, err := bls.DeserializePublicKey(withdrawPubBytes)
	if err != nil {
		return err
	}

	e.ValidatorPubkey = *pub
	e.Signature = *sig
	e.WithdrawPubkey = *withdrawPub

	return nil
}

// Hash calculates the hash of the exit.
func (e *Exit) Hash() chainhash.Hash {
	buf := bytes.NewBuffer([]byte{})
	_ = e.Encode(buf)
	return chainhash.HashH(buf.Bytes())
}
