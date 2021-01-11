package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgGovernance is the struct of the message the is transmitted upon the network.
type MsgGovernance struct {
	Data *primitives.GovernanceVote
}

// Marshal serializes the data to bytes
func (m *MsgGovernance) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgGovernance) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgGovernance) Command() string {
	return MsgGovernanceCmd
}

// MaxPayloadLength returns the maximum size of the MsgGovernance message.
func (m *MsgGovernance) MaxPayloadLength() uint64 {
	return primitives.MaxGovernanceVoteSize + 4
}

// PayloadLength returns the size of the MsgGovernance message.
func (m *MsgGovernance) PayloadLength() uint64 {
	return uint64(m.SizeSSZ())
}
