package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
)

const MaxBlocksPerMsg = 2000

type MsgBlocks struct {
	Blocks []*primitives.Block `ssz-max:"2000"`
}
