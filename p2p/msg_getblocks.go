package p2p

import (
	"github.com/golang/snappy"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

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
