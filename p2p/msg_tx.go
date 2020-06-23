package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	Txs []*primitives.Tx `ssz-max:"1000"`
}
