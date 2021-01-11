package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MaxExits define the maximum amount a exits slice message can contain
var MaxExits uint64 = 1024

// MsgExits is the struct of the message the is transmitted upon the network.
type MsgExits struct {
	Data []*primitives.Exit `ssz-max:"1024"`
}

// Marshal serializes the data to bytes
func (m *MsgExits) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgExits) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgExits) Command() string {
	return MsgExitsCmd
}

// MaxPayloadLength returns the maximum size of the MsgExits message.
func (m *MsgExits) MaxPayloadLength() uint64 {
	return primitives.MaxExitSize * MaxExits
}

// PayloadLength returns the size of the MsgExits message.
func (m *MsgExits) PayloadLength() uint64 {
	return uint64(m.SizeSSZ())
}
