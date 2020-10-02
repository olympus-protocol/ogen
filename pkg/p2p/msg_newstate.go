package p2p

import "github.com/olympus-protocol/ogen/pkg/primitives"

// MsgFinalization is the struct that contains the node information when announcing a finalization.
type MsgNewState struct {
	Data *primitives.SerializableState
}

// Marshal serializes the data to bytes
func (m *MsgNewState) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgNewState) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgNewState) Command() string {
	return MsgNewStateCmd
}

// MaxPayloadLength returns the maximum size of the MsgVersion message.
func (m *MsgNewState) MaxPayloadLength() uint64 {
	return primitives.MaxStateSize
}
