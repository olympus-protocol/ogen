package bdb

import (
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/prysmaticlabs/go-ssz"
)

// BlockNodeDisk is a block node stored on disk.
type BlockNodeDisk struct {
	StateRoot chainhash.Hash
	Height    uint64
	Slot      uint64
	Children  []chainhash.Hash
	Hash      chainhash.Hash
	Parent    chainhash.Hash
}

// Marshal encodes de data
func (bnd *BlockNodeDisk) Marshal() ([]byte, error) {
	return ssz.Marshal(bnd)
}

// Unmarshal decodes the data
func (bnd *BlockNodeDisk) Unmarshal(b []byte) error {
	return ssz.Unmarshal(b, bnd)
}
