package p2p

import (
	"errors"
	"fmt"
	"io"

	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/serializer"
)

const MaxBlockSize = 1024 * 512 // 512 KB
const MaxBlocksPerMsg = 500

type MsgBlocks struct {
	Blocks []primitives.Block
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

func (m *MsgBlocks) Encode(w io.Writer) error {
	count := len(m.Blocks)
	if count > 500 {
		str := fmt.Sprintf("too many blocks for message "+
			"[count %v, max %v]", count, 500)
		return errors.New(str)
	}
	err := serializer.WriteVarInt(w, uint64(count))
	if err != nil {
		return err
	}
	for _, b := range m.Blocks {
		err := b.Encode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgBlocks) Decode(r io.Reader) error {
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	m.Blocks = make([]primitives.Block, count)
	for i := uint64(0); i < count; i++ {
		err := m.Blocks[i].Decode(r)
		if err != nil {
			return err
		}
	}
	return nil
}
