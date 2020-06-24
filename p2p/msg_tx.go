package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	Txs []primitives.TransferSinglePayload
}

func (m *MsgTx) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

func (m *MsgTx) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
}

func (m *MsgTx) Command() string {
	return MsgTxCmd
}

func (m *MsgTx) MaxPayloadLength() uint32 {
	return maxTxSize
}
