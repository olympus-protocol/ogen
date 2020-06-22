package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
)

const MaxBlocksPerMsg = 2000

type MsgBlocks struct {
	Blocks []primitives.Block
}

// Marshal serializes the struct to bytes
func (m *MsgBlocks) Marshal() ([]byte, error) {
	return m.Marshal()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgBlocks) Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *MsgBlocks) Command() string {
	return MsgBlocksCmd
}

func NewMsgBlocks(b []primitives.Block) *MsgBlocks {
	m := &MsgBlocks{
		Blocks: b,
	}
	return m
}
