package p2p

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

type MsgGetBlocks struct {
	HashStop      chainhash.Hash
	LocatorHashes []chainhash.Hash
}

func (m *MsgGetBlocks) Marshal() ([]byte, error) {
	return ssz.Marshal(m)
}

func (m *MsgGetBlocks) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, m)
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
