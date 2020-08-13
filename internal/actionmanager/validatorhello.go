package actionmanager

import "encoding/binary"

// ValidatorHelloMessage is a message sent by validators to indicate that they are coming online.
type ValidatorHelloMessage struct {
	PublicKey [48]byte
	Timestamp uint64
	Nonce     uint64
	Signature [96]byte
}

// SignatureMessage gets the signed portion of the message.
func (v *ValidatorHelloMessage) SignatureMessage() []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, v.Timestamp)

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, v.Nonce)

	var msg []byte
	msg = append(msg, v.PublicKey[:]...)
	msg = append(msg, timeBytes...)
	msg = append(msg, nonceBytes...)

	return msg
}

// Marshal serializes the hello message to the given writer.
func (v *ValidatorHelloMessage) Marshal() ([]byte, error) {
	b, err := v.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if len(b) > MaxValidatorHelloMessageSize {
		return nil, ErrorValidatorHelloMessageSize
	}
	return b, nil
}

// Unmarshal deserializes the validator hello message from the reader.
func (v *ValidatorHelloMessage) Unmarshal(b []byte) error {
	if len(b) > MaxValidatorHelloMessageSize {
		return ErrorValidatorHelloMessageSize
	}
	return v.UnmarshalSSZ(b)
}