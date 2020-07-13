package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

const MaxBlockSize = 1024 * 512 // 512 KB
const MaxBlocksPerMsg = 100

type MsgBlocks struct {
	Blocks []primitives.Block
}

func (m *MsgBlocks) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

func (m *MsgBlocks) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
}

func (m *MsgBlocks) Command() string {
	return MsgBlocksCmd
}

func (m *MsgBlocks) MaxPayloadLength() uint32 {
	return MaxBlockSize * MaxBlocksPerMsg
}

func NewMsgBlocks(b []primitives.Block) *MsgBlocks {
	m := &MsgBlocks{b}
	return m
}
