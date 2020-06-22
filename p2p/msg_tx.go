package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	Txs []primitives.TransferSinglePayload
}

// Marshal serializes the struct to bytes
func (m *MsgTx) Marshal() ([]byte, error) {
	return m.Marshal()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgTx) Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *MsgTx) Hash() (chainhash.Hash, error) {
	ser, err := m.Marshal()
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(ser), nil
}

func (m *MsgTx) Command() string {
	return MsgTxCmd
}

func (m *MsgTx) MaxPayloadLength() uint32 {
	return maxTxSize
}
