package p2p

import (
	"bytes"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	primitives.Tx
}

func (m *MsgTx) TxHash() (chainhash.Hash, error) {
	buf := bytes.NewBuffer([]byte{})
	err := m.Encode(buf)
	if err != nil {
		return chainhash.Hash{}, err
	}
	return chainhash.DoubleHashH(buf.Bytes()), nil
}

func (m *MsgTx) Command() string {
	return MsgTxCmd
}

func (m *MsgTx) MaxPayloadLength() uint32 {
	return maxTxSize
}

func (m *MsgTx) AddPayload(payload []byte) {
	m.Payload = payload
}
