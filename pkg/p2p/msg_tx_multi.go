package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgTxMulti is the struct of the message the is transmitted upon the network.
type MsgTxMulti struct {
	Data *primitives.TxMulti
}

// Marshal serializes the data to bytes
func (m *MsgTxMulti) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgTxMulti) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgTxMulti) Command() string {
	return MsgTxMultiCmd
}

// MaxPayloadLength returns the maximum size of the MsgTxMulti message.
func (m *MsgTxMulti) MaxPayloadLength() uint64 {
	return primitives.MaxTransactionMultiSize
}
