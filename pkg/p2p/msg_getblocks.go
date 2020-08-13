package p2p

import (
	"github.com/olympus-protocol/ogen/pkg/chainhash"
)

// MsgGetBlocks is the message that contains the locator to fetch blocks.
type MsgGetBlocks struct {
	HashStop      [32]byte   `ssz-size:"32"`
	LocatorHashes [][32]byte `ssz-max:"64"`
}

// HashStopH returns the HashStop data as a hash struct
func (m *MsgGetBlocks) HashStopH() *chainhash.Hash {
	h, _ := chainhash.NewHash(m.HashStop)
	return h
}

// Marshal serializes the data to bytes
func (m *MsgGetBlocks) Marshal() ([]byte, error) {
	b, err := m.MarshalSSZ()
	if err != nil {
		return nil, err
	}
	if uint64(len(b)) > m.MaxPayloadLength() {
		return nil, ErrorSizeExceed
	}
	return b, nil
}

// Unmarshal deserializes the data
func (m *MsgGetBlocks) Unmarshal(b []byte) error {
	if uint64(len(b)) > m.MaxPayloadLength() {
		return ErrorSizeExceed
	}
	return m.UnmarshalSSZ(b)
}

// Command returns the message topic
func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

// MaxPayloadLength returns the maximum size of the MsgGetBlocks message.
func (m *MsgGetBlocks) MaxPayloadLength() uint64 {
	return 32 + (64 * 32) + 4 // 32 HashStop + 64 locators * 32 hash size + 4 for the amount of elements on slice
}
