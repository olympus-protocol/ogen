package msg

import (
	"github.com/olympus-protocol/ogen/specs/state"
)

const (
	maxTxSize = 1024 * 300 // 300 KB
)

type MsgTx struct {
	Txs []*state.Tx `ssz-max:"1000"`
}
