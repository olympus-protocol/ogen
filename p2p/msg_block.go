package p2p

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/prysmaticlabs/go-ssz"
)

// MaxBlocksPerMsg defines the maximum amount of blocks that a peer can send.
const MaxBlocksPerMsg = 32

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
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserializes the data
func (m *MsgBlocks) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
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
