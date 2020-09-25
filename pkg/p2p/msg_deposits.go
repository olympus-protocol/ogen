package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MaxDeposits define the maximum amount a deposit slice message can contain
var MaxDeposits uint64 = 1024

// MsgDeposits is the struct of the message the is transmitted upon the network.
type MsgDeposits struct {
	Data []*primitives.Deposit `ssz-max:"1024"`
}

// Marshal serializes the data to bytes
func (m *MsgDeposits) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgDeposits) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgDeposits) Command() string {
	return MsgDepositsCmd
}

// MaxPayloadLength returns the maximum size of the MsgDeposits message.
func (m *MsgDeposits) MaxPayloadLength() uint64 {
	return primitives.MaxDepositSize*MaxDeposits + 4
}
