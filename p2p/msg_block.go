package p2p

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

const MaxBlockSize = 1024 * 1024 * 5 // 5 MB
const MaxBlocksPerMsg = 500

type MsgBlocks struct {
	Blocks []primitives.Block
}

// Marshal serializes the data to bytes
func (m *MsgBlocks) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

// Unmarshal deserializes the data
func (m *MsgBlocks) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, m)
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
