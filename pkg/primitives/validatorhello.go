package primitives

import (
	"encoding/binary"
	"github.com/olympus-protocol/ogen/pkg/bitfield"
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// ValidatorHelloMessage is a message sent by validators to indicate that they are coming online.
type ValidatorHelloMessage struct {
	Timestamp  uint64
	Nonce      uint64
	Signature  [96]byte
	Validators bitfield.Bitlist `ssz:"bitlist" ssz-max:"1024"`
}

// SignatureMessage gets the signed portion of the message.
func (v *ValidatorHelloMessage) SignatureMessage() chainhash.Hash {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, v.Timestamp)

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, v.Nonce)

	var msg []byte
	msg = append(msg, timeBytes...)
	msg = append(msg, nonceBytes...)

	return chainhash.HashH(msg)
}

// Marshal serializes the hello message to the given writer.
func (v *ValidatorHelloMessage) Marshal() ([]byte, error) {
	return v.MarshalSSZ()
}

// Unmarshal deserializes the validator hello message from the reader.
func (v *ValidatorHelloMessage) Unmarshal(b []byte) error {
	return v.UnmarshalSSZ(b)
}
