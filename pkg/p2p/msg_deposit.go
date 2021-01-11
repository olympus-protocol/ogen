package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgDeposit is the struct of the message the is transmitted upon the network.
type MsgDeposit struct {
	Data *primitives.Deposit
}

// Marshal serializes the data to bytes
func (m *MsgDeposit) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgDeposit) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgDeposit) Command() string {
	return MsgDepositCmd
}

// MaxPayloadLength returns the maximum size of the MsgDeposit message.
func (m *MsgDeposit) MaxPayloadLength() uint64 {
	return primitives.DepositSize
}

// PayloadLength returns the size of the MsgDeposit message.
func (m *MsgDeposit) PayloadLength() uint64 {
	return uint64(m.SizeSSZ())
}
