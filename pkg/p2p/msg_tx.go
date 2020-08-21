package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgTx is the struct of the message the is transmitted upon the network.
type MsgTx struct {
	Data *primitives.Tx
}

// Marshal serializes the data to bytes
func (m *MsgTx) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgTx) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgTx) Command() string {
	return MsgTxCmd
}

// MaxPayloadLength returns the maximum size of the MsgTx message.
func (m *MsgTx) MaxPayloadLength() uint64 {
	return primitives.MaxTransactionSize
}
