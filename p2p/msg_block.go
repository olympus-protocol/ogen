package p2p

import (
	ssz "github.com/ferranbt/fastssz"
	"github.com/olympus-protocol/ogen/primitives"
)

const MaxBlocksPerMsg = 2000

type MsgBlocks struct {
	Blocks []primitives.Block

	ssz.Marshaler
	ssz.Unmarshaler
}

// Marshal serializes the struct to bytes
func (m *MsgBlocks) Marshal() ([]byte, error) {
	return m.MarshalSSZ()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgBlocks) Unmarshal(b []byte) error {
	return m.UnmarshalSSZ(b)
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
