package p2p

import (
	"github.com/olympus-protocol/ogen/primitives"
)

const maxBlockSize = 1024 * 512 // 512 KB

type MsgBlock struct {
	primitives.Block
}

func (m *MsgBlock) Command() string {
	return MsgBlockCmd
}

func (m *MsgBlock) MaxPayloadLength() uint32 {
	return maxBlockSize
}

func NewMsgBlock(b primitives.Block) *MsgBlock {
	m := &MsgBlock{b}
	return m
}
