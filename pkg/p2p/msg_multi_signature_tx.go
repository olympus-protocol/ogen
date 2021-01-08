package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/primitives"
)

// MsgMultiSignatureTx is the struct of the message the is transmitted upon the network.
type MsgMultiSignatureTx struct {
	Data *primitives.MultiSignatureTx
}

// Marshal serializes the data to bytes
func (m *MsgMultiSignatureTx) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the data
func (m *MsgMultiSignatureTx) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgMultiSignatureTx) Command() string {
	return MsgMultiSignatureTxCmd
}

// MaxPayloadLength returns the maximum size of the MsgTxMulti message.
func (m *MsgMultiSignatureTx) MaxPayloadLength() uint64 {
	return 4673
}
