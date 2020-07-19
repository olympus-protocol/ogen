package p2p

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

const MaxBlockSize = 1024 * 1024 * 5 // 5 MB
const MaxBlocksPerMsg = 5

// MsgBlocks is the struct of the message the is transmited upon the network.
type MsgBlocks struct {
	Blocks []primitives.Block
}

// Marshal serializes the data to bytes
func (m *MsgBlocks) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(m)
	if err != nil {
		return nil, err
	}
	if uint32(len(b)) > m.MaxPayloadLength() {
		return nil, ErrorSizeExceed
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserializes the data
func (m *MsgBlocks) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	if uint32(len(d)) > m.MaxPayloadLength() {
		return ErrorSizeExceed
	}
	return ssz.Unmarshal(d, m)
}

// Command returns the message topic
func (m *MsgBlocks) Command() string {
	return MsgBlocksCmd
}

// MaxPayloadLength returns the maximum size of the MsgBlocks message.
func (m *MsgBlocks) MaxPayloadLength() uint32 {
	return primitives.MaxBlockSize * MaxBlocksPerMsg
}
