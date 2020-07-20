package p2p

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// MsgGetBlocks is the message that contains the locator to fetch blocks.
type MsgGetBlocks struct {
	HashStop      chainhash.Hash
	LocatorHashes []chainhash.Hash
}

// Marshal serializes the data to bytes
func (m *MsgGetBlocks) Marshal() ([]byte, error) {
	b, err := ssz.Marshal(m)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, b), nil
}

// Unmarshal deserializes the data
func (m *MsgGetBlocks) Unmarshal(b []byte) error {
	d, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}
	return ssz.Unmarshal(d, m)
}

// Command returns the message topic
func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

// MaxPayloadLength returns the maximum size of the MsgGetBlocks message.
func (m *MsgGetBlocks) MaxPayloadLength() uint32 {
	return chainhash.HashSize + (40*chainhash.HashSize + 9)
}
