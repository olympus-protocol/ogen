package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgVote is the struct of the message the is transmitted upon the network.
type MsgVote struct {
	Data *primitives.MultiValidatorVote
}

// Marshal serializes the data to bytes
func (m *MsgVote) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgVote) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgVote) Command() string {
	return MsgVoteCmd
}

// MaxPayloadLength returns the maximum size of the MsgVote message.
func (m *MsgVote) MaxPayloadLength() uint64 {
	return 6474
}
