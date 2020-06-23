package msg

import (
	"github.com/olympus-protocol/ogen/specs/state"
)

const MaxBlocksPerMsg = 2000

type MsgBlocks struct {
	Blocks []*state.Block `ssz-max:"2000"`
}

func (m *MsgBlocks) Command() string {
	return MsgBlocksCmd
}
