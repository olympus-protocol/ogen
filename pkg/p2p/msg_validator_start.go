package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgValidatorStart is the struct of the message the is transmitted upon the network.
type MsgValidatorStart struct {
	Data *primitives.ValidatorHelloMessage
}

// Marshal serializes the data to bytes
func (m *MsgValidatorStart) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgValidatorStart) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgValidatorStart) Command() string {
	return MsgValidatorStartCmd
}

// MaxPayloadLength returns the maximum size of the MsgValidatorStart message.
func (m *MsgValidatorStart) MaxPayloadLength() uint64 {
	return primitives.MaxValidatorHelloMessageSize
}
