package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgExit is the struct of the message the is transmitted upon the network.
type MsgExit struct {
	Data *primitives.Exit
}

// Marshal serializes the data to bytes
func (m *MsgExit) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgExit) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgExit) Command() string {
	return MsgExitCmd
}

// MaxPayloadLength returns the maximum size of the MsgExit message.
func (m *MsgExit) MaxPayloadLength() uint64 {
	return primitives.MaxExitSize
}
