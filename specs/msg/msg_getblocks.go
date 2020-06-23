package msg

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type MsgGetBlocks struct {
	HashStop      []byte   `ssz-size:"32"`
	LocatorHashes [][]byte `ssz-size:"?,32" ssz-max:"16777216"`
}
