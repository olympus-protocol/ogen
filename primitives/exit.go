package primitives

import (
	"io"

	"github.com/olympus-protocol/ogen/bls"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

// Exit exits the validator from the queue.
type Exit struct {
	ValidatorPubkey bls.PublicKey
	Signature       bls.Signature
}

// Encode encodes the exit to the given writer.
func (e *Exit) Encode(w io.Writer) error {
	sigBytes := e.Signature.Serialize()
	pubBytes := e.ValidatorPubkey.Serialize()

	return serializer.WriteElements(w, sigBytes, pubBytes)
}

// Decode decodes the exit from the given reader.
func (e *Exit) Decode(r io.Reader) error {
	var sigBytes [96]byte
	var pubBytes [48]byte

	if err := serializer.ReadElements(r, &sigBytes, &pubBytes); err != nil {
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

	e.ValidatorPubkey = *pub
	e.Signature = *sig

	return nil
}
