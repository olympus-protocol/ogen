package p2p

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// MsgGetBlocks is the message that contains the locator to fetch blocks.
type MsgGetBlocks struct {
	HashStop      [32]byte
	LocatorHashes [][32]byte
}

// HashStopH returns the HashStop data as a hash struct
func (m *MsgGetBlocks) HashStopH() *chainhash.Hash {
	h, _ := chainhash.NewHash(m.HashStop)
	return h
}

// Marshal serializes the data to bytes
func (m *MsgGetBlocks) Marshal() ([]byte, error) {
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
func (m *MsgGetBlocks) Unmarshal(b []byte) error {
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
func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

// MaxPayloadLength returns the maximum size of the MsgGetBlocks message.
func (m *MsgGetBlocks) MaxPayloadLength() uint32 {
	return chainhash.HashSize + (40*chainhash.HashSize + 9)
}
