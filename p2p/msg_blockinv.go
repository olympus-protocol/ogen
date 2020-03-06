package p2p

import (
	"errors"
	"fmt"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
)

const MaxBlocksPerInv = 5000

type MsgBlockInv struct {
	blocks []*MsgBlock
}

func (m *MsgBlockInv) Encode(w io.Writer) error {
	count := len(m.blocks)
	if count > MaxBlocksPerInv {
		str := fmt.Sprintf("too many blocks for message "+
			"[count %v, max %v]", count, MaxBlocksPerInv)
		return errors.New(str)
	}
	err := serializer.WriteVarInt(w, uint64(count))
	if err != nil {
		return err
	}
	for _, b := range m.blocks {
		err := b.Encode(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgBlockInv) Decode(r io.Reader) error {
	count, err := serializer.ReadVarInt(r)
	if err != nil {
		return err
	}
	blockInv := make([]MsgBlock, count)
	m.blocks = make([]*MsgBlock, 0, count)
	for i := uint64(0); i < count; i++ {
		block := &blockInv[i]
		err := block.Decode(r)
		if err != nil {
			return err
		}
		err = m.AddBlock(block)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MsgBlockInv) MaxPayloadLength() uint32 {
	return MaxBlocksPerInv * maxBlockSize
}

func (m *MsgBlockInv) Command() string {
	return MsgBlocksInvCmd
}

func (m *MsgBlockInv) AddBlock(msgBlock *MsgBlock) error {
	if len(m.blocks)+1 > MaxBlocksPerInv {
		str := fmt.Sprintf("too many blocks in message [max %v]",
			MaxBlocksPerInv)
		return errors.New(str)
	}
	m.blocks = append(m.blocks, msgBlock)
	return nil
}

func (m *MsgBlockInv) GetBlocks() []*MsgBlock {
	return m.blocks
}

func (m *MsgBlockInv) GetTxs() []primitives.Tx {
	var txs []primitives.Tx
	for _, block := range m.blocks {
		txs = append(txs, block.Txs...)
	}
	return txs
}

func NewBlockInv(blocks []*MsgBlock) *MsgBlockInv {
	return &MsgBlockInv{
		blocks: blocks,
	}
}
