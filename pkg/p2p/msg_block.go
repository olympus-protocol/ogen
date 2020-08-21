package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgBlock is the struct of the message the is transmitted upon the network.
type MsgBlock struct {
	Data *primitives.Block
}

// Marshal serializes the data to bytes
func (m *MsgBlock) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgBlock) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgBlock) Command() string {
	return MsgBlockCmd
}

// MaxPayloadLength returns the maximum size of the MsgBlock message.
func (m *MsgBlock) MaxPayloadLength() uint64 {
	return primitives.MaxBlockSize
}
