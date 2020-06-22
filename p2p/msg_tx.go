package p2p

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	Txs []primitives.TransferSinglePayload

	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (m *MsgTx) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgTx) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
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
