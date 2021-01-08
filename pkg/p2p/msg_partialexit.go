package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgPartialExit is the struct of the message the is transmitted upon the network.
type MsgPartialExits struct {
	Data []*primitives.PartialExit `ssz-max:"1024"`
}

// Marshal serializes the data to bytes
func (m *MsgPartialExits) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgPartialExits) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgPartialExits) Command() string {
	return MsgPartialExitsCmd
}

// MaxPayloadLength returns the maximum size of the MsgExits message.
func (m *MsgPartialExits) MaxPayloadLength() uint64 {
	return primitives.PartialExitsSize
}
