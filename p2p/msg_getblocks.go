package p2p

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
)

type MsgGetBlocks struct {
	HashStop      chainhash.Hash
	LocatorHashes []chainhash.Hash
}

// Marshal serializes the struct to bytes
func (m *MsgGetBlocks) Marshal() ([]byte, error) {
	return m.Marshal()
}

// Unmarshal deserializes the struct from bytes
func (m *MsgGetBlocks) Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *MsgGetBlocks) Command() string {
	return MsgGetBlocksCmd
}

func (m *MsgGetBlocks) MaxPayloadLength() uint32 {
	return chainhash.HashSize + 40*chainhash.HashSize + 9
}

func NewMsgGetBlock(hashStop chainhash.Hash, locatorHashes []chainhash.Hash) *MsgGetBlocks {
	m := &MsgGetBlocks{
		HashStop:      hashStop,
		LocatorHashes: locatorHashes,
	}
	return m
}
